package components

import (
	"strings"
)

func Fetch(lines []string, regs Registers, flg Flags) {
	instr := <-regs.pc

	tokens := strings.Split(lines[instr], " ")

	switch tokens[0] {
	case "ADD":

	}
}
