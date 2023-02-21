package op

type Op struct {
	Name  string
	Opc   int
	Class string
	OpNo  int
}

const (
	//Arithmetic operations
	ADD int = iota //[regS] [regA] [regB]
	ADDI
	SUB //[regS] [regA] [regB]
	SUBI
	MUL //[regS] [regA] [regB]
	DIV //[regS] [regA] [regB]

	//Logical operations
	AND //[regS] [regA] [regB]
	OR  //[regS] [regA] [regB]
	XOR //[regS] [regA] [regB]
	CMP //[regS] [regA] [regB]

	//Data transfer operations
	LD //[regA] [mem addr]		load to register
	LDI
	MV   //[regA] [regB]			copy from register B to register A
	WRT  //[regA] [regB]			write value of regB to memory address in regA
	WRTI //[val] [regA]

	//Controll flow operations
	JMP //[val]					jump to instruction
	BEQ //[regA] [regB] [val]	branch if equals
	BZ  //[regA] [val]
	HLT //						halt
)

//Arithmetic operations

var Add = Op{
	Name:  "ADD",
	Opc:   ADD,
	Class: "ari",
	OpNo:  3,
}

var Addi = Op{
	Name:  "ADDI",
	Opc:   ADDI,
	Class: "ari",
	OpNo:  3,
}

var Sub = Op{
	Name:  "SUB",
	Opc:   SUB,
	Class: "ari",
	OpNo:  3,
}

var Subi = Op{
	Name:  "SUBI",
	Opc:   SUBI,
	Class: "ari",
	OpNo:  3,
}

var Mul = Op{
	Name:  "MUL",
	Opc:   MUL,
	Class: "ari",
	OpNo:  3,
}

var Div = Op{
	Name:  "DIV",
	Opc:   DIV,
	Class: "ari",
	OpNo:  3,
}

var And = Op{
	Name:  "AND",
	Opc:   AND,
	Class: "log",
	OpNo:  3,
}

var Or = Op{
	Name:  "OR",
	Opc:   OR,
	Class: "log",
	OpNo:  3,
}

var Xor = Op{
	Name:  "XOR",
	Opc:   XOR,
	Class: "log",
	OpNo:  3,
}

var Cmp = Op{
	Name:  "CMP",
	Opc:   CMP,
	Class: "log",
	OpNo:  3,
}

var Ld = Op{
	Name:  "LD",
	Opc:   LD,
	Class: "dat",
	OpNo:  2,
}

var Ldi = Op{
	Name:  "LDI",
	Opc:   LDI,
	Class: "dat",
	OpNo:  2,
}

var Mv = Op{
	Name:  "MV",
	Opc:   MV,
	Class: "dat",
	OpNo:  2,
}

var Wrt = Op{
	Name:  "WRT",
	Opc:   WRT,
	Class: "dat",
	OpNo:  2,
}

var Wrti = Op{
	Name:  "WRTI",
	Opc:   WRTI,
	Class: "dat",
	OpNo:  2,
}

var Jmp = Op{
	Name:  "JMP",
	Opc:   JMP,
	Class: "ctf",
	OpNo:  1,
}

var Beq = Op{
	Name:  "BEQ",
	Opc:   BEQ,
	Class: "ctf",
	OpNo:  3,
}

var Bz = Op{
	Name:  "BZ",
	Opc:   BZ,
	Class: "ctf",
	OpNo:  2,
}

var Hlt = Op{
	Name:  "HLT",
	Opc:   HLT,
	Class: "ctf",
	OpNo:  0,
}

var Instructions = []Op{
	Add, Addi, Sub, Subi, Mul, Div,
	And, Or, Xor, Cmp,
	Ld, Ldi, Mv, Wrt, Wrti,
	Jmp, Beq, Bz, Hlt,
}

func MatchName(name string) Op {
	for i := range Instructions {
		if name == Instructions[i].Name {
			return Instructions[i]
		}
	}

	return Hlt
}

func MatchOpc(opc int) Op {
	for i := range Instructions {
		if opc == Instructions[i].Opc {
			return Instructions[i]
		}
	}

	return Hlt
}
