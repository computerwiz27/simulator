package components

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

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
		memChk:  wmem,
		wbChk:   wback,
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

func bufInit() (Buffer, Buffer, Buffer, Buffer, Buffer) {
	fet_dec := make(chan []byte, 1)
	fet_dec <- make([]byte, 4)

	dec_ex := make(chan []byte, 1)
	dec_ex <- make([]byte, 14)

	ex_mem := make(chan []byte, 1)
	ex_mem <- make([]byte, 14)

	mem_wb := make(chan []byte, 1)
	mem_wb <- make([]byte, 9)

	fetBuf := Buffer{
		out: fet_dec,
	}

	decBuf := Buffer{
		in:  fet_dec,
		out: dec_ex,
	}

	exBuf := Buffer{
		in:  dec_ex,
		out: ex_mem,
	}

	memBuf := Buffer{
		in:  ex_mem,
		out: mem_wb,
	}

	wbBuf := Buffer{
		in: mem_wb,
	}

	return fetBuf, decBuf, exBuf, memBuf, wbBuf
}

func busInit() (FetChans, DecChans, ExChans, MemChans, WbChans) {
	//fetch channels
	dec_nIns := make(chan int)
	bran := make(chan int)
	bTaken := make(chan bool)

	//decode channels
	dec_dis := make(chan bool)
	dec_stall := make(chan bool)
	dec_mRegsOk := make(chan bool)

	//execute channels
	ex_stall := make(chan bool)
	ex_mRegsOk := make(chan bool)
	wbMRegs := make(chan bool)

	fetCh := FetChans{
		nIns:   dec_nIns,
		bran:   bran,
		bTaken: bTaken,
	}

	decCh := DecChans{
		nIns:   dec_nIns,
		bran:   bran,
		dis:    dec_dis,
		stall:  dec_stall,
		mRegOk: dec_mRegsOk,
	}

	exCh := ExChans{
		bTaken:      bTaken,
		dec_dis:     dec_dis,
		dec_stall:   dec_stall,
		stall:       ex_stall,
		dec_mRegsOk: dec_mRegsOk,
		mRegsOk:     ex_mRegsOk,
		wbMRegs:     wbMRegs,
	}

	memCh := MemChans{
		ex_stall: ex_stall,
	}

	wbCh := WbChans{
		ex_mRegsOk: ex_mRegsOk,
		wbMRegs:    wbMRegs,
	}

	return fetCh, decCh, exCh, memCh, wbCh
}

func cacheInit() (ModRegCache, FetCache, DecCache, ExCache) {
	modRegCa := make(ModRegCache, 1)
	modRegCa <- make([]struct {
		reg int
		val int
	}, 0)

	forks := make(chan []uint, 1)
	forks <- make([]uint, 0)
	bLast := make(chan bool, 1)
	bLast <- false
	fetCa := FetCache{
		forks: forks,
		bLast: bLast,
	}

	dec_lastCycleStall := make(chan bool, 1)
	dec_lastCycleStall <- false
	dec_stallData := make(chan []byte, 1)
	dec_stallData <- make([]byte, 14)
	decCa := DecCache{
		lcystall:  dec_lastCycleStall,
		stallData: dec_stallData,
	}

	ex_stallCycles := make(chan int, 1)
	ex_stallCycles <- 0
	ex_stallData := make(chan []byte, 1)
	ex_stallData <- make([]byte, 14)
	exCa := ExCache{
		stallCycles: ex_stallCycles,
		stallData:   ex_stallData,
	}

	return modRegCa, fetCa, decCa, exCa
}

func printRegisters(regs Registers) {

	for i := 0; i < 32; i++ {
		select {
		case val, ok := <-regs.reg[i]:
			if ok {
				fmt.Printf("reg%d: %d\n", i, val)
			}
			regs.reg[i] <- val

		default:
			fmt.Printf("reg%d: no value\n", i)
		}
	}
	fmt.Println()
}

func step(cycle uint, regs Registers, decBuf Buffer, memBuf Buffer,
	fetCa FetCache, decCa DecCache, exCa ExCache, mRegCa ModRegCache) bool {

	fmt.Printf("Cycle: %d\n", cycle)

loop:
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		command, _ := reader.ReadString('\n')
		command = command[:len(command)-1]
		fmt.Println()

		switch command {
		case "s":
			break loop

		case "r":
			printRegisters(regs)

		case "b":
			fet_dec := <-decBuf.in
			dec_ex := <-decBuf.out
			ex_mem := <-memBuf.in
			mem_wb := <-memBuf.out

			fet_decVal := binary.BigEndian.Uint32(fet_dec)
			binString := strconv.FormatUint(uint64(fet_decVal), 2)
			for len(binString) < 32 {
				binString = "0" + binString
			}
			fmt.Printf("fet_def: %s\n", binString)

			dec_exOpc := int(dec_ex[0])
			var dec_exOpds []int
			for i := 1; i < 13; i += 4 {
				uopds := binary.BigEndian.Uint32(dec_ex[i : i+4])
				dec_exOpds = append(dec_exOpds, int(uopds))
			}
			fmt.Printf("dec_ex: %d %d %d %d\n",
				dec_exOpc, dec_exOpds[0], dec_exOpds[1], dec_exOpds[2])

			ex_memStore := int(ex_mem[0])
			ex_memLoc := binary.BigEndian.Uint32(ex_mem[1:5])
			ex_memWback := int(ex_mem[5])
			ex_memDreg := binary.BigEndian.Uint32(ex_mem[6:10])
			ex_memVal := int(binary.BigEndian.Uint32(ex_mem[10:14]))
			fmt.Printf("ex_mem: %d %d %d %d %d\n",
				ex_memStore, ex_memLoc, ex_memWback, ex_memDreg, ex_memVal)

			mem_wbWback := int(mem_wb[0])
			mem_wbDreg := binary.BigEndian.Uint32(mem_wb[1:5])
			mem_wbVal := binary.BigEndian.Uint32(mem_wb[5:9])
			fmt.Printf("mem_wb: %d %d %d\n", mem_wbWback, mem_wbDreg, mem_wbVal)

			fmt.Println()

			decBuf.in <- fet_dec
			decBuf.out <- dec_ex
			memBuf.in <- ex_mem
			memBuf.out <- mem_wb

		case "c":
			fetForks := <-fetCa.forks
			decLCStall := <-decCa.lcystall
			decStallData := <-decCa.stallData
			exStallCy := <-exCa.stallCycles
			exStallData := <-exCa.stallData
			mRegCaVal := <-mRegCa

			fmt.Println("Fetch:")
			fmt.Println("Fetch Forks:")
			for i := 0; i < len(fetForks); i++ {
				fmt.Println(fetForks[i])
			}
			fmt.Println()

			fmt.Println("Decode:")
			fmt.Printf("Last cycle was stalled: %t\n", decLCStall)
			decStallOpc := int(decStallData[0])
			var decStallOpds []int
			for i := 1; i < 13; i += 4 {
				uopds := binary.BigEndian.Uint32(decStallData[i : i+4])
				decStallOpds = append(decStallOpds, int(uopds))
			}
			fmt.Printf("Stall Data: %d %d %d %d\n",
				decStallOpc, decStallOpds[0], decStallOpds[1], decStallOpds[2])
			fmt.Println()

			fmt.Println("Execute:")
			fmt.Printf("Stall cycles: %d\n", exStallCy)
			exStallStore := int(exStallData[0])
			exStallLoc := binary.BigEndian.Uint32(exStallData[1:5])
			exStallWback := int(exStallData[5])
			exStallDreg := binary.BigEndian.Uint32(exStallData[6:10])
			ex_memVal := int(binary.BigEndian.Uint32(exStallData[10:14]))
			fmt.Printf("Stall Data: %d %d %d %d %d\n",
				exStallStore, exStallLoc, exStallWback, exStallDreg, ex_memVal)
			fmt.Println()

			fmt.Println("Modified Registers Cache:")
			for i := 0; i < len(mRegCaVal); i++ {
				fmt.Printf("reg: %d, val: %d\n", mRegCaVal[i].reg, mRegCaVal[i].val)
			}

			fmt.Println()

			fetCa.forks <- fetForks
			decCa.lcystall <- decLCStall
			decCa.stallData <- decStallData
			exCa.stallCycles <- exStallCy
			exCa.stallData <- exStallData
			mRegCa <- mRegCaVal

		case "h":
			fmt.Println("Enter")
			fmt.Println("-'s' to step to next cycle")
			fmt.Println("-'r' to print register values")
			fmt.Println("-'b' to print buffer value")
			fmt.Println("-'c' to print component cache values")
			fmt.Println("-'h' to print these commands")
			fmt.Println("-'e' to exit step mode")

		case "e":
			return false

		default:
			fmt.Printf("Unrecognised command '%s'. Enter 'h' for help\n", command)
		}
	}

	return true
}

// Finishing processes after execution is done
func finish(cycles uint, registers Registers, memory Memory, memOut string) {

	fmt.Println("Done!")

	fmt.Printf("Operation took %d cycles \n", cycles)

	//Print the values in the registers
	printRegisters(registers)

	//Print memory to a file where each line is a memory address
	bytes := <-memory

	lines := string(bytes)
	_ = lines

	//Remove all bytes that are emty or contain "\n"
	for i := len(bytes) - 1; i >= 0; i-- {
		if !(bytes[i] == 0x00 || bytes[i] == 0x0a) {
			break
		}
		bytes = bytes[:len(bytes)-1]
	}

	os.WriteFile(memOut, bytes, 0644)
}

func cycleCheck(flg Flags) {
	<-flg.fetChck
	<-flg.decChk
	<-flg.exChk
	<-flg.memChk
	<-flg.wbChk
}

// Runs the simulator with the given specifications
func Run(memFile []byte, memOut string, progFile []byte, stepMode bool) {

	registers, flags, memory, program := memInit(memFile, progFile)
	fetBuf, decBuf, exBuf, memBuf, wbBuf := bufInit()
	fetChans, decChans, exChans, memChans, wbChans := busInit()
	modRegCache, fetCache, decCache, exCache := cacheInit()

	if stepMode {
		fmt.Println("Step mode. Enter")
		fmt.Println("-'s' to step to next cycle")
		fmt.Println("-'r' to print register values")
		fmt.Println("-'b' to print buffer value")
		fmt.Println("-'c' to print component cache values")
		fmt.Println("-'h' to print these commands")
		fmt.Println("-'e' to exit step mode")
		fmt.Println()
	}

	cycles := uint(0)
cycle:
	for {
		if stepMode {
			stepMode = step(cycles, registers, decBuf, memBuf,
				fetCache, decCache, exCache, modRegCache)
		}

		go Fetch(registers, flags, program, fetBuf, fetChans, fetCache)

		go Decode(registers, flags, memory, decBuf, decChans, decCache, modRegCache)

		go Execute(flags, memory, exBuf, exChans, exCache, modRegCache)

		go WriteToMemory(flags, memory, memBuf, memChans)

		go WriteBack(registers, flags, wbBuf, wbChans, modRegCache)

		cycleCheck(flags)
		cycles++

		select {
		case <-flags.halt:
			break cycle
		default:
		}
	}

	finish(cycles, registers, memory, memOut)
}
