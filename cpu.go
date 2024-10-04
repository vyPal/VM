package main

type CPU struct {
  Memory *Memory
  Registers [16]uint32
  PC uint32
  Stack *Stack
  Halted bool
}

type Stack struct {
  Stack [0x4000]uint32
  SP uint32
}

func NewStack() *Stack {
  return &Stack{
    Stack: [0x4000]uint32{},
    SP: 0,
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
    Memory: NewMemory(),
    Registers: [16]uint32{},
    PC: 0,
    Stack: NewStack(),
  }
}

func (c *CPU) Step() {
  if c.Halted {
    return
  }
  instr := DecodeInstruction(c.Memory, &c.PC)
  instr.Execute(c, instr.Operands)
}

func (c *CPU) LoadProgram(program []byte) {
  c.Memory.LoadProgram(program)
}
  
