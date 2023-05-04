package stages

import (
	"github.com/computerwiz27/simulator/binary"
	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

type DecChans struct {
	NIns          chan []int
	Bran          chan int
	Dis           chan bool
	Stall         chan bool
	Fet_stall     chan bool
	Fet_forkCount chan int
	UpdtCount     chan bool
	MRegOk        chan bool
}

type DecCache struct {
	Lcystall   chan bool
	StallData  chan []byte
	LastIns    chan []uint32
	LastCounts chan []uint32
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

func decodeIns(ins uint32, regs c.Registers, modRegs []c.CaAddr, retReg string) (
	op.Op, []int, int, int, []int) {

	var dReg []int
	var sReg []int

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

	nextIns := 1
	branch := 0

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop, op.Hlt:

		case op.Jmp:
			opds[0] = decodeSigned(ins, 5, 31)
			nextIns = opds[0]

			branch = 1

		case op.Beq:
			ra := decodeUnsigned(ins, 6, 10)
			sReg = append(sReg, ra)
			opds[0] = getRegVal(ra, regs, modRegs)

			if imd {
				opds[1] = decodeUnsigned(ins, 20, 31)
			} else {
				rb := decodeUnsigned(ins, 20, 24)
				sReg = append(sReg, rb)
				opds[1] = getRegVal(rb, regs, modRegs)
			}

			opds[2] = decodeSigned(ins, 11, 19)
			nextIns = opds[2]

			branch = 2

		case op.Bz:
			ra := decodeUnsigned(ins, 5, 9)
			sReg = append(sReg, ra)
			opds[0] = getRegVal(ra, regs, modRegs)

			opds[1] = decodeSigned(ins, 10, 31)
			nextIns = opds[1]

			branch = 2
		}

	case "ari", "log":
		if opr == op.Not {
			opds[0] = decodeUnsigned(ins, 6, 10)
			dReg = append(sReg, opds[0])

			if imd {
				opds[1] = decodeUnsigned(ins, 16, 31)
			} else {
				rs := decodeUnsigned(ins, 11, 15)
				sReg = append(sReg, rs)
				opds[1] = getRegVal(rs, regs, modRegs)
			}
		} else {
			opds[0] = decodeUnsigned(ins, 6, 10)
			dReg = append(sReg, opds[0])

			rsa := decodeUnsigned(ins, 11, 15)
			sReg = append(sReg, rsa)
			opds[1] = getRegVal(rsa, regs, modRegs)

			if imd {
				opds[2] = decodeSigned(ins, 16, 31)
			} else {
				rsb := decodeUnsigned(ins, 16, 20)
				sReg = append(sReg, rsb)
				opds[2] = getRegVal(rsb, regs, modRegs)
			}
		}

	case "dat":
		switch opr {
		case op.Ld:
			opds[0] = decodeUnsigned(ins, 7, 11)
			dReg = append(sReg, opds[0])

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				sReg = append(sReg, rsb)
				opds[1] = getRegVal(rsb, regs, modRegs)
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 31)
			} else {
				rsa := decodeUnsigned(ins, 17, 21)
				sReg = append(sReg, rsa)
				opds[2] = getRegVal(rsa, regs, modRegs)
			}

		case op.Wrt:
			rsa := decodeUnsigned(ins, 7, 11)
			sReg = append(sReg, rsa)
			opds[0] = getRegVal(rsa, regs, modRegs)

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				sReg = append(sReg, rsb)
				opds[1] = getRegVal(rsb, regs, modRegs)
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 21)
			} else {
				rd := decodeUnsigned(ins, 17, 31)
				dReg = append(sReg, rd)
				opds[2] = getRegVal(rd, regs, modRegs)
			}

		case op.Mv:
			opds[0] = decodeUnsigned(ins, 6, 10)
			dReg = append(sReg, opds[0])

			if imd {
				opds[1] = decodeSigned(ins, 11, 31)
			} else {
				rs := decodeUnsigned(ins, 11, 15)
				sReg = append(sReg, rs)
				opds[1] = getRegVal(rs, regs, modRegs)
			}
		}
	}

	var ret []int
	if retReg == "s" {
		ret = sReg
	} else if retReg == "d" {
		ret = dReg
	}

	return opr, opds, nextIns, branch, ret
}

func issue(opr1 op.Op, opr2 op.Op, opds1 []int, opds2 []int) (op.Op, op.Op,
	[]int, []int, byte) {

	datFirst := byte(1)

	datOpr := opr1
	datOpds := opds1

	brOpr := opr2
	brOpds := opds2

	if opr1 == op.Beq || opr1 == op.Bz ||
		opr2 == op.Ld || opr2 == op.Wrt {

		datFirst = 0

		datOpr = opr2
		datOpds = opds2

		brOpr = opr1
		brOpds = opds1
	}

	return datOpr, brOpr, datOpds, brOpds, datFirst
}

// Decode the instruction
func Decode(regs c.Registers, flg c.Flags, mem c.Memory,
	buf c.Buffer, bus DecChans, cache DecCache, modRegCa c.Cache) {

	fetData := <-buf.In
	lastCycleStall := <-cache.Lcystall
	stallData := <-cache.StallData
	lastIns := <-cache.LastIns
	lastCounts := <-cache.LastCounts

	stall := <-bus.Stall
	discard := <-bus.Dis

	bus.Fet_stall <- stall

	<-bus.MRegOk
	modRegs := <-modRegCa
	modRegCa <- modRegs

	ins1 := binary.BigEndian.Uint32(fetData[0:4])
	count1 := binary.BigEndian.Uint32(fetData[4:8])
	ins2 := binary.BigEndian.Uint32(fetData[8:12])
	count2 := binary.BigEndian.Uint32(fetData[12:16])
	if stall && lastCycleStall {
		ins1 = lastIns[0]
		count1 = lastCounts[0]
		ins2 = lastIns[1]
		count2 = lastCounts[1]
	}
	if stall {
		lastIns[0] = ins1
		lastCounts[0] = count1
		lastIns[1] = ins2
		lastCounts[1] = count2
	}

	opr1, opds1, nextIns1, branch1, dReg1 := decodeIns(ins1, regs, modRegs, "d")

	opr2, opds2, nextIns2, branch2, sReg2 := decodeIns(ins2, regs, modRegs, "s")
	nextIns2++

	nextIns := []int{nextIns1, nextIns2}

	dependence := false
	for i := 0; i < len(sReg2); i++ {
		if len(dReg1) == 0 {
			break
		}
		if sReg2[i] == dReg1[0] {
			dependence = true
			break
		}
	}

	updateCounter := false

	if dependence {
		opr2 = op.Nop
		nextIns[0] = -1
		nextIns[1] = 0
		updateCounter = true
	}

	if (opr1 == op.Ld && opr2 == op.Ld) || (opr1 == op.Wrt && opr2 == op.Wrt) {
		opr2 = op.Nop
		nextIns[0] = -1
		nextIns[1] = 0
		updateCounter = true
	}

	if opr1 == op.Jmp && int(count1)+nextIns1 != int(count2) {
		opr2 = op.Nop
		nextIns[0] = nextIns1 - 2
		nextIns[1] = nextIns1 - 1
		updateCounter = true
	}

	branch := 0
	if branch1 > 0 {
		if branch2 > 0 || int(count1)+nextIns1 != int(count2) {
			opr2 = op.Nop
			nextIns[0] = nextIns1 - 2
			nextIns[1] = nextIns1 - 1
			updateCounter = true
			branch = branch1
		}
	} else if branch2 > 0 {
		nextIns[0] = nextIns2 - 2
		nextIns[1] = nextIns2 - 1
		updateCounter = true
		branch = branch2
	}

	if branch1 == 2 {
		bus.Fet_forkCount <- int(count1)
	} else if branch2 == 2 {
		bus.Fet_forkCount <- int(count2)
	} else {
		bus.Fet_forkCount <- 0
	}

	bus.NIns <- nextIns
	bus.Bran <- branch
	bus.UpdtCount <- updateCounter

	datOpr, brOpr, datOpds, brOpds, datFirst := issue(opr1, opr2, opds1, opds2)

	var datData []byte
	datData = append(datData, byte(datOpr.Opc))
	for i := 0; i < 3; i++ {
		datData = binary.BigEndian.AppendUint32(datData, uint32(datOpds[i]))
	}

	var brData []byte
	brData = append(brData, byte(brOpr.Opc))
	for i := 0; i < 3; i++ {
		brData = binary.BigEndian.AppendUint32(brData, uint32(brOpds[i]))
	}

	exData := append(datData, brData...)
	exData = append(exData, datFirst)

	if discard {
		exData = make([]byte, 28)
	}

	if stall {
		lastCycleStall = true
	} else {
		lastCycleStall = false
	}

	cache.Lcystall <- lastCycleStall
	cache.StallData <- stallData
	cache.LastIns <- lastIns
	cache.LastCounts <- lastCounts

	buf.Out <- exData

	flg.DecChk <- true
}
