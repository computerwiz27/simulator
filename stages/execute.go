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
	WbMRegs     chan bool
}

type ExCache struct {
	StallCycles chan int
	StallmRegs  chan []c.CaAddr
	StallData   chan []byte
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

// Execute given instruction
func Execute(flg c.Flags, mem c.Memory, sysCa c.SysCache, buf c.Buffer, bus ExChans,
	cache ExCache, modRegCa c.Cache) {

	decData := <-buf.In

	stallCycles := <-cache.StallCycles
	stallData := <-cache.StallData
	stallmRegs := <-cache.StallmRegs

	opc := uint(decData[0])
	opr := op.MatchOpc(opc)

	var opds []int
	for i := 1; i < 13; i += 4 {
		uopd := binary.BigEndian.Uint32(decData[i : i+4])
		opd := int32(uopd)
		opds = append(opds, int(opd))
	}

	stall := false
	if stallCycles > 0 {
		stall = true
		opr = op.Nop
	} else {
		stall = opr == op.Mul || opr == op.Div || opr == op.Ld
	}
	bus.Dec_stall <- stall

	if opr.Class != "ctf" {
		bus.BTaken <- true
		bus.Dec_dis <- false
	}

	if opr != op.Ld {
		<-bus.MemOk
	}

	if opr != op.Hlt {
		bus.WbMRegs <- false
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
			bus.BTaken <- true
			bus.Dec_dis <- false

		case op.Hlt:
			bus.BTaken <- true
			bus.Dec_dis <- false
			bus.WbMRegs <- true
			flg.Halt <- true

		case op.Jmp:
			bus.BTaken <- true
			bus.Dec_dis <- false

		case op.Beq:
			branch := opds[0] == opds[1]
			bus.BTaken <- branch
			bus.Dec_dis <- !(branch)

		case op.Bz:
			branch := opds[0] == 0
			bus.BTaken <- branch
			bus.Dec_dis <- !branch
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
			stallCycles = 3

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
			<-bus.MemOk

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

	<-bus.MRegsOk
	modRegs := <-modRegCa

	if stall && wb == 1 {
		stallmRegs = updateModRegs(wb, desReg, result, modRegs)
	} else if stallCycles == 1 {
		modRegs = stallmRegs
	} else {
		modRegs = updateModRegs(wb, desReg, result, modRegs)
	}

	modRegCa <- modRegs
	bus.Dec_mRegsOk <- true

	var memData []byte

	memData = append(memData, wmem)
	memData = binary.BigEndian.AppendUint32(memData, uint32(memLoc))

	memData = append(memData, wb)
	memData = binary.BigEndian.AppendUint32(memData, uint32(desReg))

	memData = binary.BigEndian.AppendUint32(memData, uint32(result))

	if opr == op.Mul || opr == op.Div || opr == op.Ld {
		stallData = memData
		memData = make([]byte, 14)
	}

	if stall {
		stallCycles--
		if stallCycles == 0 {
			memData = stallData
		}
	}

	cache.StallCycles <- stallCycles
	cache.StallData <- stallData
	cache.StallmRegs <- stallmRegs

	buf.Out <- memData

	flg.ExChk <- true
}
