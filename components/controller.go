package components

import (
	"fmt"
	"os"
)

// Initialises the channels for registers, flags and memory
func initialise(memSize int, memFile []byte, regNo int, prog []int) (Registers, Flags, Memory, Prog) {
	//Set up registers channel with value 0
	//Channel have buffer 1 so they can store a value
	pc := make(chan uint32, 1)
	pc <- 0

	var regs []chan int32
	for i := 0; i < regNo; i++ {
		regs = append(regs, make(chan int32, 1))
		regs[i] <- 0
	}

	registers := Registers{
		pc:  pc,
		reg: regs,
	}

	//Set up flags channel
	halt := make(chan bool)

	flags := Flags{
		halt: halt,
	}

	//Set up memory channel with buffer 1
	mem := make(Memory, 1)

	var tmpMem []byte
	tmpMem = append(tmpMem, memFile...)
	tmpMem = append(tmpMem, []byte("\n")...)

	mem <- tmpMem

	// Set up program memory channel
	program := make(Prog, 1)
	program <- prog

	return registers, flags, mem, program
}

// Finishing processes after execution is done
func finish(registers Registers, memory Memory, memOut string, regNo int) {
	//Print the values in the registers
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

	//Print memory to a file where each line is a memory address
	bytes := <-memory

	//Remove all bytes that are emty or contain "\n"
	for i := len(bytes) - 1; i >= 0; i-- {
		if !(bytes[i] == 0x00 || bytes[i] == 0x0a) {
			break
		}
		bytes = bytes[:len(bytes)-1]
	}

	os.WriteFile(memOut, bytes, 0644)
}

// Runs the simulator with the given specifications
func Run(memFile []byte, memSize int, memOut string, regNo int, prog []int) {

	registers, flags, memory, program := initialise(memSize, memFile, regNo, prog)

	go Fetch(registers, flags, memory, program)

	//when the halt flag is passed the simulation is done
	<-flags.halt

	finish(registers, memory, memOut, regNo)

	fmt.Println("Done!")
}
