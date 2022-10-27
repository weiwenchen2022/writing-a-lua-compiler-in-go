package vm

type Instruction uint32

const (
	SIZE_OP = 6
	SIZE_A = 8
	SIZE_C = 9
	SIZE_B = 9
	SIZE_Bx = SIZE_B + SIZE_C

	OFFSET_OP = 0
	OFFSET_A = SIZE_OP + OFFSET_OP
	OFFSET_C = SIZE_A + OFFSET_A
	OFFSET_B = SIZE_C + OFFSET_C
	OFFSET_Bx = OFFSET_C

	MAXARG_Bx = (1 << SIZE_Bx) - 1
	MAXARG_sBx = (MAXARG_Bx >> 1)
)

func (self Instruction) Opcode() int {
	return int(self & 0x3F)
}

func (self Instruction) ABC() (a, b, c int) {
	a = int((self >> OFFSET_A) & 0xFF)
	c = int((self >> OFFSET_C) & 0x1FF)
	b = int((self >> OFFSET_B) & 0x1FF)
	return
}

func (self Instruction) ABx() (a, bx int) {
	a = int((self >> OFFSET_A) & 0xFF)
	bx = int(self >> OFFSET_C)
	return
}

func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx
}

func (self Instruction) Ax() int {
	return int(self >> OFFSET_A)
}

func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}