package main

import (
	"fmt"
)

type CPU struct {
	MemoryManager       *MemoryManager
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
		MemoryManager: NewMemoryManager(NewMemory()),
		Registers:     [16]uint32{},
		PC:            0,
		Halted:        false,
		FileTable:     make(map[uint32]interface{}),
		NextFD:        0,
	}
}

func (c *CPU) Reset() {
	c.MemoryManager = NewMemoryManager(NewMemory())
	c.Registers = [16]uint32{}
	c.PC = 0
	c.Halted = false
	for _, v := range c.FileTable {
		c.FileSystem.Close(v)
	}
	c.FileTable = make(map[uint32]interface{})
	c.NextFD = 0
}

func (c *CPU) Step() {
	if c.Halted {
		return
	}
	instr := DecodeInstruction(c.MemoryManager, &c.PC)
	instr.Execute(c, instr.Operands)
}

func (c *CPU) LoadProgram(program *Bytecode) {
	var p *ProgramInfo
	for _, sector := range program.Sectors {
		if sector.Bytecode != nil {
			if sector.StartAddress != 0x0 && sector.StartAddress < 0x80000000 {
				panic(fmt.Sprintf("Specifying a explicit start address for a sector to be stored in RAM is not allowed as it may interfere with dynamic memory allocation."))
			} else if sector.StartAddress == 0x0 {
				if p == nil {
					p = c.MemoryManager.NewProgram()
				}
				c.MemoryManager.AddSector(p, sector.StartAddress, sector.Bytecode, program.StartAddress == sector.StartAddress)
			} else {
				c.MemoryManager.Memory.LoadProgram(sector.StartAddress, sector.Bytecode)
			}
		}
	}

	c.PC = program.StartAddress

	if p != nil {
		start := c.MemoryManager.LoadProgram(p)
		if c.PC < 0x80000000 {
			c.PC = start
		}
	}
}
