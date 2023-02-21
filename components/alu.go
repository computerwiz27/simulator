package components

import (
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

// Increment the program counter channel
func increment(pc chan uint) {
	n := <-pc
	pc <- (n + 4)
}

// Execute given instruction
func Execute(regs Registers, flg Flags, mem Memory, prog Prog, ins op.Op, vars [3]int) {
	if ins.Class != "ctf" {
		increment(regs.pc)
	}

	var result int
	var desReg int
	var memLoc int

	reg := false
	wrt := false

	switch ins.Class {
	case "ari":
		switch ins {
		case op.Add:
			result = vars[1] + vars[2]
		case op.Sub:
			result = vars[1] - vars[2]
		case op.Mul:
			result = vars[1] * vars[2]
		case op.Div:
			result = vars[1] / vars[2]
		}
		desReg = vars[0]
		reg = true

	case "log":
		switch ins {
		case op.And:
			result = vars[1] & vars[2]
		case op.Or:
			result = vars[1] | vars[2]
		case op.Xor:
			result = vars[1] ^ vars[2]
		case op.Cmp:
			if vars[1] < vars[2] {
				result = -1
			} else if vars[1] == vars[2] {
				result = 0
			} else {
				result = 1
			}
		}
		desReg = vars[0]
		reg = true

	case "dat":
		switch ins {
		case op.Ld:
			result = vars[1]
			desReg = vars[0]
			reg = true
		case op.Mv:
			result = vars[1]
			desReg = vars[0]
			reg = true
		case op.Wrt:
			result = vars[1]
			memLoc = vars[0]
			wrt = true
		}

	case "ctf":
		switch ins {
		case op.Jmp:
			<-regs.pc
			regs.pc <- uint(vars[0]) * 4
		case op.Beq:
			if vars[0] == vars[1] {
				<-regs.pc
				regs.pc <- uint(vars[2]) * 4
			} else {
				increment(regs.pc)
			}
		case op.Bz:
			if vars[0] == 0 {
				<-regs.pc
				regs.pc <- uint(vars[1]) * 4
			} else {
				increment(regs.pc)
			}
		case op.Hlt:
			flg.halt <- true
			return
		}
	}

	if reg {
		WriteBack(regs, flg, mem, prog, desReg, result)
	} else if wrt {
		WriteToMemory(regs, flg, mem, prog, memLoc, result)
	} else if ins != op.Hlt {
		Fetch(regs, flg, mem, prog)
	}
}

func WriteBack(regs Registers, flg Flags, mem Memory, prog Prog, des, val int) {
	<-regs.reg[des]
	regs.reg[des] <- val

	Fetch(regs, flg, mem, prog)
}

// Write register to memory
func WriteToMemory(regs Registers, flg Flags, mem Memory, prog Prog, loc, val int) {
	lines := strings.Split(string(<-mem), "\n")

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

	Fetch(regs, flg, mem, prog)
}
