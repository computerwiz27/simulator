package components

// Fetch the next instruction from memory
func Fetch(regs Registers, flg Flags, mem Memory, prog Prog) {
	counter := <-regs.pc
	tmp := <-prog

	var tokens [4]int
	for i := 0; i < 4; i++ {
		tokens[i] = tmp[int(counter)+i]
	}

	regs.pc <- counter
	prog <- tmp

	Decode(regs, flg, mem, prog, tokens)
}
