package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

func Fetch(regs Registers, flg Flags, mem []string) {
	line := <-regs.pc

	tokens := strings.Split(mem[line], " ")

	Decode(regs, flg, mem, tokens, line)
}

func oprToInt(opr string) int {
	ret, err := strconv.Atoi(opr)

	if err != nil {
		return ret
	}

	ret, _ = strconv.Atoi(opr[3:])

	return ret
}

func Decode(regs Registers, flg Flags, mem []string, tokens []string, line uint32) {
	var opc op.Op

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
		fmt.Printf("Error: Token %s not on line %d recognised\n", tokens[0], line)
		break
	}

	var oprs []int

	for i := 0; i < op.OperandsNo(opc); i++ {
		oprs[i] = oprToInt(tokens[i+1])
	}

	Execute(regs, flg, mem, opc, oprs)
}
