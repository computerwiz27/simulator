package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

// Fetch the next instruction from memory
func Fetch(regs Registers, flg Flags, mem Memory) {
	line := <-regs.pc
	regs.pc <- line

	tmp := <-mem
	mem <- tmp

	lines := strings.Split(string(tmp), "\n")

	tokens := strings.Split(lines[line], " ")

	Decode(regs, flg, mem, tokens, line)
}

// Convert operands from string to int
func oprToInt(opr string) int {
	ret, err := strconv.Atoi(opr)

	//if there is no error the operand string only contained a number
	if err == nil {
		return ret
	}

	//otherwise the operand is of the form "regX"
	ret, _ = strconv.Atoi(opr[3:])

	return ret
}

// Decode the instruction
func Decode(regs Registers, flg Flags, mem Memory, tokens []string, line uint32) {
	var opc op.Op

	//Match the first token with an instruction
	switch tokens[0] {
	case "ADD", "add":
		opc = op.ADD

	case "ADDI", "addi":
		opc = op.ADDI

	case "SUB", "sub":
		opc = op.SUB

	case "SUBI", "subi":
		opc = op.SUBI

	case "MUL", "mul":
		opc = op.MUL

	case "DIV", "div":
		opc = op.DIV

	case "AND", "and":
		opc = op.AND

	case "OR", "or":
		opc = op.OR

	case "XOR", "xor":
		opc = op.XOR

	case "LT", "lt":
		opc = op.LT

	case "EQ", "eq":
		opc = op.EQ

	case "LD", "ld":
		opc = op.LD

	case "LDI", "ldi":
		opc = op.LDI

	case "WRT", "wrt":
		opc = op.WRT

	case "MV", "mv":
		opc = op.MV

	case "JMP", "jmp":
		opc = op.JMP

	case "BZ", "bz":
		opc = op.BZ

	case "BEQ", "beq":
		opc = op.BEQ

	case "HLT", "hlt":
		opc = op.HLT

	default:
		fmt.Printf("Error: Token \"%s\" on line %d is not recognised\n", tokens[0], line+1)
		opc = op.HLT
	}

	//Get the instruction's operands
	var oprs []int
	for i := 0; i < op.OperandsNo(opc); i++ {
		oprs = append(oprs, oprToInt(tokens[i+1]))
	}

	Execute(regs, flg, mem, opc, oprs)
}
