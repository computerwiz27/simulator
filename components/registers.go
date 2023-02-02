package components

type Registers struct {
	pc  chan uint32
	reg []chan int32
}
