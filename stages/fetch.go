package stages

import (
	"github.com/computerwiz27/simulator/binary"
	c "github.com/computerwiz27/simulator/components"
	"github.com/computerwiz27/simulator/op"
)

// Channels used to comunicate with the system
type FetChans struct {
	NIns chan []int //next instruction
	Bran chan int   //ast instruction was a branch
	//0 for no, 1 for jmp, and 2 for conditional branches
	BTaken    chan bool //last branch was taken
	Stall     chan bool
	UpdtCount chan bool
	ForkCount chan int
}

// Stage cache
type FetCache struct {
	Forks     chan []uint //array of fork locations
	LastCount chan uint
	LCyBranch chan bool //last cycle was a branch
	LCyStall  chan bool //last cycle was a stall
}

func fetchIns(prog []byte, counter uint) []byte {
	var ins []byte

	if 4*(counter) >= uint(len(prog)) {
		ins = make([]byte, 4)
	} else {
		for i := 0; i < 4; i++ {
			ins = append(ins, prog[4*int(counter)+i])
		}
	}

	return ins
}

// Fetch the next instruction from memory
func Fetch(regs c.Registers, flg c.Flags, prog c.Memory,
	buf c.Buffer, bus FetChans, cache FetCache) {

	counter := <-regs.Pc
	progMem := <-prog
	forks := <-cache.Forks
	lastCount := <-cache.LastCount
	lastCyclesBranch := <-cache.LCyBranch
	lastCycleStall := <-cache.LCyStall

	bTaken := <-bus.BTaken
	stall := <-bus.Stall
	forkCount := <-bus.ForkCount
	decIns := <-bus.NIns
	branch := <-bus.Bran
	updateCounter := <-bus.UpdtCount

	initialCounter := counter

	count1 := counter
	count2 := counter + 1

	//For these cases update the counter before fetching the instruction

	//if last cycle was a JMP instruction
	if branch > 0 {
		//increase counter to target instruction with compnsation for
		//last couple of cycles
		//counter += uint(decIns[0]) - 1
		//if the cycle before that was a branch instruction increment counter
		//to ajust for not incrementing the counter on that cycle
		if lastCyclesBranch {
			//counter++
		}

		if branch == 2 {
			forks = append(forks, uint(forkCount)+1)
		}
	}

	//assume the branch was taken, if nor return to the latest fork
	if !bTaken {
		counter = forks[len(forks)-1]
		forks = forks[:len(forks)-1]

		count1 = counter
		count2 = counter + 1
	}

	if updateCounter {
		count1 = counter + uint(decIns[0])
		count2 = counter + uint(decIns[1])
	}

	//if it is the first cycle after a stall, update the counter
	if !stall && lastCycleStall {
		counter += uint(decIns[0])
	}

	if stall && bTaken {
		counter = initialCounter //don't modify the counter
		count1 = counter
		count2 = counter + 1

		lastCycleStall = true //mark this cycle as a stall
	} else {
		lastCycleStall = false
	}

	//fetch instructions
	ins1 := fetchIns(progMem, count1)
	ins2 := fetchIns(progMem, count2)

	debugIns1 := 0b11111000 & ins1[0]
	debugIns1 = debugIns1 >> 3
	debugIns2 := 0b11111000 & ins2[0]
	debugIns2 = debugIns2 >> 3
	opr1 := op.MatchOpc(uint(debugIns1))
	opr2 := op.MatchOpc(uint(debugIns2))
	_, _ = opr1, opr2

	// If last cycle was a normal instruction increment the counter

	lastCount = counter

	counter += uint(decIns[1])
	if updateCounter {
		counter++
	}

	lastCyclesBranch = branch > 0

	decData := binary.BigEndian.AppendUint32(ins1, uint32(count1))
	decData = append(decData, ins2...)
	decData = binary.BigEndian.AppendUint32(decData, uint32(count2))

	//Pass the instruction to decode
	buf.Out <- decData

	//Return values to channels
	regs.Pc <- counter
	prog <- progMem

	cache.Forks <- forks
	cache.LastCount <- lastCount
	cache.LCyBranch <- lastCyclesBranch
	cache.LCyStall <- lastCycleStall

	//Raise done flag
	flg.FetChck <- true
}
