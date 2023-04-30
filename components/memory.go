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

type ModRegCache chan []struct {
	reg int
	val int
}

type L1Cache chan []struct {
	memLoc int
	val    int
}

type Memory chan []byte
