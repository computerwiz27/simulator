package components

import (
	"strconv"

	"github.com/computerwiz27/simulator/op"
	"golang.org/x/exp/slices"
)

func increment(pc chan uint32) {
	pc <- (<-pc + 1)
}

func Execute(regs Registers, flg Flags, mem []string, opc op.Op, oprs []int) {
	if !slices.Contains(op.ControllFLowOps, opc) {
		increment(regs.pc)
	}

	switch opc {
	case op.ADD:
		res := <-regs.reg[oprs[0]] + <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.ADDI:
		res := <-regs.reg[oprs[0]] + int32(oprs[1])
		regs.reg[oprs[0]] <- res

	case op.SUB:
		res := <-regs.reg[oprs[0]] - <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.SUBI:
		res := <-regs.reg[oprs[0]] - int32(oprs[1])
		regs.reg[oprs[0]] <- res

	case op.MUL:
		res := <-regs.reg[oprs[0]] * <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.DIV:
		res := <-regs.reg[oprs[0]] / <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.AND:
		res := <-regs.reg[oprs[0]] & <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.OR:
		res := <-regs.reg[oprs[0]] | <-regs.reg[oprs[1]]
		regs.reg[oprs[0]] <- res

	case op.XOR:
		res := <-regs.reg[oprs[0]] ^ <-regs.reg[oprs[0]]
		regs.reg[oprs[0]] <- res

	case op.LT:
		var res int32 = 0
		if <-regs.reg[oprs[0]] < <-regs.reg[oprs[1]] {
			res = 1
		}
		regs.reg[oprs[0]] <- res

	case op.EQ:
		var res int32 = 0
		if <-regs.reg[oprs[0]] == <-regs.reg[oprs[1]] {
			res = 1
		}
		regs.reg[oprs[0]] <- res

	case op.LD:
		val, _ := strconv.Atoi(mem[oprs[1]])
		regs.reg[oprs[0]] <- int32(val)

	case op.LDI:
		regs.reg[oprs[0]] <- int32(oprs[1])

	case op.WRT:
		mem[oprs[1]] = strconv.Itoa(int(<-regs.reg[oprs[0]]))

	case op.MV:
		regs.reg[oprs[0]] <- <-regs.reg[oprs[1]]

	case op.JMP:
		regs.pc <- uint32(oprs[0])

	case op.BZ:
		if <-regs.reg[oprs[0]] == 0 {
			regs.pc <- uint32(<-regs.reg[oprs[1]])
		} else {
			increment(regs.pc)
		}

	case op.BEQ:
		if <-regs.reg[oprs[0]] == <-regs.reg[oprs[1]] {
			regs.pc <- uint32(<-regs.reg[oprs[2]])
		} else {
			increment(regs.pc)
		}

	case op.HLT:
		flg.halt <- true
	}
}
