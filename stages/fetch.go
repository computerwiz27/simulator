package stages

import (
	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

type FetChans struct {
	NIns   chan int
	Bran   chan int
	BTaken chan bool
	Stall  chan bool
}

type FetCache struct {
	Forks    chan []uint
	BLast    chan bool
	Lcystall chan bool
}

// Fetch the next instruction from memory
func Fetch(regs c.Registers, flg c.Flags, prog c.Memory,
	buf c.Buffer, bus FetChans, cache FetCache) {

	counter := <-regs.Pc
	tmp := <-prog
	forks := <-cache.Forks
	lastCycleBranch := <-cache.BLast
	lastCycleStall := <-cache.Lcystall

	bTaken := <-bus.BTaken
	stall := <-bus.Stall
	decIns := <-bus.NIns
	branch := <-bus.Bran

	initialCounter := counter

	if branch == 1 {
		counter += uint(decIns) - 2
		if lastCycleBranch {
			counter++
		}
	}
	if branch == 2 {
		forks = append(forks, counter-1)
		counter += uint(decIns) - 1
		if lastCycleBranch {
			counter++
		}
	}
	if !bTaken {
		counter = forks[len(forks)-1] + 1
		forks = forks[:len(forks)-1]
	}
	if !stall && lastCycleStall {
		counter += uint(decIns)
	}

	var ins []byte
	if 4*counter >= uint(len(tmp)) {
		ins = make([]byte, 4)
	} else {
		for i := 0; i < 4; i++ {
			ins = append(ins, tmp[4*int(counter)+i])
		}
	}

	debugIns := 0b11111000 & ins[0]
	debugIns = debugIns >> 3
	opr := op.MatchOpc(uint(debugIns))
	_ = opr

	if branch == 0 {
		counter += uint(decIns)
		lastCycleBranch = false
	} else {
		lastCycleBranch = true
	}

	if stall {
		if !lastCycleStall {
			counter = initialCounter
		} else {
			counter = initialCounter
		}
		lastCycleStall = true
	} else {
		lastCycleStall = false
	}

	buf.Out <- ins

	regs.Pc <- counter
	prog <- tmp
	cache.Forks <- forks
	cache.BLast <- lastCycleBranch
	cache.Lcystall <- lastCycleStall

	flg.FetChck <- true
}
