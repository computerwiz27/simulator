package op

type op struct {
	opc   int
	class string
	opNo  int
}

const (
	//Arithmetic operations
	ADD  int = iota //[regS] [regA] [regB]
	ADDI            //[regS] [regA] [val]
	SUB             //[regS] [regA] [regB]
	SUBI            //[regS] [regA] [val]
	MUL             //[regS] [regA] [regB]
	DIV             //[regS] [regA] [regB]

	//Logical operations
	AND //[regS] [regA] [regB]
	OR  //[regS] [regA] [regB]
	XOR //[regS] [regA] [regB]
	CMP //[regS] [regA] [regB]

	//Data transfer operations
	LD  //[regA] [mem addr]		load to register
	LDI //[regA] [val]			load imidiate value to register
	MV  //[regA] [regB]			copy from register B to register A
	WRT //[regA] [regB]			write value of regA to memory address in regB

	//Controll flow operations
	JMP //[val]					jump to instruction
	BEQ //[regA] [regB] [val]	branch if equals
	HLT //						halt
)

var Add = op{
	opc:   ADD,
	class: "ari",
	opNo:  3,
}

var Sub = op{
	opc:   SUB,
	class: "ari",
	opNo:  3,
}

var Subi = op{
	opc:   SUBI,
	class: "ari",
	opNo:  3,
}

var Mul = op{
	opc:   MUL,
	class: "ari",
	opNo:  3,
}

var Div = op{
	opc:   DIV,
	class: "ari",
	opNo:  3,
}

var And = op{
	opc:   AND,
	class: "log",
	opNo:  3,
}

var Or = op{
	opc:   OR,
	class: "log",
	opNo:  3,
}

var Xor = op{
	opc:   XOR,
	class: "log",
	opNo:  3,
}

var Cmp = op{
	opc:   CMP,
	class: "log",
	opNo:  3,
}

var Ld = op{
	opc:   LD,
	class: "dat",
	opNo:  2,
}

var Ldi = op{
	opc:   LDI,
	class: "dat",
	opNo:  2,
}

var Mv = op{
	opc:   MV,
	class: "dat",
	opNo:  2,
}

var Wrt = op{
	opc:   WRT,
	class: "dat",
	opNo:  2,
}

var Jmp = op{
	opc:   LD,
	class: "ctf",
	opNo:  1,
}

var Beq = op{
	opc:   BEQ,
	class: "ctf",
	opNo:  3,
}

var Hlt = op{
	opc:   BEQ,
	class: "ctf",
	opNo:  0,
}
