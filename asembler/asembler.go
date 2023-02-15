package compiler

import (
	"strings"

	"github.com/computerwiz27/simulator/op"
)

func match(insStr string) op.Op {
	for ins := range op.Instructions {
		if insStr == op.Instructions[ins].Name {
			return op.Instructions[ins]
		}
	}

	return op.Hlt
}

func Asemble(file []byte) []int {
	var memory []int

	lines := strings.Split(string(file), "\n")

	for i := range lines {
		tokens := strings.Split(lines[i], " ")

		ins := match(tokens[0])

		for j := 1; j < 4; j++ {
			if j+1 >= ins.OpNo {
				memory = append(memory, 0)
			}
		}
	}

	return memory
}
