package components

func Run(lines []string, n int) {

	pc := make(chan uint32)
	regs := make([]chan int32, n)

	registers := Registers{
		pc:  pc,
		reg: regs,
	}

	halt := make(chan bool)

	flags := Flags{
		halt: halt,
	}

	registers.pc <- 0
	flags.halt <- false

	Fetch(lines, registers, flags)
}
