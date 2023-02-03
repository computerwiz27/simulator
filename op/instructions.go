package op

import "golang.org/x/exp/slices"

type Op byte

const (
	//Arithmetic operations
	ADD  Op = iota //[regA] [regB]
	ADDI           //[regA] [val]
	SUB            //[regA] [regB]
	SUBI           //[regA] [val]
	MUL            //[regA] [regB]
	DIV            //[regA] [regB]

	//Logical operations
	AND //[regA] [regB]
	OR  //[regA] [regB]
	XOR //[regA] [regB]
	LT  //[regA] [regB]
	EQ  //[regA] [regB]

	//Data transfer operations
	LD  //[regA] [mem addr]		load to register
	LDI //[regA] [val]			load imidiate value to register
	WRT //[regA] [mem addr]		write to memory
	MV  //[regA] [regB]			copy from one register to another

	//Controll flow operations
	JMP //[val]					jump to instruction
	BZ  //[regA] [val]			branch if 0
	BEQ //[regA] [regB] [val]	branch if equals
	HLT //						halt
)

var opds = [][]Op{
	//0 operands
	{HLT},

	//1 operand
	{JMP},

	//2 operands
	{ADD, ADDI, SUB, SUBI, MUL, DIV,
		AND, OR, XOR, LT, EQ,
		LD, LDI, WRT, MV,
		BZ,
	},

	//3 operands
	{BEQ},
}

func OperandsNo(op Op) int {
	var n int

	for i := 0; i <= 3; i++ {
		if slices.Contains(opds[i], op) {
			n = i
			break
		}
	}

	return n
}
