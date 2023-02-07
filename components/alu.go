package components

import (
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
	"golang.org/x/exp/slices"
)

// Increment the program counter channel
func increment(pc chan uint32) {
	n := <-pc
	pc <- (n + 1)
}

// Execute given instruction
func Execute(regs Registers, flg Flags, mem Memory, opc op.Op, oprs []int) {
	//Increment pc if it isn't a controll flow instruction
	if !slices.Contains(op.ControllFLowOps, opc) {
		increment(regs.pc)
	}

	//Execute the instruction
	//! take care that the value in registers that are not output are not modified
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

	//Only WRT modifies memory
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

	//Don't call fetch given the halt instruction
	if opc != op.HLT {
		Fetch(regs, flg, mem)
	}
}

// Write register to memory
func WriteBack(regs Registers, mem Memory, regA int, regB int) {
	lines := strings.Split(string(<-mem), "\n")

	val := <-regs.reg[regA]
	regs.reg[regA] <- val

	loc := <-regs.reg[regB]
	regs.reg[regB] <- loc

	if int(loc) >= len(lines) {
		for i := len(lines); i <= int(loc); i++ {
			lines = append(lines, "")
		}
	}

	lines[loc] = strconv.Itoa(int(val)) + "\n"

	//make a temporary variable and append the new memory bytes
	var tmp []byte
	for i := range lines {
		tmp = append(tmp, []byte(lines[i])...)
		tmp = append(tmp, []byte("\n")...) //add a new line after every line
	}

	mem <- tmp
}
