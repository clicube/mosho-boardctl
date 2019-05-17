package pkg

import (
	"bytes"
	"testing"
	"time"
)

type MockPortCreator struct {
	creator func() (IoPort, error)
}

func (c *MockPortCreator) Create() (IoPort, error) {
	return c.creator()
}

type MockPort struct {
	onRead  func([]byte) (int, error)
	onWrite func([]byte) (int, error)
	onClose func() error
}

func (p MockPort) Read(buf []byte) (int, error) {
	return p.onRead(buf)
}
func (p MockPort) Write(data []byte) (int, error) {
	return p.onWrite(data)
}
func (p MockPort) Close() error {
	return p.onClose()
}

func TestGetEnvsOk(t *testing.T) {

	// Arrange
	internalBuf := bytes.Buffer{}
	readBuf := bytes.Buffer{}
	miniBuf := make([]byte, 4)

	port := MockPort{
		onRead: func(buf []byte) (int, error) {
			n, err := internalBuf.Read(miniBuf)
			if err != nil {
				time.Sleep(200 * time.Millisecond)
			} else {
				readBuf.Write(miniBuf[0:n])
			}
			return readBuf.Read(buf)
		},
		onWrite: func(data []byte) (int, error) {
			internalBuf.Write(data)
			switch string(data) {
			case "temp_read\n":
				go func() {
					time.Sleep(time.Millisecond * 500)
					internalBuf.WriteString("TMP: 320\n")
					internalBuf.WriteString("HUM: 550\n")
					internalBuf.WriteString("RasPi-ExtBoard> ")
				}()
			case "bri_read\n":
				go func() {
					time.Sleep(time.Millisecond * 200)
					internalBuf.WriteString("BRI: 20\n")
					internalBuf.WriteString("RasPi-ExtBoard> ")
				}()
			default:
				internalBuf.WriteString("ERROR\n")
				internalBuf.WriteString("RasPi-ExtBoard> ")
			}
			return len(data), nil
		},
		onClose: func() error {
			return nil
		},
	}

	portCreator := &MockPortCreator{
		creator: func() (IoPort, error) {
			return port, nil
		},
	}

	sut, err := NewBoard(portCreator)

	// Act
	res, err := sut.GetEnv()

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if res.Temperature != 32 {
		t.Fatalf("Envs.Temperature is not 32.0 but %f", res.Temperature)
	}
	if res.Humidity != 55 {
		t.Fatalf("Envs.Humidity is not 55.0 but %f", res.Humidity)
	}
	if res.Brightness != 20 {
		t.Fatalf("Envs.Brightness is not 32.0 but %f", res.Brightness)
	}
}

func TestGetEnvsTimeout(t *testing.T) {

	port := MockPort{
		onRead: func(buf []byte) (int, error) {
			time.Sleep(10 * time.Second)
			return 0, nil
		},
		onWrite: func(data []byte) (int, error) {
			return len(data), nil
		},
		onClose: func() error {
			return nil
		},
	}

	portCreator := &MockPortCreator{
		creator: func() (IoPort, error) {
			return port, nil
		},
	}

	sut := &Board{portCreator}
	_, err := sut.GetEnv()
	if err == nil {
		t.Fatal("Timeout is not occured")
	}
}
