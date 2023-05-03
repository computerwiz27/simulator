package stages

import (
	"encoding/binary"

	c "github.com/computerwiz27/simulator/components"
)

type WbChans struct {
	Ex_mRegsOk chan bool
}

func removeModReg(tReg int, val int, modRegCa c.Cache) {
	modRegs := <-modRegCa

	for i := 0; i < len(modRegs); i++ {
		if modRegs[i].Loc == int(tReg) && modRegs[i].Val == val {
			modRegs[i] = modRegs[len(modRegs)-1]
			modRegs = modRegs[:len(modRegs)-1]
		}
	}
	modRegCa <- modRegs
}

func WriteBack(regs c.Registers, flg c.Flags, buf c.Buffer, bus WbChans,
	modRegCa c.Cache) {

	memData := <-buf.In

	write1 := false
	if memData[0] == 1 {
		write1 = true
	}

	des1 := binary.BigEndian.Uint32(memData[1:5])

	uval1 := binary.BigEndian.Uint32(memData[5:9])
	val1 := int(int32(uval1))

	write2 := false
	if memData[9] == 1 {
		write2 = true
	}

	des2 := binary.BigEndian.Uint32(memData[10:15])

	uval2 := binary.BigEndian.Uint32(memData[15:19])
	val2 := int(int32(uval2))

	oneFirst := true
	if memData[19] == 0 {
		oneFirst = false
	}

	if !(write1 || write2) {
		bus.Ex_mRegsOk <- true
		flg.WbChk <- true
		return
	}

	if oneFirst {
		removeModReg(int(des1), val1, modRegCa)

		<-regs.Reg[des1]
		regs.Reg[des1] <- val1

		removeModReg(int(des2), val2, modRegCa)

		<-regs.Reg[des2]
		regs.Reg[des2] <- val2
	} else {
		removeModReg(int(des2), val2, modRegCa)

		<-regs.Reg[des2]
		regs.Reg[des2] <- val2

		removeModReg(int(des1), val1, modRegCa)

		<-regs.Reg[des1]
		regs.Reg[des1] <- val1
	}

	bus.Ex_mRegsOk <- true

	flg.WbChk <- true
}
