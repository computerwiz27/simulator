package components

import "encoding/binary"

type WbChans struct {
	wbMRegs    chan bool
	ex_mRegsOk chan bool
}

func removeModReg(tReg int, modRegCa ModRegCache) {
	modRegs := <-modRegCa

	for i := 0; i < len(modRegs); i++ {
		if modRegs[i].reg == int(tReg) {
			modRegs[i] = modRegs[len(modRegs)-1]
			modRegs = modRegs[:len(modRegs)-1]
			break
		}
	}

	modRegCa <- modRegs
}

func WriteBack(regs Registers, flg Flags, buf Buffer, bus WbChans,
	modRegCa ModRegCache) {

	memData := <-buf.in

	write := false
	if memData[0] == 1 {
		write = true
	}

	des := binary.BigEndian.Uint32(memData[1:5])

	uval := binary.BigEndian.Uint32(memData[5:9])
	val := int(uval)

	dumpModRegs := <-bus.wbMRegs
	if dumpModRegs {
		modReg := <-modRegCa
		modRegCa <- modReg

		for i := 0; i < len(modReg); i++ {
			removeModReg(modReg[i].reg, modRegCa)

			<-regs.reg[modReg[i].reg]
			regs.reg[modReg[i].reg] <- modReg[i].val
		}

		if write {
			<-regs.reg[des]
			regs.reg[des] <- val
		}

		bus.ex_mRegsOk <- true
		flg.wbChk <- true
		return
	}

	if !write {
		bus.ex_mRegsOk <- true
		flg.wbChk <- true
		return
	}

	removeModReg(int(des), modRegCa)

	<-regs.reg[des]
	regs.reg[des] <- val

	bus.ex_mRegsOk <- true

	flg.wbChk <- true
}
