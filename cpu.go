package main

type CPU struct {
	MemoryManager 		 *MemoryManager
	Registers           [16]uint32
	PC                  uint32
	Halted              bool
	LastAccessedAddress uint32
	FileSystem          VFS
	FileTable           map[uint32]interface{}
	NextFD              uint32
}

func NewCPU() *CPU {
	return &CPU{
		MemoryManager:    NewMemoryManager(NewMemory()),
		Registers: [16]uint32{},
		PC:        0,
		Halted:    false,
		FileTable: make(map[uint32]interface{}),
		NextFD:    0,
	}
}

func (c *CPU) Reset() {
	c.MemoryManager.Memory.Clear()
	c.Registers = [16]uint32{}
	c.PC = 0
	c.Halted = false
}

func (c *CPU) Step() {
	if c.Halted {
		return
	}
	instr := DecodeInstruction(c.MemoryManager, &c.PC)
	instr.Execute(c, instr.Operands)
}

func (c *CPU) LoadProgram(program *Bytecode) {
	for _, sector := range program.Sectors {
		if sector.Bytecode != nil {
			c.MemoryManager.Memory.LoadProgram(sector.StartAddress, sector.Bytecode)
		}
	}
	c.PC = program.StartAddress
}
