package components

import (
	"encoding/binary"

	"github.com/computerwiz27/simulator/op"
)

type DecChans struct {
	nIns   chan int
	bran   chan int
	dis    chan bool
	stall  chan bool
	mRegOk chan bool
}

type DecCache struct {
	lcystall  chan bool
	stallData chan []byte
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

func modifiedReg(reg int, mRes []CaAddr) (bool, int) {

	for i := 0; i < len(mRes); i++ {
		if mRes[i].loc == reg {
			return true, mRes[i].val
		}
	}
	return false, 0
}

func getRegVal(targetReg int, regs Registers, mRegs []CaAddr) int {

	mod, mVal := modifiedReg(targetReg, mRegs)

	var ret int

	if mod {
		ret = mVal
	} else {
		ret = <-regs.reg[targetReg]
		regs.reg[targetReg] <- ret
	}

	return ret
}

// Decode the instruction
func Decode(regs Registers, flg Flags, mem Memory,
	buf Buffer, bus DecChans, cache DecCache, modRegCa Cache) {

	fetData := <-buf.in
	lastCycleStall := <-cache.lcystall
	stallData := <-cache.stallData

	ins := binary.BigEndian.Uint32(fetData)

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

	stall := <-bus.stall
	discard := <-bus.dis

	if discard || stall {
		opr = op.Nop
	}

	if opr.Class != "ctf" {
		if stall {
			bus.nIns <- 0
		} else {
			bus.nIns <- 1
		}
		bus.bran <- 0
	}

	<-bus.mRegOk
	modRegs := <-modRegCa
	modRegCa <- modRegs

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop, op.Hlt:
			if stall {
				bus.nIns <- 0
			} else {
				bus.nIns <- 1
			}
			bus.bran <- 0

		case op.Jmp:
			opds[0] = decodeSigned(ins, 5, 31)
			if stall {
				bus.nIns <- 0
			} else {
				bus.nIns <- opds[0] - 1
			}
			bus.bran <- 1

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
				bus.nIns <- 0
			} else {
				bus.nIns <- opds[2] - 1
			}

			bus.bran <- 2

		case op.Bz:
			ra := decodeUnsigned(ins, 5, 9)
			opds[0] = getRegVal(ra, regs, modRegs)

			opds[1] = decodeSigned(ins, 10, 31)
			if stall {
				bus.nIns <- 0
			} else {
				bus.nIns <- opds[1] - 1
			}

			bus.bran <- 2
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

	// if stall {
	// 	if !lastCycleStall {
	// 		stallData = exData
	// 	}
	// 	exData = stallData
	// 	lastCycleStall = true
	// }

	// if !stall && lastCycleStall {
	// 	exData = stallData
	// 	lastCycleStall = false
	// }

	cache.lcystall <- lastCycleStall
	cache.stallData <- stallData
	buf.out <- exData

	flg.decChk <- true
}
