package stages

import (
	"encoding/binary"
	"strconv"
	"strings"

	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

type ExChans struct {
	BTaken      chan bool
	Dec_dis     chan bool
	Dec_stall   chan bool
	MemOk       chan bool
	Dec_mRegsOk chan bool
	MRegsOk     chan bool
	Wb_wrtMRegs chan bool
	Wb_mRegsOk  chan bool
}

type ExCache struct {
	StallCycles  chan int
	StallmRegs   chan []c.CaAddr
	StallData    chan []byte
	StallBrTaken chan bool
	StallHalt    chan bool
}

func mvEle2Front(array []c.CaAddr, pos int) []c.CaAddr {
	ele := array[pos]
	array = append(array[:pos], array[pos+1:]...)
	array = append([]c.CaAddr{ele}, array...)

	return array
}

func replaceEle(array []c.CaAddr, ele c.CaAddr) []c.CaAddr {
	array = append([]c.CaAddr{ele}, array[:len(array)-1]...)
	return array
}

func readFromMemory(loc int, mem c.Memory, sysCa c.SysCache) (int, int) {
	l1Cache := <-sysCa.L1

	for i := 0; i < len(l1Cache); i++ {
		if l1Cache[i].Loc == loc {
			l1Cache = mvEle2Front(l1Cache, i)

			sysCa.L1 <- l1Cache

			return l1Cache[0].Val, 3
		}
	}

	l2Cache := <-sysCa.L2

	for i := 0; i < len(l2Cache); i++ {
		if l2Cache[i].Loc == loc {
			l2Cache = mvEle2Front(l2Cache, i)
			l1Cache = replaceEle(l1Cache, l2Cache[i])

			sysCa.L1 <- l1Cache
			sysCa.L2 <- l2Cache

			return l2Cache[0].Val, 10
		}
	}

	l3Cache := <-sysCa.L3

	for i := 0; i < len(l3Cache); i++ {
		if l3Cache[i].Loc == loc {
			l3Cache = mvEle2Front(l3Cache, i)

			l1Cache = replaceEle(l1Cache, l3Cache[i])
			l2Cache = replaceEle(l2Cache, l3Cache[i])

			sysCa.L1 <- l1Cache
			sysCa.L2 <- l2Cache
			sysCa.L3 <- l3Cache

			return l3Cache[0].Val, 40
		}

	}

	memBytes := <-mem

	lines := strings.Split(string(memBytes), "\n")

	val, err := strconv.Atoi(lines[loc])

	if loc >= len(lines) || lines[loc] == "" || err != nil {
		l1Cache = replaceEle(l1Cache, c.CaAddr{Loc: loc, Val: 0})
		l2Cache = replaceEle(l2Cache, c.CaAddr{Loc: loc, Val: 0})
		l3Cache = replaceEle(l3Cache, c.CaAddr{Loc: loc, Val: 0})
		val = 0
	}

	l1Cache = replaceEle(l1Cache, c.CaAddr{Loc: loc, Val: val})
	l2Cache = replaceEle(l2Cache, c.CaAddr{Loc: loc, Val: val})
	l3Cache = replaceEle(l3Cache, c.CaAddr{Loc: loc, Val: val})

	sysCa.L1 <- l1Cache
	sysCa.L2 <- l2Cache
	sysCa.L3 <- l3Cache
	mem <- memBytes

	return val, 100
}

func updateModRegs(wb byte, desReg int, result int, modRegs []c.CaAddr) []c.CaAddr {
	if wb == 1 {
		mod := false
		for i := 0; i < len(modRegs); i++ {
			if modRegs[i].Loc == desReg {
				modRegs[i].Val = result
				mod = true
				break
			}
		}
		if !mod {
			modRegs = append(modRegs, c.CaAddr{Loc: desReg, Val: result})
		}
	}
	return modRegs
}

func executeDat(opr op.Op, opds []int, mem c.Memory, sysCa c.SysCache) (
	int, int, int, byte, byte, int, bool) {

	wb := byte(0)
	wmem := byte(0)
	halt := false
	stallCycles := 0

	var desReg int
	var memLoc int
	var result int

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop:

		case op.Hlt:
			halt = true

		case op.Jmp:

		}

	case "ari":
		wb = 1
		desReg = opds[0]

		switch opr {
		case op.Add:
			result = opds[1] + opds[2]

		case op.Sub:
			result = opds[1] - opds[2]

		case op.Mul:
			stallCycles = 10

			result = opds[1] * opds[2]

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

			result, stallCycles = readFromMemory(memLoc, mem, sysCa)

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

	return result, desReg, memLoc, wb, wmem, stallCycles, halt
}

func executeBr(opr op.Op, opds []int) (int, int, byte, bool, int, bool) {
	wb := byte(0)
	halt := false
	stallCycles := 0
	branchTaken := true

	var desReg int
	var result int

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop:

		case op.Hlt:
			halt = true

		case op.Beq:
			stallCycles = 1

			branchTaken = opds[0] == opds[1]

		case op.Bz:
			stallCycles = 1

			branchTaken = opds[0] == 0

		case op.Jmp:

		}

	case "ari":
		wb = 1
		desReg = opds[0]

		switch opr {
		case op.Add:
			result = opds[1] + opds[2]

		case op.Sub:
			result = opds[1] - opds[2]

		case op.Mul:
			stallCycles = 10

			result = opds[1] * opds[2]

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
		case op.Mv:
			wb = 1
			desReg = opds[0]

			result = opds[1]
		}
	}

	return result, desReg, wb, branchTaken, stallCycles, halt
}

// Execute given instruction
func Execute(flg c.Flags, mem c.Memory, sysCa c.SysCache, buf c.Buffer, bus ExChans,
	cache ExCache, modRegCa c.Cache) {

	decData := <-buf.In

	stallCycles := <-cache.StallCycles
	stallData := <-cache.StallData
	stallmRegs := <-cache.StallmRegs
	stallBrTaken := <-cache.StallBrTaken
	stallHalt := <-cache.StallHalt

	datOpc := uint(decData[0])
	datOpr := op.MatchOpc(datOpc)

	var datOpds []int
	for i := 1; i < 13; i += 4 {
		opd := int32(binary.BigEndian.Uint32(decData[i : i+4]))
		datOpds = append(datOpds, int(opd))
	}

	brOpc := uint(decData[13])
	brOpr := op.MatchOpc(brOpc)

	var brOpds []int
	for i := 14; i < 26; i += 4 {
		opd := int32(binary.BigEndian.Uint32(decData[i : i+4]))
		brOpds = append(brOpds, int(opd))
	}

	datFirst := decData[26]

	lastCycleStall := false
	stall := false
	if stallCycles > 0 {
		lastCycleStall = true
		stall = true
		datOpr = op.Nop
		brOpr = op.Nop
		stallCycles--
	}

	<-bus.MemOk
	datResult, datDReg, memLoc, datWb,
		wmem, datStallCys, datHalt := executeDat(datOpr, datOpds, mem, sysCa)

	brResult, brDReg, brWb, brTaken,
		brStallCys, brHalt := executeBr(brOpr, brOpds)

	if !stall {
		stallCycles = datStallCys
		if brStallCys > stallCycles {
			stallCycles = brStallCys
		}
	}

	if stallCycles > 0 {
		stall = true
	}

	bus.Dec_stall <- stall

	halt := datHalt || brHalt

	if !stall {
		bus.BTaken <- brTaken
		bus.Dec_dis <- !brTaken

		bus.Wb_wrtMRegs <- halt
		flg.Halt <- halt
	} else if stallCycles == 0 {
		bus.BTaken <- stallBrTaken
		bus.Dec_dis <- !stallBrTaken

		bus.Wb_wrtMRegs <- stallHalt
		flg.Halt <- stallHalt
	} else {
		bus.BTaken <- true
		bus.Dec_dis <- false

		bus.Wb_wrtMRegs <- false
		flg.Halt <- false
		if !lastCycleStall {
			stallBrTaken = brTaken
		}
	}

	<-bus.MRegsOk
	modRegs := <-modRegCa

	wb := byte(0)
	if datWb == 1 || brWb == 1 {
		wb = 1
	}

	if stall && wb == 1 {
		if datFirst == 1 {
			stallmRegs = updateModRegs(datWb, datDReg, datResult, modRegs)
			stallmRegs = updateModRegs(brWb, brDReg, brResult, stallmRegs)
		} else {
			stallmRegs = updateModRegs(brWb, brDReg, brResult, modRegs)
			stallmRegs = updateModRegs(datWb, datDReg, datResult, stallmRegs)
		}
	} else if stallCycles == 1 {
		modRegs = stallmRegs
	} else {
		if datFirst == 1 {
			modRegs = updateModRegs(datWb, datDReg, datResult, modRegs)
			modRegs = updateModRegs(brWb, brDReg, brResult, modRegs)
		} else {
			modRegs = updateModRegs(brWb, brDReg, brResult, modRegs)
			modRegs = updateModRegs(datWb, datDReg, datResult, modRegs)
		}
	}

	modRegCa <- modRegs
	bus.Dec_mRegsOk <- true
	if halt {
		bus.Wb_mRegsOk <- true
	}

	var memData []byte

	memData = append(memData, wmem)
	memData = binary.BigEndian.AppendUint32(memData, uint32(memLoc))
	memData = binary.BigEndian.AppendUint32(memData, uint32(datResult))

	memData = append(memData, datWb)
	memData = binary.BigEndian.AppendUint32(memData, uint32(datDReg))
	memData = binary.BigEndian.AppendUint32(memData, uint32(datResult))

	memData = append(memData, brWb)
	memData = binary.BigEndian.AppendUint32(memData, uint32(brDReg))
	memData = binary.BigEndian.AppendUint32(memData, uint32(brResult))

	memData = append(memData, datFirst)

	if datOpr == op.Mul || datOpr == op.Div || datOpr == op.Ld ||
		brOpr == op.Mul || brOpr == op.Div || brOpr == op.Beq || brOpr == op.Bz {
		stallData = memData
		memData = make([]byte, 29)
	}

	if stall {
		if stallCycles == 0 {
			memData = stallData
		}
	}

	cache.StallCycles <- stallCycles
	cache.StallData <- stallData
	cache.StallmRegs <- stallmRegs
	cache.StallBrTaken <- stallBrTaken
	cache.StallHalt <- stallHalt

	buf.Out <- memData

	flg.ExChk <- true
}
