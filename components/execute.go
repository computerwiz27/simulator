package components

import (
	"encoding/binary"
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
func Execute(regs Registers, flg Flags, mem Memory, prog Memory,
	dec_ex Buffer /*, ex_wb Buffer, ex_wm Buffer*/) {

	var dec_data []byte
	select {
	case dec_data = <-dec_ex:

	default:
		for i := 0; i < 13; i++ {
			dec_data[i] = 0
		}
	}

	opc := uint(dec_data[0])
	opr := op.MatchOpc(opc)

	var opds []int
	for i := 1; i < 13; i += 4 {
		uopds := binary.BigEndian.Uint32(dec_data[i : i+4])
		opds = append(opds, int(uopds))
	}

	if opr.Class != "ctf" {
		increment(regs.pc)
	}

	wb := false
	wmem := false

	var result int
	var desReg int
	var memLoc int

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop:

		case op.Hlt:
			flg.halt <- true

		case op.Jmp:
			n := <-regs.pc
			regs.pc <- uint(int(n) + opds[0])

		case op.Beq:
			if opds[0] == opds[2] {
				n := <-regs.pc
				regs.pc <- uint(int(n) + opds[1])
			} else {
				increment(regs.pc)
			}

		case op.Bz:
			if opds[0] == 0 {
				n := <-regs.pc
				regs.pc <- uint(int(n) + opds[1])
			} else {
				increment(regs.pc)
			}
		}

	case "ari":
		wb = true
		desReg = opds[0]

		switch opr {
		case op.Add:
			result = opds[1] + opds[2]

		case op.Sub:
			result = opds[1] - opds[2]

		case op.Mul:
			result = opds[1] * opds[2]

		case op.Div:
			result = opds[1] / opds[2]
		}

	case "log":
		wb = true
		desReg = opds[0]

		switch opr {
		case op.And:
			result = opds[1] & opds[2]

		case op.Or:
			result = opds[1] | opds[2]

		case op.Xor:
			result = opds[1] ^ opds[2]

		case op.Not:
			result = ^opds[1]

		case op.Cmp:
			if opds[1] > opds[2] {
				result = 1
			} else if opds[1] == opds[2] {
				result = 0
			} else {
				result = -1
			}
		}

	case "dat":
		switch opr {
		case op.Ld:
			wb = true
			desReg = opds[0] + opds[1]

			result = opds[2]

		case op.Wrt:
			wmem = true
			memLoc = opds[2] + opds[1]

			result = opds[0]

		case op.Mv:
			wb = true
			desReg = opds[0]

			result = opds[1]
		}
	}

	if wb {
		WriteBack(regs, flg, mem, prog, desReg, result)
	} else if wmem {
		WriteToMemory(regs, flg, mem, prog, memLoc, result)
	} else if opr != op.Hlt {
		//flg.exChk <- true
		flg.halt <- true
	}
}

func WriteBack(regs Registers, flg Flags, mem Memory, prog Memory, des, val int) {
	<-regs.reg[des]
	regs.reg[des] <- val

	//flg.exChk <- true
}

// Write register to memory
func WriteToMemory(regs Registers, flg Flags, mem Memory, prog Memory, loc, val int) {
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

	//flg.exChk <- true
}
