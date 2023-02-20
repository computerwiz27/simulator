package components

import "github.com/computerwiz27/simulator/op"

// Fetch the next instruction from memory
func Fetch(regs Registers, flg Flags, mem Memory, prog Prog) {
	counter := <-regs.pc
	tmp := <-prog

	var tokens [4]int
	for i := 0; i < 4; i++ {
		tokens[i] = tmp[counter+1]
	}

	regs.pc <- counter
	prog <- tmp

	Decode(regs, flg, mem, prog, tokens)
}

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
		vars[0] = tokens[1]

		if ins == op.Ldi {
			vars[1] = tokens[2]
		} else {
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

	Execute(regs, flg, mem, prog, opc, vars)
}
