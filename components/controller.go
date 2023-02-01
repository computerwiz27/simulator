package components

func Run(f []byte) {

	pc := make(chan uint32)
	reg0 := make(chan int32)
	reg1 := make(chan int32)
	reg2 := make(chan int32)
	reg3 := make(chan int32)

	registers := Registers{
		pc:   pc,
		reg0: reg0,
		reg1: reg1,
		reg2: reg2,
		reg3: reg3,
	}

	registers.pc <- 0
}
