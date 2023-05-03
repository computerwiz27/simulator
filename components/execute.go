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

func mvEle2Front(array []CaAddr, pos int) []CaAddr {
	ele := array[pos]
	array = append(array[:pos], array[pos+1:]...)
	array = append([]CaAddr{ele}, array...)

	return array
}

func replaceEle(array []CaAddr, ele CaAddr) []CaAddr {
	array = append([]CaAddr{ele}, array[:len(array)-1]...)
	return array
}

func readFromMemory(loc int, mem Memory, sysCa SysCache) (int, int) {
	l1Cache := <-sysCa.l1

	for i := 0; i < len(l1Cache); i++ {
		if l1Cache[i].loc == loc {
			l1Cache = mvEle2Front(l1Cache, i)

			sysCa.l1 <- l1Cache

			return l1Cache[0].val, 1 //3
		}
	}

	l2Cache := <-sysCa.l2

	for i := 0; i < len(l2Cache); i++ {
		if l2Cache[i].loc == loc {
			l2Cache = mvEle2Front(l2Cache, i)
			l1Cache = replaceEle(l1Cache, l2Cache[i])

			sysCa.l1 <- l1Cache
			sysCa.l2 <- l2Cache

			return l2Cache[0].val, 1 // 10
		}
	}

	l3Cache := <-sysCa.l3

	for i := 0; i < len(l3Cache); i++ {
		if l3Cache[i].loc == loc {
			l3Cache = mvEle2Front(l3Cache, i)

			l1Cache = replaceEle(l1Cache, l3Cache[i])
			l2Cache = replaceEle(l2Cache, l3Cache[i])

			sysCa.l1 <- l1Cache
			sysCa.l2 <- l2Cache
			sysCa.l3 <- l3Cache

			return l3Cache[0].val, 1 //40
		}

	}

	memBytes := <-mem

	lines := strings.Split(string(memBytes), "\n")

	val, err := strconv.Atoi(lines[loc])

	if loc >= len(lines) || lines[loc] == "" || err != nil {
		l1Cache = replaceEle(l1Cache, CaAddr{loc: loc, val: 0})
		l2Cache = replaceEle(l2Cache, CaAddr{loc: loc, val: 0})
		l3Cache = replaceEle(l3Cache, CaAddr{loc: loc, val: 0})
		val = 0
	}

	l1Cache = replaceEle(l1Cache, CaAddr{loc: loc, val: val})
	l2Cache = replaceEle(l2Cache, CaAddr{loc: loc, val: val})
	l3Cache = replaceEle(l3Cache, CaAddr{loc: loc, val: val})

	sysCa.l1 <- l1Cache
	sysCa.l2 <- l2Cache
	sysCa.l3 <- l3Cache
	mem <- memBytes

	return val, 1 //100
}

// Execute given instruction
func Execute(flg Flags, mem Memory, sysCa SysCache, buf Buffer, bus ExChans,
	cache ExCache, modRegCa Cache) {

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

	<-bus.mRegsOk
	modRegs := <-modRegCa

	if wb == 1 {
		mod := false
		for i := 0; i < len(modRegs); i++ {
			if modRegs[i].loc == desReg {
				modRegs[i].val = result
				mod = true
				break
			}
		}
		if !mod {
			modRegs = append(modRegs, struct {
				loc int
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
