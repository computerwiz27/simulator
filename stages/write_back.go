package stages

import (
	"encoding/binary"

	c "github.com/computerwiz27/simulator/components"
)

type WbChans struct {
	WbMRegs    chan bool
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

	write := false
	if memData[0] == 1 {
		write = true
	}

	des := binary.BigEndian.Uint32(memData[1:5])

	uval := binary.BigEndian.Uint32(memData[5:9])
	val := int(int32(uval))

	dumpModRegs := <-bus.WbMRegs
	if dumpModRegs {
		modReg := <-modRegCa
		modRegCa <- modReg

		for i := 0; i < len(modReg); i++ {
			<-regs.Reg[modReg[i].Loc]
			regs.Reg[modReg[i].Loc] <- modReg[i].Val
		}

		if write {
			<-regs.Reg[des]
			regs.Reg[des] <- val
		}

		bus.Ex_mRegsOk <- true
		flg.WbChk <- true
		return
	}

	if !write {
		bus.Ex_mRegsOk <- true
		flg.WbChk <- true
		return
	}

	removeModReg(int(des), val, modRegCa)

	<-regs.Reg[des]
	regs.Reg[des] <- val

	bus.Ex_mRegsOk <- true

	flg.WbChk <- true
}
