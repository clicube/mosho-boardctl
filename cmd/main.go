package main

import (
	"fmt"
	"os"

	"mosho-boardctl/internal"
)

func main() {
	res, err := internal.Exec()
	fmt.Println(res)
	if err != nil {
		os.Exit(255)
	} else {
		os.Exit(0)
	}
}
