package stages

import (
	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

// Channels used to comunicate with the system
type FetChans struct {
	NIns chan int //next instruction
	Bran chan int //ast instruction was a branch
	//.           //0 for no, 1 for jmp, and 2 for conditional branches
	BTaken chan bool //last branch was taken
	Stall  chan bool
}

// Stage cache
type FetCache struct {
	Forks     chan []uint //array of fork locations
	LCyBranch chan bool   //last cycle was a branch
	LCyStall  chan bool   //last cycle was a stall
}

// Fetch the next instruction from memory
func Fetch(regs c.Registers, flg c.Flags, prog c.Memory,
	buf c.Buffer, bus FetChans, cache FetCache) {

	counter := <-regs.Pc
	progMem := <-prog
	forks := <-cache.Forks
	lastCycleBranch := <-cache.LCyBranch
	lastCycleStall := <-cache.LCyStall

	bTaken := <-bus.BTaken
	stall := <-bus.Stall
	decIns := <-bus.NIns
	branch := <-bus.Bran

	initialCounter := counter

	//For these cases update the counter before fetching the instruction

	//if last cycle was a JMP instruction
	if branch == 1 {
		//increase counter to target instruction with compnsation for
		//last couple of cycles
		counter += uint(decIns) - 2
		//if the cycle before that was a branch instruction increment counter
		//to ajust for not incrementing the counter on that cycle
		if lastCycleBranch {
			counter++
		}
	}

	//if last cycle was a conditional branch
	if branch == 2 {
		// append last intruction as a fork
		forks = append(forks, counter-1)
		//likewise increase counter with comensation
		counter += uint(decIns) - 1
		//likewise compensate if last cycle was a branch
		if lastCycleBranch {
			counter++
		}
	}

	//assume the branch was take, if nor return to the latest fork
	if !bTaken {
		counter = forks[len(forks)-1] + 1
		forks = forks[:len(forks)-1]
	}

	//if it is the first cycle after a stall, update the counter
	if !stall && lastCycleStall {
		counter += uint(decIns)
	}

	//fetch instruction
	var ins []byte
	if 4*counter >= uint(len(progMem)) {
		ins = make([]byte, 4)
	} else {
		for i := 0; i < 4; i++ {
			ins = append(ins, progMem[4*int(counter)+i])
		}
	}

	debugIns := 0b11111000 & ins[0]
	debugIns = debugIns >> 3
	opr := op.MatchOpc(uint(debugIns))
	_ = opr

	// If last cycle was a normal instruction increment the counter
	if branch == 0 {
		counter += uint(decIns)
		lastCycleBranch = false
	} else {
		lastCycleBranch = true
	}

	//if this a stall cycle
	if stall {
		counter = initialCounter //don't modify the counter

		lastCycleStall = true //mark this cycle as a stall
	} else {
		lastCycleStall = false
	}

	//Pass the instruction to decode
	buf.Out <- ins

	//Return values to channels
	regs.Pc <- counter
	prog <- progMem

	cache.Forks <- forks
	cache.LCyBranch <- lastCycleBranch
	cache.LCyStall <- lastCycleStall

	//Raise done flag
	flg.FetChck <- true
}
