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

	f, err := os.ReadFile(args[1])
	check(err)

	lines := strings.Split(string(f), "\n")

	components.Run(lines, regNo)
}
