package components

import (
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
	"golang.org/x/exp/slices"
)

func increment(pc chan uint32) {
	n := <-pc
	pc <- (n + 1)
}

func Execute(regs Registers, flg Flags, mem Memory, opc op.Op, oprs []int) {
	if !slices.Contains(op.ControllFLowOps, opc) {
		increment(regs.pc)
	}

	switch opc {
	case op.ADD:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA + regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.ADDI:
		regA, val := <-regs.reg[oprs[0]], int32(oprs[1])
		res := regA + val
		regs.reg[oprs[0]] <- res

	case op.SUB:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA - regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.SUBI:
		regA, val := <-regs.reg[oprs[0]], int32(oprs[1])
		res := regA - val
		regs.reg[oprs[0]] <- res

	case op.MUL:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA * regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.DIV:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA / regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.AND:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA & regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB
	case op.OR:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA | regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.XOR:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		res := regA ^ regB
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.LT:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		var res int32 = 0
		if regA < regB {
			res = 1
		}
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.EQ:
		regA, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		var res int32 = 0
		if regA == regB {
			res = 1
		}
		regs.reg[oprs[0]] <- res
		regs.reg[oprs[1]] <- regB

	case op.LD:
		_, tmp := <-regs.reg[oprs[0]], <-mem
		lines := strings.Split(string(tmp), "\n")
		val, _ := strconv.Atoi(lines[oprs[1]])
		regs.reg[oprs[0]] <- int32(val)
		mem <- tmp

	case op.LDI:
		<-regs.reg[oprs[0]]
		regs.reg[oprs[0]] <- int32(oprs[1])

	case op.WRT:
		WriteBack(regs, mem, oprs[0], oprs[1])

	case op.MV:
		_, regB := <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- regB
		regs.reg[oprs[1]] <- regB

	case op.JMP:
		<-regs.pc
		regs.pc <- uint32(oprs[0])

	case op.BZ:
		pc, regA := <-regs.pc, <-regs.reg[oprs[0]]
		if regA == 0 {
			regs.pc <- uint32(<-regs.reg[oprs[1]])
		} else {
			regs.pc <- pc
			increment(regs.pc)
		}

	case op.BEQ:
		pc, regA, regB := <-regs.pc, <-regs.reg[oprs[0]], <-regs.reg[oprs[1]]
		if regA == regB {
			regs.pc <- uint32(<-regs.reg[oprs[1]])
		} else {
			regs.pc <- pc
			increment(regs.pc)
		}

	case op.HLT:
		flg.halt <- true
	}

	Fetch(regs, flg, mem)
}

func WriteBack(regs Registers, mem Memory, reg int, loc int) {
	lines := strings.Split(string(<-mem), "\n")

	val := <-regs.reg[reg]
	regs.reg[reg] <- val

	lines[loc] = strconv.Itoa(int(val)) + "\n"

	var tmp []byte
	for i := range lines {
		tmp = append(tmp, []byte(lines[i])...)
	}

	mem <- tmp
}
