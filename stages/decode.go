package stages

import (
	"encoding/binary"

	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

type DecChans struct {
	NIns      chan int
	Bran      chan int
	Dis       chan bool
	Stall     chan bool
	Fet_stall chan bool
	MRegOk    chan bool
}

type DecCache struct {
	Lcystall  chan bool
	StallData chan []byte
	LastIns   chan uint32
}

func imdCheck(ins uint32) bool {
	ins = ins << 5
	ins = ins >> 31

	return ins == 1
}

func offSetCheck(ins uint32) bool {
	ins = ins << 6
	ins = ins >> 31

	return ins == 1
}

func decodeUnsigned(val uint32, start int, end int) int {
	val = val << start
	val = val >> (start + (31 - end))

	return int(val)
}

func decodeSigned(val uint32, start int, end int) int {
	uval := val << (start + 1)
	uval = uval >> (start + (31 - end) + 1)

	signBit := val << start
	signBit = signBit >> 31

	var sign int
	if signBit == 1 {
		sign = -1
	} else {
		sign = 1
	}

	return int(uval) * sign
}

func modifiedReg(reg int, mRes []c.CaAddr) (bool, int) {

	for i := 0; i < len(mRes); i++ {
		if mRes[i].Loc == reg {
			return true, mRes[i].Val
		}
	}
	return false, 0
}

func getRegVal(targetReg int, regs c.Registers, mRegs []c.CaAddr) int {

	mod, mVal := modifiedReg(targetReg, mRegs)

	var ret int

	if mod {
		ret = mVal
	} else {
		ret = <-regs.Reg[targetReg]
		regs.Reg[targetReg] <- ret
	}

	return ret
}

// Decode the instruction
func Decode(regs c.Registers, flg c.Flags, mem c.Memory,
	buf c.Buffer, bus DecChans, cache DecCache, modRegCa c.Cache) {

	fetData := <-buf.In
	lastCycleStall := <-cache.Lcystall
	stallData := <-cache.StallData
	lastIns := <-cache.LastIns

	stall := <-bus.Stall
	discard := <-bus.Dis

	ins := binary.BigEndian.Uint32(fetData)
	if stall && lastCycleStall {
		ins = lastIns
	}
	if stall {
		lastIns = ins
	}

	opc := ((0b11111 << 27) & ins) >> 27
	opr := op.MatchOpc(uint(opc))

	imd := false
	if opr.Imd {
		imd = imdCheck(ins)
	}

	offSet := false
	if opr.OffSet {
		offSet = offSetCheck(ins)
	}

	var opds []int
	for i := 0; i < 3; i++ {
		opds = append(opds, 0)
	}

	bus.Fet_stall <- stall

	nextIns := 1

	if discard {
		opr = op.Nop
	}

	if stall {
		nextIns = 0
	}

	if opr.Class != "ctf" {
		bus.NIns <- nextIns
		bus.Bran <- 0
	}

	<-bus.MRegOk
	modRegs := <-modRegCa
	modRegCa <- modRegs

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop, op.Hlt:
			bus.NIns <- nextIns
			bus.Bran <- 0

		case op.Jmp:
			opds[0] = decodeSigned(ins, 5, 31)
			if stall {
				bus.NIns <- nextIns
			} else {
				bus.NIns <- opds[0]
			}
			bus.Bran <- 1

		case op.Beq:
			ra := decodeUnsigned(ins, 6, 10)
			opds[0] = getRegVal(ra, regs, modRegs)

			if imd {
				opds[1] = decodeUnsigned(ins, 20, 31)
			} else {
				rb := decodeUnsigned(ins, 20, 24)
				opds[1] = getRegVal(rb, regs, modRegs)
			}

			opds[2] = decodeSigned(ins, 11, 19)
			if stall {
				bus.NIns <- nextIns
			} else {
				bus.NIns <- opds[2]
			}

			bus.Bran <- 2

		case op.Bz:
			ra := decodeUnsigned(ins, 5, 9)
			opds[0] = getRegVal(ra, regs, modRegs)

			opds[1] = decodeSigned(ins, 10, 31)
			if stall {
				bus.NIns <- 0
			} else {
				bus.NIns <- opds[1]
			}

			bus.Bran <- 2
		}

	case "ari", "log":
		if opr == op.Not {
			opds[0] = decodeUnsigned(ins, 6, 10)

			if imd {
				opds[1] = decodeUnsigned(ins, 16, 31)
			} else {
				rs := decodeUnsigned(ins, 11, 15)
				opds[1] = getRegVal(rs, regs, modRegs)
			}
		} else {
			opds[0] = decodeUnsigned(ins, 6, 10)

			rsa := decodeUnsigned(ins, 11, 15)
			opds[1] = getRegVal(rsa, regs, modRegs)

			if imd {
				opds[2] = decodeSigned(ins, 16, 31)
			} else {
				rsb := decodeUnsigned(ins, 16, 20)
				opds[2] = getRegVal(rsb, regs, modRegs)
			}
		}

	case "dat":
		switch opr {
		case op.Ld:
			opds[0] = decodeUnsigned(ins, 7, 11)

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				opds[1] = getRegVal(rsb, regs, modRegs)
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 31)
			} else {
				rsa := decodeUnsigned(ins, 17, 21)
				opds[2] = getRegVal(rsa, regs, modRegs)
			}

		case op.Wrt:
			rsa := decodeUnsigned(ins, 7, 11)
			opds[0] = getRegVal(rsa, regs, modRegs)

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				opds[1] = getRegVal(rsb, regs, modRegs)
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 21)
			} else {
				rd := decodeUnsigned(ins, 17, 31)
				opds[2] = getRegVal(rd, regs, modRegs)
			}

		case op.Mv:
			opds[0] = decodeUnsigned(ins, 6, 10)

			if imd {
				opds[1] = decodeSigned(ins, 11, 31)
			} else {
				rs := decodeUnsigned(ins, 11, 15)
				opds[1] = getRegVal(rs, regs, modRegs)
			}
		}
	}

	var exData []byte
	exData = append(exData, byte(opc))
	for i := 0; i < 3; i++ {
		exData = binary.BigEndian.AppendUint32(exData, uint32(opds[i]))
	}

	if discard {
		exData = make([]byte, 14)
	}

	if stall {
		lastCycleStall = true
	} else {
		lastCycleStall = false
	}

	cache.Lcystall <- lastCycleStall
	cache.StallData <- stallData
	cache.LastIns <- lastIns

	buf.Out <- exData

	flg.DecChk <- true
}
