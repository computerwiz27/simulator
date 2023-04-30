package components

import (
	"encoding/binary"
	"strconv"
	"strings"
)

type MemChans struct {
	ex_stall chan bool
}

func WriteToMemory(flg Flags, mem Memory, buf Buffer, bus MemChans) {
	exData := <-buf.in

	store := false
	if exData[0] == 1 {
		store = true
	}

	loc := binary.BigEndian.Uint32(exData[1:5])

	uval := binary.BigEndian.Uint32(exData[10:14])
	val := int(uval)

	if store {
		lines := strings.Split(string(<-mem), "\n")

		if int(loc) >= len(lines) {
			for i := len(lines); i <= int(loc); i++ {
				lines = append(lines, "")
			}
		}

		lines[loc] = strconv.Itoa(int(val))

		//make a temporary variable and append the new memory bytes
		var tmp []byte
		for i := range lines {
			tmp = append(tmp, []byte(lines[i])...)
			tmp = append(tmp, []byte("\n")...) //add a new line after every line
			debugStr := string(tmp)
			_ = debugStr
		}

		mem <- tmp
	}

	wbData := exData[5:14]

	buf.out <- wbData

	flg.memChk <- true
}
