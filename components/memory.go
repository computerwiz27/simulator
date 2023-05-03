package components

type Registers struct {
	Pc  chan uint
	Reg []chan int
}

type Flags struct {
	Halt    chan bool
	FetChck chan bool
	DecChk  chan bool
	ExChk   chan bool
	WbChk   chan bool
	MemChk  chan bool
}

type Buffer struct {
	In  chan []byte
	Out chan []byte
}

type CaAddr struct {
	Loc int
	Val int
}

type Cache chan []CaAddr

type SysCache struct {
	L1 Cache
	L2 Cache
	L3 Cache
}

type Memory chan []byte
