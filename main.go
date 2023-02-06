package main

import (
	"os"
	"strings"

	"github.com/computerwiz27/simulator/components"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	regNo := 4

	args := os.Args

	var f []byte
	var err error

	if len(args) != 1 {
		f, err = os.ReadFile(args[1])
	} else {
		f, err = os.ReadFile("mem.txt")
	}

	check(err)

	lines := strings.Split(string(f), "\n")

	components.Run(lines, regNo)
}
