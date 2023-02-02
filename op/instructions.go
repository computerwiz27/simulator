package op

type Ops byte

const (
	//Arithmetic operations
	ADD Ops = iota
	ADDI
	SUB
	SUBI
	MUL

	//Logical operations
	AND
	OR
	XOR
	LT
	EQ

	//Data transfer operations
	LD  //load to register
	WRT //write to memory
	MV  //copy from one register to another

	//Controll flow operations
	JMP //jump to instruction
	BZ  //branch if 0
	BEQ //branch if equals
	HLT //halt

)
