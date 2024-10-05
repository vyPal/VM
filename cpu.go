package main

type CPU struct {
	Memory              *Memory
	Registers           [16]uint32
	PC                  uint32
	Stack               *Stack
	Halted              bool
	LastAccessedAddress uint32
}

type Stack struct {
	Stack [0x4000]uint32
	SP    uint32
}

func NewStack() *Stack {
	return &Stack{
		Stack: [0x4000]uint32{},
		SP:    0,
	}
}

func (s *Stack) Push(data uint32) {
	s.Stack[s.SP] = data
	s.SP++
}

func (s *Stack) Pop() uint32 {
	s.SP--
	return s.Stack[s.SP]
}

func NewCPU() *CPU {
	return &CPU{
		Memory:    NewMemory(),
		Registers: [16]uint32{},
		PC:        0,
		Stack:     NewStack(),
	}
}

func (c *CPU) Reset() {
	c.Memory.Clear()
	c.Registers = [16]uint32{}
	c.PC = 0
	c.Stack = NewStack()
	c.Halted = false
}

func (c *CPU) Step() {
	if c.Halted {
		return
	}
	instr := DecodeInstruction(c.Memory, &c.PC)
	instr.Execute(c, instr.Operands)
}

func (c *CPU) LoadProgram(program *Parser) {
	fisrtSectorWithInstructions := -1
	for sid, sector := range program.Sectors {
		if sector.Instructions != nil && fisrtSectorWithInstructions == -1 {
			fisrtSectorWithInstructions = sid
		}
		if sector.Program != nil {
			c.Memory.LoadProgram(sector.BaseAddress, sector.Program)
		}
	}
	if fisrtSectorWithInstructions == -1 {
		return
	}
	c.PC = program.Sectors[fisrtSectorWithInstructions].BaseAddress
}
