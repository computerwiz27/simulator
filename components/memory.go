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
	memChk  chan bool
}

type Buffer struct {
	in  chan []byte
	out chan []byte
}

type CaAddr struct {
	loc int
	val int
}

type Cache chan []CaAddr

type SysCache struct {
	l1 Cache
	l2 Cache
	l3 Cache
}

type Memory chan []byte
