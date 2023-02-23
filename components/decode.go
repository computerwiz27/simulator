package components

import (
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

// Decode the instruction
func Decode(regs Registers, flg Flags, mem Memory, prog Prog, tokens [4]int) {
	ins := op.MatchOpc(tokens[0])
	opc := ins.Opc
	var vars [3]int

	switch ins.Class {

	case "ari", "log":
		vars[0] = tokens[1]
		vars[1] = <-regs.reg[tokens[2]]
		regs.reg[tokens[2]] <- vars[1]

		if ins == op.Addi || ins == op.Subi {
			opc--
			ins = op.MatchOpc(opc)

			vars[2] = tokens[3]
		} else {
			vars[2] = <-regs.reg[tokens[3]]
			regs.reg[tokens[3]] <- vars[2]
		}

	case "dat":
		switch ins {
		case op.Ld:
			tmp := <-mem
			loc := <-regs.reg[tokens[2]]

			vars[0] = tokens[1]

			lines := strings.Split(string(tmp), "\n")
			val, _ := strconv.Atoi(lines[loc])
			vars[1] = val

			mem <- tmp
			regs.reg[tokens[2]] <- loc

		case op.Ldi:
			opc--
			ins = op.MatchOpc(opc)

			tmp := <-mem

			vars[0] = tokens[1]

			lines := strings.Split(string(tmp), "\n")
			val, _ := strconv.Atoi(lines[tokens[2]])
			vars[1] = val

			mem <- tmp

		case op.Mv:
			vars[0] = tokens[1]

			vars[1] = <-regs.reg[tokens[2]]
			regs.reg[tokens[2]] <- vars[1]

		case op.Mvi:
			opc--
			ins = op.MatchOpc(opc)

			vars[0] = tokens[1]
			vars[1] = tokens[2]

		case op.Wrt:
			vars[0] = <-regs.reg[tokens[1]]
			regs.reg[tokens[1]] <- vars[0]

			vars[1] = <-regs.reg[tokens[2]]
			regs.reg[tokens[2]] <- vars[1]

		case op.Wrti:
			opc--
			ins = op.MatchOpc(opc)

			vars[0] = tokens[1]

			vars[1] = <-regs.reg[tokens[2]]
			regs.reg[tokens[2]] <- vars[1]
		}

		vars[2] = 0

	case "ctf":
		switch ins {
		case op.Jmp:
			vars[0] = tokens[1]
			vars[1] = 0
			vars[2] = 0

		case op.Beq:
			vars[0] = <-regs.reg[tokens[1]]
			regs.reg[tokens[1]] <- vars[0]
			vars[1] = <-regs.reg[tokens[2]]
			regs.reg[tokens[2]] <- vars[1]
			vars[2] = tokens[3]

		case op.Bz:
			vars[0] = <-regs.reg[tokens[1]]
			regs.reg[tokens[1]] <- vars[0]
			vars[1] = tokens[2]
			vars[2] = 0

		case op.Hlt:
			vars[0] = 0
			vars[1] = 0
			vars[2] = 0
		}

	}

	Execute(regs, flg, mem, prog, ins, vars)
}
