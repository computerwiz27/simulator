package compiler

import (
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/op"
)

func oprToint(opr string) int {
	ret, err := strconv.Atoi(opr)

	//if there is no error the operand string only contained a number
	if err == nil {
		return ret
	}

	//otherwise the operand is of the form "regX"
	ret, _ = strconv.Atoi(opr[3:])

	return ret
}

func Asemble(file []byte) []int {
	var memory []int

	lines := strings.Split(string(file), "\n")

	for i := range lines {
		tokens := strings.Split(lines[i], " ")

		ins := op.MatchName(tokens[0])

		memory = append(memory, ins.Opc)

		for j := 1; j < 4; j++ {
			if j > ins.OpNo {
				memory = append(memory, 0)
			} else {
				memory = append(memory, oprToint(tokens[j]))
			}
		}
	}

	return memory
}
