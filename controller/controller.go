package controller

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	c "github.com/computerwiz27/simulator/components"
	s "github.com/computerwiz27/simulator/stages"
)

// Initialises the channels for registers, flags, system memory and program memory
func memInit(memFile []byte, progFile []byte) (c.Registers, c.Flags, c.Memory, c.Memory) {
	//Set up registers channel with value 0
	//Channel have buffer 1 so they can store a value
	pc := make(chan uint, 1)
	pc <- 0

	var regs []chan int
	for i := 0; i < 32; i++ {
		regs = append(regs, make(chan int, 1))
		regs[i] <- 0
	}

	registers := c.Registers{
		Pc:  pc,
		Reg: regs,
	}

	//Set up flags channel
	halt := make(chan bool, 1)
	fetch := make(chan bool)
	decode := make(chan bool)
	execute := make(chan bool)
	wback := make(chan bool)
	wmem := make(chan bool)

	flags := c.Flags{
		Halt:    halt,
		FetChck: fetch,
		DecChk:  decode,
		ExChk:   execute,
		MemChk:  wmem,
		WbChk:   wback,
	}

	//Set up memory channel with buffer 1
	mem := make(c.Memory, 1)

	var tmpMem []byte
	tmpMem = append(tmpMem, memFile...)
	tmpMem = append(tmpMem, []byte("\n")...)

	mem <- tmpMem

	// Set up program memory channel
	program := make(c.Memory, 1)
	program <- Assemble(progFile)

	return registers, flags, mem, program
}

// Initialise the channels for the buffers that connect the execution stages
func bufInit() (c.Buffer, c.Buffer, c.Buffer, c.Buffer, c.Buffer) {

	//Initialise the buffer channels with all zeros
	fet_dec := make(chan []byte, 1)
	fet_dec <- make([]byte, 4)

	dec_ex := make(chan []byte, 1)
	dec_ex <- make([]byte, 14)

	ex_mem := make(chan []byte, 1)
	ex_mem <- make([]byte, 14)

	mem_wb := make(chan []byte, 1)
	mem_wb <- make([]byte, 9)

	fetBuf := c.Buffer{
		Out: fet_dec,
	}

	decBuf := c.Buffer{
		In:  fet_dec,
		Out: dec_ex,
	}

	exBuf := c.Buffer{
		In:  dec_ex,
		Out: ex_mem,
	}

	memBuf := c.Buffer{
		In:  ex_mem,
		Out: mem_wb,
	}

	wbBuf := c.Buffer{
		In: mem_wb,
	}

	return fetBuf, decBuf, exBuf, memBuf, wbBuf
}

// Initialise the channels that act as a bus between the execution stages and the controller
func busInit() (s.FetChans, s.DecChans, s.ExChans, s.MemChans, s.WbChans) {
	//these channels don't have a buffer since the stages operate in parallel

	//fetch channels
	dec_nIns := make(chan int)
	bran := make(chan int)
	bTaken := make(chan bool)
	fet_stall := make(chan bool)

	//decode channels
	dec_dis := make(chan bool)
	dec_stall := make(chan bool)
	dec_mRegsOk := make(chan bool)

	//execute channels
	ex_memOk := make(chan bool)
	ex_mRegsOk := make(chan bool)
	wbMRegs := make(chan bool)

	fetCh := s.FetChans{
		NIns:   dec_nIns,
		Bran:   bran,
		BTaken: bTaken,
		Stall:  fet_stall,
	}

	decCh := s.DecChans{
		NIns:      dec_nIns,
		Bran:      bran,
		Dis:       dec_dis,
		Stall:     dec_stall,
		Fet_stall: fet_stall,
		MRegOk:    dec_mRegsOk,
	}

	exCh := s.ExChans{
		BTaken:      bTaken,
		Dec_dis:     dec_dis,
		Dec_stall:   dec_stall,
		MemOk:       ex_memOk,
		Dec_mRegsOk: dec_mRegsOk,
		MRegsOk:     ex_mRegsOk,
		WbMRegs:     wbMRegs,
	}

	memCh := s.MemChans{
		Ex_memOk: ex_memOk,
	}

	wbCh := s.WbChans{
		Ex_mRegsOk: ex_mRegsOk,
		WbMRegs:    wbMRegs,
	}

	return fetCh, decCh, exCh, memCh, wbCh
}

// Initialise the system cache and the stage specific caches
func cacheInit() (c.Cache, c.Cache, c.Cache, c.Cache,
	s.FetCache, s.DecCache, s.ExCache) {

	//Level 1 cache has 16 addresable spaces, with a size of 0.5 kibibytes
	l1Cache := make(c.Cache, 1)
	l1Ca := make([]c.CaAddr, 16)
	for i := range l1Ca {
		l1Ca[i].Loc = -1
	}
	l1Cache <- l1Ca

	//Level 2 cache has 128 addresable spaces, with a size of 4 kibibytes
	l2Cache := make(c.Cache, 1)
	l2Ca := make([]c.CaAddr, 128)
	for i := range l2Ca {
		l2Ca[i].Loc = -1
	}
	l2Cache <- l2Ca

	//Level 1 cache has 1024 addresable spaces, with a size of 32 kibibytes
	l3Cache := make(c.Cache, 1)
	l3Ca := make([]c.CaAddr, 1024) //32 kibibytes
	for i := range l3Ca {
		l3Ca[i].Loc = -1
	}
	l3Cache <- l3Ca

	//The modified register cache does not have a predefined size
	modRegCa := make(chan []c.CaAddr, 1)
	modRegCa <- make([]c.CaAddr, 0)

	//Initialise the component cache with the default values

	//Fetch cache
	forks := make(chan []uint, 1)
	forks <- make([]uint, 0)
	bLast := make(chan bool, 1)
	bLast <- false
	fet_lastStall := make(chan bool, 1)
	fet_lastStall <- false
	fetCa := s.FetCache{
		Forks:     forks,
		LCyBranch: bLast,
		LCyStall:  fet_lastStall,
	}

	//Decode cache
	dec_lastCycleStall := make(chan bool, 1)
	dec_lastCycleStall <- false
	dec_stallData := make(chan []byte, 1)
	dec_stallData <- make([]byte, 14)
	dec_lastIns := make(chan uint32, 1)
	dec_lastIns <- uint32(0)
	decCa := s.DecCache{
		Lcystall:  dec_lastCycleStall,
		StallData: dec_stallData,
		LastIns:   dec_lastIns,
	}

	//Execute cache
	ex_stallCycles := make(chan int, 1)
	ex_stallCycles <- 0
	ex_stallData := make(chan []byte, 1)
	ex_stallData <- make([]byte, 14)
	ex_stallmRegs := make(chan []c.CaAddr, 1)
	ex_stallmRegs <- make([]c.CaAddr, 0)
	exCa := s.ExCache{
		StallCycles: ex_stallCycles,
		StallData:   ex_stallData,
		StallmRegs:  ex_stallmRegs,
	}

	return l1Cache, l2Cache, l3Cache, modRegCa, fetCa, decCa, exCa
}

// Print registers to terminal
func printRegisters(regs c.Registers) {

	//for each of the 32 registers
	for i := 0; i < 32; i++ {
		select {
		// if the register holds a value, which it should
		case val, ok := <-regs.Reg[i]:
			if ok {
				fmt.Printf("reg%d: %d\n", i, val) //print it
			}
			//return the value to the register
			regs.Reg[i] <- val

		//this default case is in place for the exeption where the register chan
		//does not hold a value. This shouldn't happen!
		default:
			fmt.Printf("reg%d: no value\n", i)
		}
	}
	fmt.Println()
}

// Handles the step function of the program
func step(cycle uint, regs c.Registers, decBuf c.Buffer, memBuf c.Buffer,
	fetCa s.FetCache, decCa s.DecCache, exCa s.ExCache, mRegCa c.Cache) bool {

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
			fet_dec := <-decBuf.In
			dec_ex := <-decBuf.Out
			ex_mem := <-memBuf.In
			mem_wb := <-memBuf.Out

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

			decBuf.In <- fet_dec
			decBuf.Out <- dec_ex
			memBuf.In <- ex_mem
			memBuf.Out <- mem_wb

		case "c":
			fetForks := <-fetCa.Forks
			decLCStall := <-decCa.Lcystall
			decStallData := <-decCa.StallData
			exStallCy := <-exCa.StallCycles
			exStallData := <-exCa.StallData
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
				fmt.Printf("reg: %d, val: %d\n", mRegCaVal[i].Loc, mRegCaVal[i].Val)
			}

			fmt.Println()

			fetCa.Forks <- fetForks
			decCa.Lcystall <- decLCStall
			decCa.StallData <- decStallData
			exCa.StallCycles <- exStallCy
			exCa.StallData <- exStallData
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
func finish(cycles uint, registers c.Registers, memory c.Memory, memOut string) {

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

// Check if the execution of the stages is finished
func cycleCheck(flg c.Flags) {
	//dump all the done flag channels
	<-flg.FetChck
	<-flg.DecChk
	<-flg.ExChk
	<-flg.MemChk
	<-flg.WbChk
}

// Runs the simulator with the given specifications
func Run(memFile []byte, memOut string, progFile []byte, stepMode bool) {

	//initialise registers
	registers, flags, memory, program := memInit(memFile, progFile)
	//initialise buffers between stages
	fetBuf, decBuf, exBuf, memBuf, wbBuf := bufInit()
	//initialise the stages' channels
	fetChans, decChans, exChans, memChans, wbChans := busInit()
	//initialise caches
	l1Cache, l2Cache, l3Cache, modRegCache, fetCache, decCache, exCache := cacheInit()
	sysCache := c.SysCache{
		L1: l1Cache,
		L2: l2Cache,
		L3: l3Cache,
	}

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

		go s.Fetch(registers, flags, program, fetBuf, fetChans, fetCache)

		go s.Decode(registers, flags, memory, decBuf, decChans, decCache, modRegCache)

		go s.Execute(flags, memory, sysCache, exBuf, exChans, exCache, modRegCache)

		go s.Mem(flags, memory, sysCache, memBuf, memChans)

		go s.WriteBack(registers, flags, wbBuf, wbChans, modRegCache)

		cycleCheck(flags)
		cycles++

		//if the halt flag has been raised, break and exit
		select {
		case <-flags.Halt:
			break cycle
		default:
		}
	}

	finish(cycles, registers, memory, memOut)
}
