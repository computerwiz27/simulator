package main

import (
	"os"

	"github.com/computerwiz27/simulator/components"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args

	f, err := os.ReadFile(args[1])
	check(err)

	components.Run(f)
}
