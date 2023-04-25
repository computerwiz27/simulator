package components

// Fetch the next instruction from memory
func Fetch(regs Registers, flg Flags, mem Memory, prog Memory, fet_dec Buffer) {
	counter := <-regs.pc
	tmp := <-prog

	var ins []byte
	for i := 0; i < 4; i++ {
		ins = append(ins, tmp[int(counter)+i])
	}

	fet_dec <- ins

	regs.pc <- counter
	prog <- tmp

	//flg.fetChck <- true
}
