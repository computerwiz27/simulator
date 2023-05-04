package stages

import (
	"encoding/binary"
	"strconv"
	"strings"

	c "github.com/computerwiz27/simulator/components"
)

type MemChans struct {
	Ex_memOk chan bool
}

func writeToMem(mem []byte, loc uint32, val int) []byte {
	lines := strings.Split(string(mem), "\n")

	if int(loc) >= len(lines) {
		for i := len(lines); i <= int(loc); i++ {
			lines = append(lines, "")
		}
	}

	lines[loc] = strconv.Itoa(int(val))

	//make a temporary variable and append the new memory bytes
	mem = make([]byte, 0)
	for i := range lines {
		mem = append(mem, []byte(lines[i])...)
		mem = append(mem, []byte("\n")...) //add a new line after every line
		debugStr := string(mem)
		_ = debugStr
	}

	return mem
}

func writeToCache(cache []c.CaAddr, uloc uint32, val int) []c.CaAddr {
	loc := int(uloc)
	for i := 0; i < len(cache); i++ {
		if cache[i].Loc == loc {
			cache[i].Val = val
			break
		}
	}

	return cache
}

func Mem(flg c.Flags, mem c.Memory, sysCa c.SysCache,
	buf c.Buffer, bus MemChans) {

	exData := <-buf.In
	tmpMem := <-mem
	tmpL1 := <-sysCa.L1
	tmpL2 := <-sysCa.L2
	tmpL3 := <-sysCa.L3

	store := false
	if exData[0] == 1 {
		store = true
	}

	loc := binary.BigEndian.Uint32(exData[1:5])

	uval := binary.BigEndian.Uint32(exData[5:9])
	val := int(int32(uval))

	if !store {
		mem <- tmpMem
		sysCa.L1 <- tmpL1
		sysCa.L2 <- tmpL2
		sysCa.L3 <- tmpL3

		bus.Ex_memOk <- true

		wbData := exData[9:28]
		buf.Out <- wbData

		flg.MemChk <- true

		return
	}

	tmpMem = writeToMem(tmpMem, loc, val)

	tmpL1 = writeToCache(tmpL1, loc, val)
	tmpL2 = writeToCache(tmpL2, loc, val)
	tmpL3 = writeToCache(tmpL3, loc, val)

	mem <- tmpMem
	sysCa.L1 <- tmpL1
	sysCa.L2 <- tmpL2
	sysCa.L3 <- tmpL3

	bus.Ex_memOk <- true

	wbData := exData[5:29]

	buf.Out <- wbData

	flg.MemChk <- true
}
