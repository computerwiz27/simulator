package main

import (
	"flag"
	"os"

	"github.com/computerwiz27/simulator/components"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	regNo := flag.Int("reg", 8, "Number of simulated registers")

	memFile := flag.String("mem", "mem.txt", "Text file location containing simulated memory")

	progFile := flag.String("prog", "prog.txt", "Text file location containing program ")

	// memSize := flag.Int("memSize", 256, "Size of simulated memory in bytes")

	memOut := flag.String("memOut", "mem.txt", "Location for output memory file")

	// assemble := flag.Bool("asb", false, "Assemble from assembly to machine code")

	flag.Parse()

	mem, err := os.ReadFile(*memFile)
	check(err)

	prog, err := os.ReadFile(*progFile)

	components.Run(mem, *memOut, *regNo, prog)
}
