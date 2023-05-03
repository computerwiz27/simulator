package components

import (
	"encoding/binary"
	"strconv"
	"strings"
)

type MemChans struct {
	ex_memOk chan bool
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

func writeToCache(cache []CaAddr, uloc uint32, val int) []CaAddr {
	loc := int(uloc)
	for i := 0; i < len(cache); i++ {
		if cache[i].loc == loc {
			cache[i].val = val
			break
		}
	}

	return cache
}

func Mem(flg Flags, mem Memory, sysCa SysCache,
	buf Buffer, bus MemChans) {

	exData := <-buf.in
	tmpMem := <-mem
	tmpL1 := <-sysCa.l1
	tmpL2 := <-sysCa.l2
	tmpL3 := <-sysCa.l3

	store := false
	if exData[0] == 1 {
		store = true
	}

	loc := binary.BigEndian.Uint32(exData[1:5])

	uval := binary.BigEndian.Uint32(exData[10:14])
	val := int(int32(uval))

	if !store {
		mem <- tmpMem
		sysCa.l1 <- tmpL1
		sysCa.l2 <- tmpL2
		sysCa.l3 <- tmpL3

		bus.ex_memOk <- true

		wbData := exData[5:14]
		buf.out <- wbData

		flg.memChk <- true

		return
	}

	tmpMem = writeToMem(tmpMem, loc, val)

	tmpL1 = writeToCache(tmpL1, loc, val)
	tmpL2 = writeToCache(tmpL2, loc, val)
	tmpL3 = writeToCache(tmpL3, loc, val)

	mem <- tmpMem
	sysCa.l1 <- tmpL1
	sysCa.l2 <- tmpL2
	sysCa.l3 <- tmpL3

	bus.ex_memOk <- true

	wbData := exData[5:14]

	buf.out <- wbData

	flg.memChk <- true
}
