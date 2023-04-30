package components

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

type ExChans struct {
	bTaken      chan bool
	dec_dis     chan bool
	dec_stall   chan bool
	stall       chan bool
	dec_mRegsOk chan bool
	mRegsOk     chan bool
	wbMRegs     chan bool
}

type ExCache struct {
	stallCycles chan int
	stallData   chan []byte
}

func readFromMemory(loc int, mem Memory) (int, int) {
	lMem := <-mem
	mem <- lMem

	lines := strings.Split(string(lMem), "\n")

	if loc >= len(lines) {
		return 0, 10
	}
	if lines[loc] == "" {
		return 0, 10
	}

	val, err := strconv.Atoi(lines[loc])
	if err != nil {
		return 0, 10
	}

	return val, 10
}

// Execute given instruction
func Execute(flg Flags, mem Memory, buf Buffer, bus ExChans,
	cache ExCache, modRegCa ModRegCache) {

	decData := <-buf.in
	stallCycles := <-cache.stallCycles
	stallData := <-cache.stallData

	opc := uint(decData[0])
	opr := op.MatchOpc(opc)

	var opds []int
	for i := 1; i < 13; i += 4 {
		uopds := binary.BigEndian.Uint32(decData[i : i+4])
		opds = append(opds, int(uopds))
	}

	stall := false
	if stallCycles > 0 {
		stall = true
		opr = op.Nop
	} else {
		stall = opr == op.Mul || opr == op.Div || opr == op.Ld
	}
	bus.dec_stall <- stall

	if opr.Class != "ctf" {
		bus.bTaken <- true
		bus.dec_dis <- false
	}

	if opr != op.Hlt {
		bus.wbMRegs <- false
	}

	var wmem byte
	var wb byte

	var result int
	var desReg int
	var memLoc int

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop:
			bus.bTaken <- true
			bus.dec_dis <- false

		case op.Hlt:
			bus.bTaken <- true
			bus.dec_dis <- false
			bus.wbMRegs <- true
			flg.halt <- true

		case op.Jmp:
			bus.bTaken <- true
			bus.dec_dis <- false

		case op.Beq:
			branch := opds[0] == opds[1]
			bus.bTaken <- branch
			bus.dec_dis <- !(branch)

		case op.Bz:
			branch := opds[0] == 0
			bus.bTaken <- branch
			bus.dec_dis <- !branch
		}

	case "ari":
		wb = 1
		desReg = opds[0]

		switch opr {
		case op.Add:
			result = opds[1] + opds[2]

		case op.Sub:
			result = opds[1] - opds[2]

		//3 cycles
		case op.Mul:
			stallCycles = 3

			result = opds[1] * opds[2]

		//16 cycles
		case op.Div:
			stallCycles = 16

			if opds[2] == 0 {
				result = 0
			} else {
				result = opds[1] / opds[2]
			}
		}

	case "log":
		wb = 1
		desReg = opds[0]

		switch opr {
		case op.And:
			result = opds[1] & opds[2]

		case op.Or:
			result = opds[1] | opds[2]

		case op.Xor:
			result = opds[1] ^ opds[2]

		case op.Not:
			result = ^opds[1]

		case op.Cmp:
			if opds[1] > opds[2] {
				result = 1
			} else if opds[1] == opds[2] {
				result = 0
			} else {
				result = -1
			}
		}

	case "dat":
		switch opr {
		case op.Ld:
			wb = 1
			desReg = opds[0]
			memLoc = opds[1] + opds[2]

			result, stallCycles = readFromMemory(memLoc, mem)

		case op.Wrt:
			wmem = 1
			memLoc = opds[2] + opds[1]

			result = opds[0]

		case op.Mv:
			wb = 1
			desReg = opds[0]

			result = opds[1]
		}
	}

	<-bus.mRegsOk
	modRegs := <-modRegCa

	if wb == 1 {
		mod := false
		for i := 0; i < len(modRegs); i++ {
			if modRegs[i].reg == desReg {
				modRegs[i].val = result
				mod = true
				break
			}
		}
		if !mod {
			modRegs = append(modRegs, struct {
				reg int
				val int
			}{desReg, result})
		}
	}

	modRegCa <- modRegs
	bus.dec_mRegsOk <- true

	var memData []byte

	memData = append(memData, wmem)
	memData = binary.BigEndian.AppendUint32(memData, uint32(memLoc))

	memData = append(memData, wb)
	memData = binary.BigEndian.AppendUint32(memData, uint32(desReg))

	memData = binary.BigEndian.AppendUint32(memData, uint32(result))

	if opr == op.Mul || opr == op.Div {
		stallData = memData
		memData = make([]byte, 14)
	}

	if stall {
		stallCycles--
		if stallCycles == 0 {
			memData = stallData
		}
	}

	cache.stallCycles <- stallCycles
	cache.stallData <- stallData
	buf.out <- memData

	flg.exChk <- true
}
