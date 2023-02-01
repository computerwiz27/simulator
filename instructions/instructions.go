package instructions

const (
	//Arithmetic operations
	ADD  = 0
	ADDI = 1
	SUB  = 2
	SUBI = 3
	MUL  = 4

	//Logical operations
	AND = 5
	OR  = 6
	XOR = 7
	LT  = 8
	EQ  = 9

	//Data transfer operations
	LD  = 10 //load to register
	WRT = 11 //write to memory
	MV  = 12 //copy from one register to another
)
