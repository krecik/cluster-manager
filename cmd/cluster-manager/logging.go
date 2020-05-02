package main

import (
	"fmt"
	"os"
)

func print(v ...interface{}) {
	fmt.Println("#", fmt.Sprintln(v...))
}

func fatal(v ...interface{}) {
	print(v)
	os.Exit(1)
}
