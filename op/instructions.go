package op

import (
	"errors"
	"strings"
)

type Op struct {
	Name   string
	Opc    uint
	Class  string
	OpNo   int
	Stall  int
	Imd    bool
	OffSet bool
}

const (
	//Controll flow operations
	NOP uint = iota
	HLT
	JMP
	BEQ
	BZ

	//Arithmetic operations
	ADD
	SUB
	MUL
	DIV
	SL
	SR

	//Logical operations
	AND
	OR
	XOR
	NOT
	CMP

	//Data transfer operations
	LD
	WRT
	MV
)

// Controll Flow operations

// NOP
var Nop = Op{
	Name:   "NOP",
	Opc:    NOP,
	Class:  "ctf",
	OpNo:   0,
	Stall:  0,
	Imd:    false,
	OffSet: false,
}

// HLT
var Hlt = Op{
	Name:   "HLT",
	Opc:    HLT,
	Class:  "ctf",
	OpNo:   0,
	Stall:  0,
	Imd:    false,
	OffSet: false,
}

// JMP [inst no]
var Jmp = Op{
	Name:   "JMP",
	Opc:    JMP,
	Class:  "ctf",
	OpNo:   1,
	Stall:  0,
	Imd:    false,
	OffSet: false,
}

// BEQ regSA regSB/val [inst no]
var Beq = Op{
	Name:   "BEQ",
	Opc:    BEQ,
	Class:  "ctf",
	OpNo:   3,
	Stall:  1,
	Imd:    true,
	OffSet: false,
}

// BZ regS [inst no]
var Bz = Op{
	Name:   "BZ",
	Opc:    BZ,
	Class:  "ctf",
	OpNo:   2,
	Stall:  1,
	Imd:    false,
	OffSet: false,
}

//Arithmetic operations

// ADD regD regSA regSB
var Add = Op{
	Name:   "ADD",
	Opc:    ADD,
	Class:  "ari",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// SUB regD regSA regSB
var Sub = Op{
	Name:   "SUB",
	Opc:    SUB,
	Class:  "ari",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// MUL regD regSA regSB
var Mul = Op{
	Name:   "MUL",
	Opc:    MUL,
	Class:  "ari",
	OpNo:   3,
	Stall:  3,
	Imd:    true,
	OffSet: false,
}

// DIV regD regSA regSB
var Div = Op{
	Name:   "DIV",
	Opc:    DIV,
	Class:  "ari",
	OpNo:   3,
	Stall:  16,
	Imd:    true,
	OffSet: false,
}

var Sl = Op{
	Name:   "SL",
	Opc:    SL,
	Class:  "ari",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

var Sr = Op{
	Name:   "SR",
	Opc:    SR,
	Class:  "ari",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// Logical operations

// AND regD regSA regSB
var And = Op{
	Name:   "AND",
	Opc:    AND,
	Class:  "log",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// OR regD regSA regSB
var Or = Op{
	Name:   "OR",
	Opc:    OR,
	Class:  "log",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// XOR regD regSA regSB
var Xor = Op{
	Name:   "XOR",
	Opc:    XOR,
	Class:  "log",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// NOT regD regS
var Not = Op{
	Name:   "NOT",
	Opc:    NOT,
	Class:  "log",
	OpNo:   2,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// CMP regD regSA regSB
var Cmp = Op{
	Name:   "CMP",
	Opc:    CMP,
	Class:  "log",
	OpNo:   3,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// Data transfer operations

// LD regD regS
var Ld = Op{
	Name:   "LD",
	Opc:    LD,
	Class:  "dat",
	OpNo:   2,
	Stall:  0,
	Imd:    true,
	OffSet: true,
}

// WRT regD regS
var Wrt = Op{
	Name:   "WRT",
	Opc:    WRT,
	Class:  "dat",
	OpNo:   2,
	Stall:  0,
	Imd:    true,
	OffSet: true,
}

// MV regD regS
var Mv = Op{
	Name:   "MV",
	Opc:    MV,
	Class:  "dat",
	OpNo:   2,
	Stall:  0,
	Imd:    true,
	OffSet: false,
}

// List of all instructions
var instructions = []Op{
	Nop, Jmp, Beq, Bz, Hlt,
	Add, Sub, Mul, Div,
	And, Or, Xor, Not, Cmp,
	Ld, Mv, Wrt,
	Jmp, Beq, Bz, Hlt, Nop,
}

/* Match the name passed as a string to an instruction*/
// Defaults to Nop if name is not recognised.
func MatchName(name string) (Op, error) {
	for i := range instructions {
		if strings.ToUpper(name) == instructions[i].Name {
			return instructions[i], nil
		}
	}

	return Nop, errors.New("unrecognised symbol '" + name + "'")
}

// Match the op code passed as an integer to an instruction
// Defaults to Nop if opc is not recognised.
func MatchOpc(opc uint) Op {
	for i := range instructions {
		if opc == instructions[i].Opc {
			return instructions[i]
		}
	}

	return Nop
}
