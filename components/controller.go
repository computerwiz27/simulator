package components

import (
	"fmt"
	"os"
)

func initialise(memSize int, memFile []byte, regNo int) (Registers, Flags, Memory) {
	pc := make(chan uint32, 1)

	var regs []chan int32
	for i := 0; i < regNo; i++ {
		regs = append(regs, make(chan int32, 1))
		regs[i] <- 0
	}

	registers := Registers{
		pc:  pc,
		reg: regs,
	}

	registers.pc <- 0

	halt := make(chan bool)

	flags := Flags{
		halt: halt,
	}

	mem := make(Memory, 1)

	var tmpMem []byte
	tmpMem = append(tmpMem, memFile...)
	tmpMem = append(tmpMem, []byte("\n")...)

	for i := len(tmpMem); i < memSize; i++ {
		tmpMem = append(tmpMem, 0)
	}

	mem <- tmpMem

	return registers, flags, mem
}

func finish(registers Registers, memory Memory, memOut string, regNo int) {
	for i := 0; i < regNo; i++ {
		select {
		case val, ok := <-registers.reg[i]:
			if ok {
				fmt.Printf("reg%d: %d\n", i, val)
			}
		default:
			fmt.Printf("reg%d: no value\n", i)
		}
	}

	bytes := <-memory
	for i := len(bytes) - 1; i >= 0; i-- {
		if bytes[i] != 0x0a {
			break
		}
		bytes = bytes[:len(bytes)-1]
	}

	os.WriteFile(memOut, bytes, 0644)
}

func Run(memFile []byte, memSize int, memOut string, regNo int) {

	registers, flags, memory := initialise(memSize, memFile, regNo)

	go Fetch(registers, flags, memory)

	<-flags.halt

	finish(registers, memory, memOut, regNo)

	fmt.Println("done")
}
