package internal

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"

	"mosho-boardctl/pkg"
)

type jsonEnv struct {
	Result      string  `json:"result"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Brightness  float64 `json:"brightness"`
}

func Exec() (string, error) {
	flag.Parse()

	var res string
	var err error
	command := flag.Arg(0)
	switch command {
	case "env":
		res, err = invokeEnv()
	case "cmd":
		res, err = invokeCmd(flag.Arg(1), flag.Arg(2))
	case "":
		err = fmt.Errorf("Command required")
	default:
		err = fmt.Errorf("Unknown command: " + command)
	}
	if err != nil {
		res = errToJson(err)
	}

	return res, err
}

func errToJson(err error) string {
	obj := map[string]string{"result": "ng", "message": err.Error()}
	jsonb, _ := json.Marshal(obj)
	return string(jsonb)
}

func invokeEnv() (string, error) {

	board, err := pkg.NewBoard(nil)
	if err != nil {
		return "", err
	}

	env, err := board.GetEnv()
	if err != nil {
		return "", err
	}

	jenv := &jsonEnv{
		Result: "ok",
		Temperature: env.Temperature,
		Humidity: env.Humidity,
		Brightness: env.Brightness,
	}

	jsonb, _ := json.Marshal(jenv)

	return string(jsonb), nil
}

func invokeCmd(intervalStr string, pattern string) (string, error) {

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return "", err
	}

	data := &pkg.IrData{
		Interval: interval,
		Pattern:  pattern,
	}

	board, err := pkg.NewBoard(nil)
	if err != nil {
		return "", err
	}

	err = board.SendIr(data)
	if err != nil {
		return "", err
	}

	return "{\"result\":\"ok\"}", nil
}
