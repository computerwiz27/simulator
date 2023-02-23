package op

type Op struct {
	Name  string
	Opc   int
	Class string
	OpNo  int
}

const (
	//Arithmetic operations
	ADD int = iota
	ADDI
	SUB
	SUBI
	MUL
	DIV

	//Logical operations
	AND
	OR
	XOR
	CMP

	//Data transfer operations
	LD
	LDI
	MV
	MVI
	WRT
	WRTI

	//Controll flow operations
	JMP
	BEQ
	BZ
	HLT
)

//Arithmetic operations

// ADD regD regSA regSB
var Add = Op{
	Name:  "ADD",
	Opc:   ADD,
	Class: "ari",
	OpNo:  3,
}

// ADDI regD regS [val]
var Addi = Op{
	Name:  "ADDI",
	Opc:   ADDI,
	Class: "ari",
	OpNo:  3,
}

// SUB regD regSA regSB
var Sub = Op{
	Name:  "SUB",
	Opc:   SUB,
	Class: "ari",
	OpNo:  3,
}

// SUBI regD regS [val]
var Subi = Op{
	Name:  "SUBI",
	Opc:   SUBI,
	Class: "ari",
	OpNo:  3,
}

// MUL regD regSA regSB
var Mul = Op{
	Name:  "MUL",
	Opc:   MUL,
	Class: "ari",
	OpNo:  3,
}

// DIV regD regSA regSB
var Div = Op{
	Name:  "DIV",
	Opc:   DIV,
	Class: "ari",
	OpNo:  3,
}

// Logical operations

// AND regD regSA regSB
var And = Op{
	Name:  "AND",
	Opc:   AND,
	Class: "log",
	OpNo:  3,
}

// OR regD regSA regSB
var Or = Op{
	Name:  "OR",
	Opc:   OR,
	Class: "log",
	OpNo:  3,
}

// XOR regD regSA regSB
var Xor = Op{
	Name:  "XOR",
	Opc:   XOR,
	Class: "log",
	OpNo:  3,
}

// CMP regD regSA regSB
var Cmp = Op{
	Name:  "CMP",
	Opc:   CMP,
	Class: "log",
	OpNo:  3,
}

// Data transfer operations

// LD regD regS
var Ld = Op{
	Name:  "LD",
	Opc:   LD,
	Class: "dat",
	OpNo:  2,
}

// LDI regD [mem addr]
var Ldi = Op{
	Name:  "LDI",
	Opc:   LDI,
	Class: "dat",
	OpNo:  2,
}

// MV regD regS
var Mv = Op{
	Name:  "MV",
	Opc:   MV,
	Class: "dat",
	OpNo:  2,
}

// MVI regD val
var Mvi = Op{
	Name:  "MVI",
	Opc:   MVI,
	Class: "dat",
	OpNo:  2,
}

// WRT regD regS
var Wrt = Op{
	Name:  "WRT",
	Opc:   WRT,
	Class: "dat",
	OpNo:  2,
}

// WRTI [mem addr] regS
var Wrti = Op{
	Name:  "WRTI",
	Opc:   WRTI,
	Class: "dat",
	OpNo:  2,
}

// JMP [inst no]
var Jmp = Op{
	Name:  "JMP",
	Opc:   JMP,
	Class: "ctf",
	OpNo:  1,
}

// BEQ regSA regSB [inst no]
var Beq = Op{
	Name:  "BEQ",
	Opc:   BEQ,
	Class: "ctf",
	OpNo:  3,
}

// BZ regS [inst no]
var Bz = Op{
	Name:  "BZ",
	Opc:   BZ,
	Class: "ctf",
	OpNo:  2,
}

// HLT
var Hlt = Op{
	Name:  "HLT",
	Opc:   HLT,
	Class: "ctf",
	OpNo:  0,
}

// List of all instructions
var Instructions = []Op{
	Add, Addi, Sub, Subi, Mul, Div,
	And, Or, Xor, Cmp,
	Ld, Ldi, Mv, Mvi, Wrt, Wrti,
	Jmp, Beq, Bz, Hlt,
}

/* Match the name passed as a string to an instruction*/
// Defaults to Hlt if name is not recognised.
func MatchName(name string) Op {
	for i := range Instructions {
		if name == Instructions[i].Name {
			return Instructions[i]
		}
	}

	return Hlt
}

// Match the op code passed as an integer to an instruction
// Defaults to Hlt if opc is not recognised.
func MatchOpc(opc int) Op {
	for i := range Instructions {
		if opc == Instructions[i].Opc {
			return Instructions[i]
		}
	}

	return Hlt
}
