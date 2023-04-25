package components

type Registers struct {
	pc  chan uint
	reg []chan int
}

type Flags struct {
	halt    chan bool
	fetChck chan bool
	decChk  chan bool
	exChk   chan bool
	wbChk   chan bool
	wmChk   chan bool
}

type Buffer chan []byte

type Memory chan []byte
