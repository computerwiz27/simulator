package controller

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/computerwiz27/simulator/binary"
	"github.com/computerwiz27/simulator/op"
)

func errCheck(e error, line int) {
	if e != nil {
		fmt.Printf("Error on line %d: ", line+1)
		fmt.Println(e)
		os.Exit(2)
	}
}

func tokenToInt(token string) (bool, int, error) {
	imd := false
	val, strErr := strconv.Atoi(token)
	var err error

	if strErr == nil {
		imd = true
	} else {
		val, strErr = strconv.Atoi(token[1:])

		if val < 0 || val > 31 {
			err = errors.New("register not in scope")
		}

		if string(token[0]) != "r" || strErr != nil {
			err = errors.New("unrecognised token: '" + token + "'")
		}
	}

	return imd, val, err
}

func nonImdTokenChk(token string, line int) uint32 {
	imd, val, e := tokenToInt(token)
	errCheck(e, line)

	if imd {
		e = errors.New("token '" + token + "' can not be an immediate value")
		errCheck(e, line)
	}

	return uint32(val)
}

func imdTokenChk(token string, resBits int, line int) uint32 {
	imd, val, e := tokenToInt(token)
	errCheck(e, line)

	ret := uint32(0)

	if imd {
		ret = ret | (1 << 26) //set imd flag

		mask := int(math.Pow(2, (float64(resBits)+4))) - 1

		if val < 0 {
			uval := -val & mask
			val = uval + (1 << (resBits + 4))
		} else {
			val = val & mask
		}

		ret = ret | uint32(val)
	} else {
		ret = ret | (uint32(val) << resBits)
	}

	return ret
}

func numTokenChk(token string, size int, line int) uint32 {
	imd, val, e := tokenToInt(token)
	errCheck(e, line)

	if !imd {
		e = errors.New("token '" + token + "' must be a numeric value")
		errCheck(e, line)
	}

	mask := int(math.Pow(2, float64(size-1))) - 1

	if val < 0 {
		uval := -val & mask
		val = uval + (1 << (size - 1))
	} else {
		val = val & mask
	}

	return uint32(val)
}

func nonEmptyStrings(strings []string) []string {
	var nonEmptyStrings []string
	for i := 0; i < len(strings); i++ {
		if !(strings[i] == " " || strings[i] == "") {
			nonEmptyStrings = append(nonEmptyStrings, strings[i])
		}
	}

	return nonEmptyStrings
}

func removeComments(tokens []string, comDel string) []string {
	var uncomTokens []string
	for i := 0; i < len(tokens); i++ {
		match, _ := regexp.MatchString(comDel+"*", tokens[i])
		if match {
			break
		}

		uncomTokens = append(uncomTokens, tokens[i])
	}

	return uncomTokens
}

func Assemble(file []byte) []byte {
	var instructions []byte

	comDel := ";;"

	lines := strings.Split(string(file), "\n")

	for line := range lines {
		var ins uint32

		tokens := strings.Split(lines[line], " ")
		tokens = removeComments(tokens, comDel)
		tokens = nonEmptyStrings(tokens)

		if len(tokens) == 0 {
			continue
		}

		args := len(tokens) - 1
		opr, e := op.MatchName(tokens[0])
		errCheck(e, line)

		if args != opr.OpNo && opr.Class != "dat" {
			eStr := "improper use of " + opr.Name + "; instruction takes " +
				strconv.Itoa(opr.OpNo) + " arguments"
			e := errors.New(eStr)
			errCheck(e, line)
		}

		ins = uint32(opr.Opc) << 27

		switch opr.Class {
		case "ari", "log":

			if opr == op.Not {
				val := nonImdTokenChk(tokens[1], line)
				ins = ins | (val << 21)

				val = imdTokenChk(tokens[2], 16, line)
				ins = ins | val

				// Assembly for standard binary operations
			} else {
				for i := 1; i <= 2; i++ {
					val := nonImdTokenChk(tokens[i], line)
					ins = ins | (val << (21 - 5*(i-1)))
				}

				val := imdTokenChk(tokens[3], 11, line)
				ins = ins | val
			}

		case "dat":

			switch opr {
			case op.Ld:
				val := nonImdTokenChk(tokens[1], line)
				ins = ins | (val << 20)

				if args == 3 {
					ins = ins | (1 << 25) //set offset flag

					val = nonImdTokenChk(tokens[3], line)
					ins = ins | (val << 15)
				} else if args != 2 {
					eStr := "improper use of " + opr.Name + "; instruction takes" +
						strconv.Itoa(opr.OpNo) + " arguments"
					e := errors.New(eStr)
					errCheck(e, line)
				}

				val = imdTokenChk(tokens[2], 10, line)
				ins = ins | val

			case op.Wrt:
				val := nonImdTokenChk(tokens[2], line)
				ins = ins | (val << 20)

				if args == 3 {
					ins = ins | (1 << 25) //set offset flag

					val = nonImdTokenChk(tokens[3], line)
					ins = ins | (val << 15)
				} else if args != 2 {
					eStr := "improper use of " + opr.Name + "; instruction takes" +
						strconv.Itoa(opr.OpNo) + " arguments"
					e := errors.New(eStr)
					errCheck(e, line)
				}

				val = imdTokenChk(tokens[1], 10, line)
				ins = ins | val

			case op.Mv:
				if args != opr.OpNo && opr.Class != "dat" {
					eStr := "improper use of " + opr.Name + "; instruction takes" +
						strconv.Itoa(opr.OpNo) + " arguments"
					e := errors.New(eStr)
					errCheck(e, line)
				}

				val := nonImdTokenChk(tokens[1], line)
				ins = ins | (val << 21)

				val = imdTokenChk(tokens[2], 16, line)
				ins = ins | val
			}

		case "ctf":
			switch opr {
			case op.Jmp:
				val := numTokenChk(tokens[1], 27, line)
				ins = ins | val

			case op.Bz:
				val := nonImdTokenChk(tokens[1], line)
				ins = ins | (val << 22)

				val = numTokenChk(tokens[2], 22, line)
				ins = ins | val

			case op.Beq:
				val := nonImdTokenChk(tokens[1], line)
				ins = ins | (val << 21)

				val = numTokenChk(tokens[3], 9, line)
				ins = ins | (val << 12)

				val = imdTokenChk(tokens[2], 7, line)
				ins = ins | val

			case op.Hlt, op.Nop:

			}
		}

		instructions = binary.BigEndian.AppendUint32(instructions, ins)
	}

	return instructions
}
