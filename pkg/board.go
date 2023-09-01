package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/tarm/serial"
)

const (
	lockFilePath = "/var/tmp/boardctl.lock"
	prompt       = "RasPi-ExtBoard> "
)

var reTmp = regexp.MustCompile("TMP: ([0-9]+)")
var reHum = regexp.MustCompile("HUM: ([0-9]+)")
var reBri = regexp.MustCompile("BRI: ([0-9]+)")

type IoPort interface {
	Read(buf []byte) (int, error)
	Write(data []byte) (int, error)
	Close() error
}

type IoPortCreator interface {
	Create() (IoPort, error)
}

type DefaultPortCreator struct{}

func (*DefaultPortCreator) Create() (IoPort, error) {
	return serial.OpenPort(&serial.Config{
		Name:        "/dev/ttyAMA0",
		Baud:        9600,
		ReadTimeout: time.Millisecond * 200,
	})
}

type Board struct {
	portCreator IoPortCreator
}

func NewBoard(portCreator IoPortCreator) (*Board, error) {
	return &Board{portCreator}, nil
}

func (b *Board) GetEnv() (*Env, error) {

	tempHumRes, err := execCommand(b.portCreator, "temp_read")
	if err != nil {
		return nil, fmt.Errorf("Failed to execute temp_read: %s", err)
	}

	temp, err := parseResult(tempHumRes, reTmp, "Tempreture")
	if err != nil {
		return nil, fmt.Errorf("Failed to parse temp: %s", err)
	}
	temp /= 10

	hum, err := parseResult(tempHumRes, reHum, "Humidity")
	if err != nil {
		return nil, fmt.Errorf("Failed to parse hum: %s", err)
	}
	hum /= 10

	briRes, err := execCommand(b.portCreator, "bri_read")
	if err != nil {
		return nil, fmt.Errorf("Failed to execute bri_read: %s", err)
	}

	bri, err := parseResult(briRes, reBri, "Brightness")
	if err != nil {
		return nil, fmt.Errorf("Failed to parse bri: %s", err)
	}

	return &Env{
		Temperature: temp,
		Humidity:    hum,
		Brightness:  bri,
	}, nil
}

func (b *Board) SendIr(data *IrData) error {

	_, err := execCommand(b.portCreator, fmt.Sprintf("ir_send %d\n%s", data.Interval, data.Pattern))
	if err != nil {
		return fmt.Errorf("Failed to execute ir_send: %s", err)
	}

	return nil
}

func lock() (*flock.Flock, error) {
	fileLock := flock.New(lockFilePath)
	err := fileLock.Lock()
	if err != nil {
		return nil, fmt.Errorf("Failed to lock the lock file %s: %s", lockFilePath, err)
	}
	return fileLock, nil
}

func unlock(fileLock *flock.Flock) error {
	err := fileLock.Unlock()
	if err != nil {
		return err
	}
	err = os.Remove(lockFilePath)
	return err
}

func execCommand(portCreatorArg IoPortCreator, cmd string) (string, error) {

	// Take a lock
	lock, err := lock()
	if err != nil {
		return "", fmt.Errorf("Failed to lock: %s", err)
	}
	defer unlock(lock)

	// Create a port
	var portCreator IoPortCreator
	if portCreatorArg == nil {
		portCreator = &DefaultPortCreator{}
	} else {
		portCreator = portCreatorArg
	}
	port, err := portCreator.Create()
	if err != nil {
		return "", fmt.Errorf("Failed to open port: %s", err)
	}
	defer port.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	ch := make(chan string)
	errch := make(chan error)

	go func() {
		_, err = port.Write([]byte(cmd + "\n"))
		if err != nil {
			errch <- fmt.Errorf("Failed to write to port: %s", err)
			return
		}

		buf := bytes.Buffer{}

		for {
			rbuf := make([]byte, 128)
			n, err := port.Read(rbuf)
			if err != nil {
				// No data received
			} else {
				buf.Write(rbuf[0:n])

				// Check if command is complete
				bufstr := buf.String()
				if strings.HasSuffix(bufstr, prompt) {
					ch <- bufstr
					break
				}
			}
		}
	}()
	select {
	case res := <-ch:
		lines := strings.Split(res, "\n")
		filteredLines := lines[1 : len(lines)-1]
		return strings.Join(filteredLines, "\n"), nil
	case err := <-errch:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("Command timeout: %s", cmd)
	}
}

func parseResult(data string, re *regexp.Regexp, name string) (float64, error) {
	matched := re.FindStringSubmatch(data)
	if matched == nil {
		return 0, fmt.Errorf("Failed to get %s", name)
	}
	val, err := strconv.ParseFloat(matched[1], 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
