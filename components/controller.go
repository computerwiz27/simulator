package components

import "fmt"

func Run(mem []string, n int) {

	pc := make(chan uint32, 1)

	var regs []chan int32
	for i := 0; i < n; i++ {
		regs = append(regs, make(chan int32, 1))
	}

	registers := Registers{
		pc:  pc,
		reg: regs,
	}

	halt := make(chan bool)

	flags := Flags{
		halt: halt,
	}

	registers.pc <- 0

	go Fetch(registers, flags, mem)

	<-flags.halt

	for reg := range regs {
		select {
		case val, ok := <-registers.reg[reg]:
			if ok {
				fmt.Printf("reg%d: %d\n", reg, val)
			}
		default:
			fmt.Printf("reg%d: no value\n", reg)
		}
	}

	fmt.Println("done")
}
