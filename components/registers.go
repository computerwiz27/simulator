package components

type Registers struct {
	pc   chan uint32
	reg0 chan int32
	reg1 chan int32
	reg2 chan int32
	reg3 chan int32
}
