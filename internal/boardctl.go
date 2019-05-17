package internal

import (
	"flag"
	"fmt"
	"strconv"
	"encoding/json"

	"mosho-boardctl/pkg"
)

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
	
	jsonb, _ := json.Marshal(env)
	
	return string(jsonb), nil
}

func invokeCmd(intervalStr string, pattern string) (string, error) {

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return "", err
	}

	data := &pkg.IrData{
		Interval: interval,
		Pattern: pattern,
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
