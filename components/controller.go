package components

import (
	"fmt"
	"os"

	"github.com/computerwiz27/simulator/compiler"
)

// Initialises the channels for registers, flags and memory
func memInit(memFile []byte, progFile []byte) (Registers, Flags, Memory, Memory) {
	//Set up registers channel with value 0
	//Channel have buffer 1 so they can store a value
	pc := make(chan uint, 1)
	pc <- 0

	var regs []chan int
	for i := 0; i < 32; i++ {
		regs = append(regs, make(chan int, 1))
		regs[i] <- 0
	}

	registers := Registers{
		pc:  pc,
		reg: regs,
	}

	//Set up flags channel
	halt := make(chan bool, 1)
	fetch := make(chan bool)
	decode := make(chan bool)
	execute := make(chan bool)
	wback := make(chan bool)
	wmem := make(chan bool)

	flags := Flags{
		halt:    halt,
		fetChck: fetch,
		decChk:  decode,
		exChk:   execute,
		wbChk:   wback,
		wmChk:   wmem,
	}

	//Set up memory channel with buffer 1
	mem := make(Memory, 1)

	var tmpMem []byte
	tmpMem = append(tmpMem, memFile...)
	tmpMem = append(tmpMem, []byte("\n")...)

	mem <- tmpMem

	// Set up program memory channel
	program := make(Memory, 1)
	program <- compiler.Assemble(progFile)

	return registers, flags, mem, program
}

func bufInit() (Buffer, Buffer, Buffer, Buffer) {
	fet_dec := make(Buffer, 1)
	dec_ex := make(Buffer, 1)
	ex_wb := make(Buffer, 1)
	ex_wm := make(Buffer, 1)

	return fet_dec, dec_ex, ex_wb, ex_wm
}

func cycleCheck(flg Flags) {
	<-flg.fetChck
	<-flg.decChk
	<-flg.exChk
	<-flg.wbChk
	<-flg.wmChk
}

// Finishing processes after execution is done
func finish(registers Registers, memory Memory, memOut string) {
	//Print the values in the registers
	for i := 0; i < 32; i++ {
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
func Run(memFile []byte, memOut string, progFile []byte) {

	registers, flags, memory, program := memInit(memFile, progFile)
	fet_dec, dec_ex, _, _ := bufInit()

	cycles := 0

cycle:
	for {
		Fetch(registers, flags, memory, program, fet_dec)
		//<-flags.fetChck
		cycles++

		Decode(registers, flags, memory, program, fet_dec, dec_ex)
		//<-flags.decChk
		cycles++

		Execute(registers, flags, memory, program, dec_ex)
		//<-flags.exChk
		cycles++

		// WriteBack(registers, flags, memory, program)
		// <-flags.wbChk

		// WriteToMemory(registers, flags, memory, program)
		// <-flags.wmChk

		//cycles++

		select {
		case <-flags.halt:
			break cycle

		default:
		}

	}

	finish(registers, memory, memOut)

	fmt.Println("Done!")
}
