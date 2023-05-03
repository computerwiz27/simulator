package components

// Resgister file
type Registers struct {
	Pc  chan uint
	Reg []chan int
}

// Buffer
type Buffer struct {
	In  chan []byte
	Out chan []byte
}

// Cache addresses are made of a location component that references the
// location in system memory, and a value component that holds the value
// at that location
type CaAddr struct {
	Loc int
	Val int
}

// Caches are made of cache addresses
type Cache chan []CaAddr

// System cache is made of three levels
type SysCache struct {
	L1 Cache
	L2 Cache
	L3 Cache
}

type Memory chan []byte
