package components

// System flags
type Flags struct {
	Halt    chan bool
	FetChck chan bool
	DecChk  chan bool
	ExChk   chan bool
	WbChk   chan bool
	MemChk  chan bool
}
