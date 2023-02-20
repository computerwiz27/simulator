package components

type Registers struct {
	pc  chan uint
	reg []chan int
}

type Flags struct {
	halt chan bool
}

type Memory chan []byte

type Prog chan []int
