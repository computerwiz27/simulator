package components

type Registers struct {
	pc  chan uint32
	reg []chan int32
}

type Flags struct {
	halt chan bool
}

type Memory chan []byte
