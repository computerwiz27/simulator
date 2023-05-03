package components

import "github.com/computerwiz27/simulator/op"

type FetChans struct {
	nIns   chan int
	bran   chan int
	bTaken chan bool
	stall  chan bool
}

type FetCache struct {
	forks    chan []uint
	bLast    chan bool
	lcystall chan bool
}

// Fetch the next instruction from memory
func Fetch(regs Registers, flg Flags, prog Memory,
	buf Buffer, bus FetChans, cache FetCache) {

	counter := <-regs.pc
	tmp := <-prog
	forks := <-cache.forks
	lastCycleBranch := <-cache.bLast
	lastCycleStall := <-cache.lcystall

	bTaken := <-bus.bTaken
	stall := <-bus.stall
	decIns := <-bus.nIns
	branch := <-bus.bran

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

	buf.out <- ins

	regs.pc <- counter
	prog <- tmp
	cache.forks <- forks
	cache.bLast <- lastCycleBranch
	cache.lcystall <- lastCycleStall

	flg.fetChck <- true
}
