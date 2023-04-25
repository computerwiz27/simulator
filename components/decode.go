package components

import (
	"encoding/binary"

	"github.com/computerwiz27/simulator/op"
)

func imdCheck(ins uint32) bool {
	ins = ins << 5
	ins = ins >> 31

	return ins == 1
}

func offSetCheck(ins uint32) bool {
	ins = ins << 6
	ins = ins >> 31

	return ins == 1
}

func decodeUnsigned(val uint32, start int, end int) int {
	val = val << start
	val = val >> (start + (31 - end))

	return int(val)
}

func decodeSigned(val uint32, start int, end int) int {
	uval := val << (start + 1)
	uval = uval >> (start + (31 - end) + 1)

	signBit := val << start
	signBit = signBit >> 31

	var sign int
	if signBit == 1 {
		sign = -1
	} else {
		sign = 1
	}

	return int(uval) * sign
}

// Decode the instruction
func Decode(regs Registers, flg Flags, mem Memory, prog Memory,
	fet_dec Buffer, dec_ex Buffer) {

	var fet_data []byte
	select {
	case fet_data = <-fet_dec:

	default:
		for i := 0; i < 4; i++ {
			fet_data = append(fet_data, 0)
		}
	}

	ins := binary.BigEndian.Uint32(fet_data)

	opc := ((0b11111 << 27) & ins) >> 27
	opr := op.MatchOpc(uint(opc))

	imd := false
	if opr.Imd {
		imd = imdCheck(ins)
	}

	offSet := false
	if opr.OffSet {
		offSet = offSetCheck(ins)
	}

	var opds []int
	for i := 0; i < 3; i++ {
		opds = append(opds, 0)
	}

	switch opr.Class {
	case "ctf":
		switch opr {
		case op.Nop, op.Hlt:

		case op.Jmp:
			opds[0] = decodeSigned(ins, 5, 31)

		case op.Beq:
			ra := decodeUnsigned(ins, 6, 10)
			opds[0] = <-regs.reg[ra]
			regs.reg[ra] <- opds[0]

			opds[1] = decodeSigned(ins, 11, 19)

			if imd {
				opds[2] = decodeUnsigned(ins, 20, 31)
			} else {
				rb := decodeUnsigned(ins, 20, 24)
				opds[2] = <-regs.reg[rb]
				regs.reg[rb] <- opds[2]
			}

		case op.Bz:
			ra := decodeUnsigned(ins, 5, 9)
			opds[0] = <-regs.reg[ra]
			regs.reg[ra] <- opds[0]

			opds[1] = decodeSigned(ins, 10, 31)
		}

	case "ari", "log":
		if opr == op.Not {
			opds[0] = decodeUnsigned(ins, 6, 10)

			if imd {
				opds[1] = decodeUnsigned(ins, 16, 31)
			} else {
				rs := decodeUnsigned(ins, 11, 15)
				opds[1] = <-regs.reg[rs]
				regs.reg[rs] <- opds[1]
			}
		} else {
			opds[0] = decodeUnsigned(ins, 6, 10)

			rsa := decodeUnsigned(ins, 11, 15)
			opds[1] = <-regs.reg[rsa]
			regs.reg[rsa] <- opds[1]

			if imd {
				opds[2] = decodeUnsigned(ins, 16, 31)
			} else {
				rsb := decodeUnsigned(ins, 16, 20)
				opds[2] = <-regs.reg[rsb]
				regs.reg[rsb] <- opds[2]
			}
		}

	case "dat":
		switch opr {
		case op.Ld:
			opds[0] = decodeUnsigned(ins, 7, 11)

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				opds[1] = <-regs.reg[rsb]
				regs.reg[rsb] <- opds[1]
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 31)
			} else {
				rsa := decodeUnsigned(ins, 17, 21)
				opds[2] = <-regs.reg[rsa]
				regs.reg[rsa] <- opds[2]
			}

		case op.Wrt:
			rsa := decodeUnsigned(ins, 7, 11)
			opds[0] = <-regs.reg[rsa]
			regs.reg[rsa] <- opds[0]

			if offSet {
				rsb := decodeUnsigned(ins, 12, 16)
				opds[1] = <-regs.reg[rsb]
				regs.reg[rsb] <- opds[1]
			}

			if imd {
				opds[2] = decodeUnsigned(ins, 17, 21)
			} else {
				rd := decodeUnsigned(ins, 17, 31)
				opds[2] = <-regs.reg[rd]
				regs.reg[rd] <- opds[2]
			}
		}
	}

	var ex_data []byte
	ex_data = append(ex_data, byte(opc))
	for i := 0; i < 3; i++ {
		ex_data = binary.BigEndian.AppendUint32(ex_data, uint32(opds[i]))
	}

	dec_ex <- ex_data

	//flg.decChk <- true
}
