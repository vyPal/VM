package main

import (
	"bytes"
	"encoding/binary"
)

// Oh how I love writing repetative code :D

type Instruction struct {
	Opcode   uint8
	Name     string
	Operands []Operand
	Execute  func(cpu *CPU, operands []Operand)
}

type OperandType int

const (
	Reg OperandType = iota
	DMem
	IMem
	Imm
)

type RegOperand struct {
	RegNum byte
	Size   byte // 0x0 = 32-bit, 0x1 = 16-bit, 0x2 = 8-bit
}

type MemType int

const (
	Address MemType = iota
	Register
	Offset
)

type DMemOperand struct {
	Type     MemType
	Addr     uint32
	Register byte
}

func (d *DMemOperand) ComputeAddress(cpu *CPU) uint32 {
	switch d.Type {
	case Address:
		return d.Addr
	case Register:
		return cpu.Registers[d.Register]
	case Offset:
		return cpu.Registers[d.Register] + d.Addr
	}
	return 0
}

type IMemOperand struct {
	Type     MemType
	Addr     uint32
	Offset   uint32
	Register byte
}

func (i *IMemOperand) ComputeAddress(cpu *CPU) uint32 {
	switch i.Type {
	case Address:
		return i.Addr
	case Register:
		return cpu.Registers[i.Register]
	case Offset:
		return cpu.Registers[i.Register] + i.Offset
	}
	return 0
}

type ImmOperand struct {
	Value uint32
}

type Operand struct {
	Type         OperandType
	AllowedTypes []OperandType
	Value        interface{}
}

func GetInstruction(inst string) *Instruction {
	for _, i := range instructionSet {
		if i.Name == inst {
			c := *i
			c.Operands = make([]Operand, len(i.Operands))
			copy(c.Operands, i.Operands)
			return &c
		}
	}
	return nil
}

func GetInstructionByOpcode(opcode byte) *Instruction {
	for _, i := range instructionSet {
		if i.Opcode == opcode {
			c := *i
			c.Operands = make([]Operand, len(i.Operands))
			copy(c.Operands, i.Operands)
			return &c
		}
	}
	return nil
}

var instructionSet = map[uint8]*Instruction{
	0x00: {
		Opcode: 0x00,
		Name:   "NOP",
		Execute: func(cpu *CPU, operands []Operand) {
			return
		},
	},
	0x01: {
		Opcode: 0x01,
		Name:   "LD",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] = cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] = uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] = uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] = uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] = uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] = cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] = uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] = uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] = operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] = uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] = uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x02: {
		Opcode: 0x02,
		Name:   "ST",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				cpu.LastAccessedAddress = operands[0].Value.(*DMemOperand).ComputeAddress(cpu)
				r := operands[1].Value.(*RegOperand)
				if r.Size == 0x0 {
					cpu.MemoryManager.WriteMemoryDWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu), cpu.Registers[r.RegNum])
				} else if r.Size == 0x1 {
					cpu.MemoryManager.WriteMemoryWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu), uint16(cpu.Registers[r.RegNum]))
				} else if r.Size == 0x2 {
					cpu.MemoryManager.WriteMemory(operands[0].Value.(*DMemOperand).ComputeAddress(cpu), uint8(cpu.Registers[r.RegNum]))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))
				r := operands[1].Value.(*RegOperand)
				if r.Size == 0x0 {
					cpu.MemoryManager.WriteMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)), cpu.Registers[r.RegNum])
				} else if r.Size == 0x1 {
					cpu.MemoryManager.WriteMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)), uint16(cpu.Registers[r.RegNum]))
				} else if r.Size == 0x2 {
					cpu.MemoryManager.WriteMemory(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)), uint8(cpu.Registers[r.RegNum]))
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem}}, // A - Dest
			{Type: Reg}, // B - Source
		},
	},
	0x03: {
		Opcode: 0x03,
		Name:   "ADD",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] += cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] += uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] += uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] += cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] += uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] += uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] += cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] += uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] += uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] += operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] += uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] += uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x04: {
		Opcode: 0x04,
		Name:   "SUB",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] -= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] -= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] -= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] -= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] -= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] -= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] -= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] -= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] -= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] -= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] -= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] -= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x05: {
		Opcode: 0x05,
		Name:   "MUL",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] *= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] *= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] *= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] *= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] *= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] *= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] *= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] *= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] *= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] *= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] *= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] *= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x06: {
		Opcode: 0x06,
		Name:   "DIV",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] /= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] /= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] /= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] /= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] /= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] /= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] /= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] /= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] /= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] /= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] /= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] /= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x07: {
		Opcode: 0x07,
		Name:   "MOD",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] %= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] %= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] %= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] %= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] %= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] %= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] %= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] %= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] %= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] %= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] %= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] %= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x08: {
		Opcode: 0x08,
		Name:   "AND",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] &= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] &= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] &= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] &= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] &= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] &= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] &= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] &= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] &= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] &= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] &= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] &= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x09: {
		Opcode: 0x09,
		Name:   "OR",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] |= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] |= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] |= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] |= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] |= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] |= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] |= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] |= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] |= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] |= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] |= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] |= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x0A: {
		Opcode: 0x0A,
		Name:   "XOR",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] ^= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] ^= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] ^= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] ^= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] ^= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] ^= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] ^= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] ^= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] ^= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] ^= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] ^= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] ^= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x0B: {
		Opcode: 0x0B,
		Name:   "NOT",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			if r.Size == 0x0 {
				cpu.Registers[r.RegNum] = ^cpu.Registers[r.RegNum]
			} else if r.Size == 0x1 {
				cpu.Registers[r.RegNum] = ^uint32(uint16(cpu.Registers[r.RegNum]))
			} else if r.Size == 0x2 {
				cpu.Registers[r.RegNum] = ^uint32(uint8(cpu.Registers[r.RegNum]))
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
		},
	},
	0x0C: {
		Opcode: 0x0C,
		Name:   "SHL",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] <<= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] <<= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] <<= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] <<= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] <<= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] <<= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] <<= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] <<= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] <<= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] <<= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] <<= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] <<= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x0D: {
		Opcode: 0x0D,
		Name:   "SHR",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] >>= cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] >>= uint32(uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] >>= uint32(uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]))
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] >>= cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] >>= uint32(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] >>= uint32(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] >>= cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] >>= uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] >>= uint32(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
				}
			case Imm:
				if r.Size == 0x0 {
					cpu.Registers[r.RegNum] >>= operands[1].Value.(*ImmOperand).Value
				} else if r.Size == 0x1 {
					cpu.Registers[r.RegNum] >>= uint32(uint16(operands[1].Value.(*ImmOperand).Value))
				} else if r.Size == 0x2 {
					cpu.Registers[r.RegNum] >>= uint32(uint8(operands[1].Value.(*ImmOperand).Value))
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x0E: {
		Opcode: 0x0E,
		Name:   "CMP",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			switch operands[1].Type {
			case Reg:
				if r.Size == 0x0 {
					if cpu.Registers[r.RegNum] == cpu.Registers[operands[1].Value.(*RegOperand).RegNum] {
						cpu.Registers[0xF] = 0x0
					} else if cpu.Registers[r.RegNum] > cpu.Registers[operands[1].Value.(*RegOperand).RegNum] {
						cpu.Registers[0xF] = 0x1
					} else if cpu.Registers[r.RegNum] < cpu.Registers[operands[1].Value.(*RegOperand).RegNum] {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x1 {
					if uint16(cpu.Registers[r.RegNum]) == uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x0
					} else if uint16(cpu.Registers[r.RegNum]) > uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x1
					} else if uint16(cpu.Registers[r.RegNum]) < uint16(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x2 {
					if uint8(cpu.Registers[r.RegNum]) == uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x0
					} else if uint8(cpu.Registers[r.RegNum]) > uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x1
					} else if uint8(cpu.Registers[r.RegNum]) < uint8(cpu.Registers[operands[1].Value.(*RegOperand).RegNum]) {
						cpu.Registers[0xF] = 0x2
					}
				}
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				if r.Size == 0x0 {
					if cpu.Registers[r.RegNum] == cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)) {
						cpu.Registers[0xF] = 0x0
					} else if cpu.Registers[r.RegNum] > uint32(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x1
					} else if cpu.Registers[r.RegNum] < uint32(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x1 {
					if uint16(cpu.Registers[r.RegNum]) == uint16(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x0
					} else if uint16(cpu.Registers[r.RegNum]) > uint16(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x1
					} else if uint16(cpu.Registers[r.RegNum]) < uint16(cpu.MemoryManager.ReadMemoryWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x2 {
					if uint8(cpu.Registers[r.RegNum]) == uint8(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x0
					} else if uint8(cpu.Registers[r.RegNum]) > uint8(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x1
					} else if uint8(cpu.Registers[r.RegNum]) < uint8(cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x2
					}
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				if r.Size == 0x0 {
					if cpu.Registers[r.RegNum] == cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))) {
						cpu.Registers[0xF] = 0x0
					} else if cpu.Registers[r.RegNum] > uint32(cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x1
					} else if cpu.Registers[r.RegNum] < uint32(cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x1 {
					if uint16(cpu.Registers[r.RegNum]) == uint16(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x0
					} else if uint16(cpu.Registers[r.RegNum]) > uint16(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x1
					} else if uint16(cpu.Registers[r.RegNum]) < uint16(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x2 {
					if uint8(cpu.Registers[r.RegNum]) == uint8(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x0
					} else if uint8(cpu.Registers[r.RegNum]) > uint8(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x1
					} else if uint8(cpu.Registers[r.RegNum]) < uint8(cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))) {
						cpu.Registers[0xF] = 0x2
					}
				}
			case Imm:
				if r.Size == 0x0 {
					if cpu.Registers[r.RegNum] == operands[1].Value.(*ImmOperand).Value {
						cpu.Registers[0xF] = 0x0
					} else if cpu.Registers[r.RegNum] > uint32(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x1
					} else if cpu.Registers[r.RegNum] < uint32(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x1 {
					if uint16(cpu.Registers[r.RegNum]) == uint16(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x0
					} else if uint16(cpu.Registers[r.RegNum]) > uint16(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x1
					} else if uint16(cpu.Registers[r.RegNum]) < uint16(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x2
					}
				} else if r.Size == 0x2 {
					if uint8(cpu.Registers[r.RegNum]) == uint8(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x0
					} else if uint8(cpu.Registers[r.RegNum]) > uint8(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x1
					} else if uint8(cpu.Registers[r.RegNum]) < uint8(operands[1].Value.(*ImmOperand).Value) {
						cpu.Registers[0xF] = 0x2
					}
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Source
		},
	},
	0x0F: {
		Opcode: 0x0F,
		Name:   "JMP",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				cpu.Registers[0xF] = cpu.PC
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.Registers[0xF] = cpu.PC
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				cpu.Registers[0xF] = cpu.PC
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x10: {
		Opcode: 0x10,
		Name:   "JEQ",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] == 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] == 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] == 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x11: {
		Opcode: 0x11,
		Name:   "JNE",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] != 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] != 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] != 0x0 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x12: {
		Opcode: 0x12,
		Name:   "JGT",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x13: {
		Opcode: 0x13,
		Name:   "JLT",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x14: {
		Opcode: 0x14,
		Name:   "JGE",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x15: {
		Opcode: 0x15,
		Name:   "JLE",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
				}
			case IMem:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
				}
			case Imm:
				if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
					cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
				}
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x16: {
		Opcode: 0x16,
		Name:   "CALL",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case DMem:
				cpu.MemoryManager.Push(cpu.PC)
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.MemoryManager.Push(cpu.PC)
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				cpu.MemoryManager.Push(cpu.PC)
				cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, operands[0].Value.(*ImmOperand).Value)
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
		},
	},
	0x17: {
		Opcode: 0x17,
		Name:   "RET",
		Execute: func(cpu *CPU, operands []Operand) {
			cpu.PC = cpu.MemoryManager.ExecuteJump(cpu.PC, cpu.MemoryManager.Pop())
		},
	},
	0x18: {
		Opcode: 0x18,
		Name:   "PUSH",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case Reg:
				cpu.MemoryManager.Push(cpu.Registers[operands[0].Value.(*RegOperand).RegNum])
			case DMem:
				cpu.LastAccessedAddress = operands[0].Value.(*DMemOperand).ComputeAddress(cpu)
				cpu.MemoryManager.Push(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu)))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))
				cpu.MemoryManager.Push(uint32(cpu.MemoryManager.ReadMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))))
			case Imm:
				cpu.MemoryManager.Push(operands[0].Value.(*ImmOperand).Value)
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // A - Source
		},
	},
	0x19: {
		Opcode: 0x19,
		Name:   "POP",
		Execute: func(cpu *CPU, operands []Operand) {
			switch operands[0].Type {
			case Reg:
				cpu.Registers[operands[0].Value.(*RegOperand).RegNum] = cpu.MemoryManager.Pop()
			case DMem:
				cpu.LastAccessedAddress = operands[0].Value.(*DMemOperand).ComputeAddress(cpu)
				cpu.MemoryManager.WriteMemoryDWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu), cpu.MemoryManager.Pop())
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))
				cpu.MemoryManager.WriteMemoryWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)), uint16(cpu.MemoryManager.Pop()))
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{Reg, DMem, IMem}}, // A - Dest
		},
	},
	0x1A: {
		Opcode: 0x1A,
		Name:   "HLT",
		Execute: func(cpu *CPU, operands []Operand) {
			cpu.Halted = true
		},
	},
	0x1B: {
		Opcode: 0x1B,
		Name:   "INC",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			if r.Size == 0x0 {
				cpu.Registers[r.RegNum]++
			} else if r.Size == 0x1 {
				cpu.Registers[r.RegNum] = uint32(uint16(cpu.Registers[r.RegNum]) + 1)
			} else if r.Size == 0x2 {
				cpu.Registers[r.RegNum] = uint32(uint8(cpu.Registers[r.RegNum]) + 1)
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
		},
	},
	0x1C: {
		Opcode: 0x1C,
		Name:   "DEC",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			if r.Size == 0x0 {
				cpu.Registers[r.RegNum]--
			} else if r.Size == 0x1 {
				cpu.Registers[r.RegNum] = uint32(uint16(cpu.Registers[r.RegNum]) - 1)
			} else if r.Size == 0x2 {
				cpu.Registers[r.RegNum] = uint32(uint8(cpu.Registers[r.RegNum]) - 1)
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
		},
	},
	0x1D: {
		Opcode: 0x1D,
		Name:   "OPEN",
		Execute: func(cpu *CPU, operands []Operand) {
			r := operands[0].Value.(*RegOperand)
			fd := cpu.NextFD
			var err error
			switch operands[1].Type {
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				cpu.FileTable[fd], err = cpu.FileSystem.Open(cpu.MemoryManager.ReadMemoryString(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				cpu.FileTable[fd], err = cpu.FileSystem.Open(cpu.MemoryManager.ReadMemoryString(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
			}
			if err != nil {
				cpu.Registers[r.RegNum] = 0xFFFFFFFF
			} else {
				cpu.Registers[r.RegNum] = uint32(fd)
				cpu.NextFD++
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - Dest
			{AllowedTypes: []OperandType{DMem, IMem}}, // B - Filename
		},
	},
	0x1E: {
		Opcode: 0x1E,
		Name:   "READ",
		Execute: func(cpu *CPU, operands []Operand) {
			fd := cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			var length uint32
			switch operands[2].Type {
			case Reg:
				length = cpu.Registers[operands[2].Value.(*RegOperand).RegNum]
			case DMem:
				cpu.LastAccessedAddress = operands[2].Value.(*DMemOperand).ComputeAddress(cpu)
				length = cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*IMemOperand).ComputeAddress(cpu))
				length = cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				length = operands[2].Value.(*ImmOperand).Value
			}

			if length == 0 {
				cpu.Registers[0xF] = 0xFFFFFFFF
				return
			}

			if operands[1].Type == Reg && length > 1 {
				cpu.Registers[0xF] = 0xFFFFFFFF
				return
			}

			data := make([]byte, length)
			n, err := cpu.FileSystem.Read(cpu.FileTable[fd], data)
			if err != nil {
				cpu.Registers[0xF] = 0xFFFFFFFF
			} else {
				cpu.Registers[0xF] = uint32(n)
				switch operands[1].Type {
				case Reg:
					cpu.Registers[operands[1].Value.(*RegOperand).RegNum] = uint32(n)
				case DMem:
					cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
					for i, b := range data {
						cpu.MemoryManager.WriteMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)+uint32(i), b)
					}
				case IMem:
					cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
					for i, b := range data {
						cpu.MemoryManager.WriteMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))+uint32(i), b)
					}
				}
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - FD
			{AllowedTypes: []OperandType{Reg, DMem, IMem}},      // B - Dest
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // C - Length
		},
	},
	0x1F: {
		Opcode: 0x1F,
		Name:   "WRITE",
		Execute: func(cpu *CPU, operands []Operand) {
			fd := cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			var length uint32
			switch operands[2].Type {
			case Reg:
				length = cpu.Registers[operands[2].Value.(*RegOperand).RegNum]
			case DMem:
				cpu.LastAccessedAddress = operands[2].Value.(*DMemOperand).ComputeAddress(cpu)
				length = cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*IMemOperand).ComputeAddress(cpu))
				length = cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[2].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				length = operands[2].Value.(*ImmOperand).Value
			}

			if length == 0 {
				cpu.Registers[0xF] = 0xFFFFFFFF
				return
			}

			data := make([]byte, length)
			switch operands[1].Type {
			case Reg:
				if length > 1 {
					cpu.Registers[0xF] = 0xFFFFFFFF
					return
				}
				data[0] = byte(cpu.Registers[operands[1].Value.(*RegOperand).RegNum])
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				for i := range data {
					data[i] = cpu.MemoryManager.ReadMemory(operands[1].Value.(*DMemOperand).ComputeAddress(cpu) + uint32(i))
				}
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				for i := range data {
					data[i] = cpu.MemoryManager.ReadMemory(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)) + uint32(i))
				}
			}

			n, err := cpu.FileSystem.Write(cpu.FileTable[fd], data)
			if err != nil {
				cpu.Registers[0xF] = 0xFFFFFFFF
			} else {
				cpu.Registers[0xF] = uint32(n)
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - FD
			{AllowedTypes: []OperandType{Reg, DMem, IMem}},      // B - Source
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // C - Length
		},
	},
	0x20: {
		Opcode: 0x20,
		Name:   "SEEK",
		Execute: func(cpu *CPU, operands []Operand) {
			fd := cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			var offset int64
			switch operands[1].Type {
			case Reg:
				offset = int64(cpu.Registers[operands[1].Value.(*RegOperand).RegNum])
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				offset = int64(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu)))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				offset = int64(cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))))
			case Imm:
				offset = int64(operands[1].Value.(*ImmOperand).Value)
			}
			whence := int(operands[2].Value.(*ImmOperand).Value)
			_, err := cpu.FileSystem.Seek(cpu.FileTable[fd], offset, whence)
			if err != nil {
				cpu.Registers[0xF] = 0xFFFFFFFF
			} else {
				cpu.Registers[0xF] = 0x0
			}
		},
		Operands: []Operand{
			{Type: Reg}, // A - FD
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Offset
			{Type: Imm}, // C - Whence
		},
	},
	0x21: {
		Opcode: 0x21,
		Name:   "LOADBIN",
		Execute: func(cpu *CPU, operands []Operand) {
			fd := cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			cpu.Registers[operands[1].Value.(*RegOperand).RegNum] = cpu.FileSystem.LoadBinary(cpu.FileTable[fd], cpu.MemoryManager)
		},
		Operands: []Operand{
			{Type: Reg}, // A - FD
			{Type: Reg}, // B - Dest
		},
	},
	0x22: {
		Opcode: 0x22,
		Name:   "CLOSE",
		Execute: func(cpu *CPU, operands []Operand) {
			fd := cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			cpu.FileSystem.Close(cpu.FileTable[fd])
			cpu.FileTable[fd] = nil
		},
		Operands: []Operand{
			{Type: Reg}, // A - FD
		},
	},
	0x23: {
		Opcode: 0x23,
		Name:   "MALLOC",
		Execute: func(cpu *CPU, operands []Operand) {
			var size uint32
			switch operands[0].Type {
			case Reg:
				size = cpu.Registers[operands[0].Value.(*RegOperand).RegNum]
			case DMem:
				cpu.LastAccessedAddress = operands[0].Value.(*DMemOperand).ComputeAddress(cpu)
				size = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))
				size = cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				size = operands[0].Value.(*ImmOperand).Value
			}
			addr, err := cpu.MemoryManager.Malloc(size)
			if err != nil {
				cpu.Registers[operands[1].Value.(*RegOperand).RegNum] = 0xFFFFFFFF
			}
			cpu.Registers[operands[1].Value.(*RegOperand).RegNum] = addr
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // A - Size
			{Type: Reg}, // B - Dest
		},
	},
	0x24: {
		Opcode: 0x24,
		Name:   "FREE",
		Execute: func(cpu *CPU, operands []Operand) {
			var size uint32
			switch operands[1].Type {
			case Reg:
				size = cpu.Registers[operands[1].Value.(*RegOperand).RegNum]
			case DMem:
				cpu.LastAccessedAddress = operands[1].Value.(*DMemOperand).ComputeAddress(cpu)
				size = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*DMemOperand).ComputeAddress(cpu))
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu))
				size = cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[1].Value.(*IMemOperand).ComputeAddress(cpu)))
			case Imm:
				size = operands[1].Value.(*ImmOperand).Value
			}
			switch operands[0].Type {
			case Reg:
				cpu.MemoryManager.Free(cpu.Registers[operands[0].Value.(*RegOperand).RegNum], size)
			case DMem:
				cpu.LastAccessedAddress = operands[0].Value.(*DMemOperand).ComputeAddress(cpu)
				cpu.MemoryManager.Free(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*DMemOperand).ComputeAddress(cpu)), size)
			case IMem:
				cpu.LastAccessedAddress = cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))
				cpu.MemoryManager.Free(cpu.MemoryManager.ReadMemoryDWord(cpu.MemoryManager.ReadMemoryDWord(operands[0].Value.(*IMemOperand).ComputeAddress(cpu))), size)
			}
		},
		Operands: []Operand{
			{AllowedTypes: []OperandType{Reg, DMem, IMem}},      // A - Start Address
			{AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // B - Size
		},
	},
}

func EncodeInstruction(inst *Instruction) []byte {
	var buf bytes.Buffer
	buf.WriteByte(inst.Opcode)
	for _, operand := range inst.Operands {
		if len(operand.AllowedTypes) == 0 {
			switch operand.Type {
			case Reg:
				buf.WriteByte(byte(operand.Value.(*RegOperand).RegNum) | byte(operand.Value.(*RegOperand).Size)<<4)
			case DMem:
				buf.WriteByte(byte(operand.Value.(*DMemOperand).Type))
				if operand.Value.(*DMemOperand).Type == Address {
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
				} else if operand.Value.(*DMemOperand).Type == Register {
					buf.WriteByte(byte(operand.Value.(*DMemOperand).Register))
				} else {
					buf.WriteByte(byte(operand.Value.(*DMemOperand).Register))
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
				}
			case IMem:
				buf.WriteByte(byte(operand.Value.(*IMemOperand).Type))
				if operand.Value.(*IMemOperand).Type == Address {
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
				} else if operand.Value.(*IMemOperand).Type == Register {
					buf.WriteByte(byte(operand.Value.(*IMemOperand).Register))
				} else {
					buf.WriteByte(byte(operand.Value.(*IMemOperand).Register))
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
				}
			case Imm:
				binary.Write(&buf, binary.LittleEndian, operand.Value.(*ImmOperand).Value)
			}
		} else {
			switch operand.Type {
			case Reg:
				buf.WriteByte(byte(Reg))
				buf.WriteByte(byte(operand.Value.(*RegOperand).RegNum) | byte(operand.Value.(*RegOperand).Size)<<4)
			case DMem:
				buf.WriteByte(byte(DMem))
				buf.WriteByte(byte(operand.Value.(*DMemOperand).Type))
				if operand.Value.(*DMemOperand).Type == Address {
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
				} else if operand.Value.(*DMemOperand).Type == Register {
					buf.WriteByte(byte(operand.Value.(*DMemOperand).Register))
				} else {
					buf.WriteByte(byte(operand.Value.(*DMemOperand).Register))
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
				}
			case IMem:
				buf.WriteByte(byte(IMem))
				buf.WriteByte(byte(operand.Value.(*IMemOperand).Type))
				if operand.Value.(*IMemOperand).Type == Address {
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
				} else if operand.Value.(*IMemOperand).Type == Register {
					buf.WriteByte(byte(operand.Value.(*IMemOperand).Register))
				} else {
					buf.WriteByte(byte(operand.Value.(*IMemOperand).Register))
					binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
				}
			case Imm:
				buf.WriteByte(byte(Imm))
				binary.Write(&buf, binary.LittleEndian, operand.Value.(*ImmOperand).Value)
			}
		}
	}
	return buf.Bytes()
}

func DecodeInstruction(mem *MemoryManager, pc *uint32) *Instruction {
	data := []byte{mem.ReadMemory(*pc)}
	inst := GetInstructionByOpcode(data[0])
	operands := make([]Operand, len(inst.Operands))
	offset := 1
	for i, operand := range inst.Operands {
		if len(operand.AllowedTypes) == 0 {
			switch operand.Type {
			case Reg:
				data = append(data, mem.ReadMemory(*pc+uint32(offset)))
				operands[i] = Operand{Type: Reg, Value: &RegOperand{RegNum: data[offset] & 0xF, Size: data[offset] >> 4}}
				offset++
			case DMem:
				data = append(data, mem.ReadMemory(*pc+uint32(offset)))
				switch data[offset] {
				case byte(Address):
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+1), 4)...)
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Address, Addr: binary.LittleEndian.Uint32(data[offset+1 : offset+5])}}
					offset += 5
				case byte(Register):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Register, Register: data[offset+1]}}
					offset += 2
				case byte(Offset):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+2), 4)...)
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Offset, Register: data[offset+1], Addr: binary.LittleEndian.Uint32(data[offset+2 : offset+6])}}
					offset += 6
				}
			case IMem:
				data = append(data, mem.ReadMemory(*pc+uint32(offset)))
				switch data[offset] {
				case byte(Address):
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+1), 4)...)
					operands[i] = Operand{Type: IMem, Value: &IMemOperand{Type: Address, Addr: binary.LittleEndian.Uint32(data[offset+1 : offset+5])}}
					offset += 5
				case byte(Register):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
					operands[i] = Operand{Type: IMem, Value: &IMemOperand{Type: Register, Register: data[offset+1]}}
					offset += 2
				case byte(Offset):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+2), 4)...)
					operands[i] = Operand{Type: IMem, Value: &IMemOperand{Type: Offset, Register: data[offset+1], Addr: binary.LittleEndian.Uint32(data[offset+2 : offset+6])}}
					offset += 6
				}
			case Imm:
				data = append(data, mem.ReadMemoryN(*pc+uint32(offset), 4)...)
				operands[i] = Operand{Type: Imm, Value: &ImmOperand{Value: binary.LittleEndian.Uint32(data[offset : offset+4])}}
				offset += 4
			}
		} else {
			data = append(data, mem.ReadMemory(*pc+uint32(offset)))
			switch data[offset] {
			case byte(Reg):
				data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
				operands[i] = Operand{Type: Reg, Value: &RegOperand{RegNum: data[offset+1] & 0xF, Size: data[offset+1] >> 4}}
				offset += 2
			case byte(DMem):
				data = append(data, mem.ReadMemory(*pc+uint32(offset+1)))
				switch data[offset+1] {
				case byte(Address):
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+2), 4)...)
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Address, Addr: binary.LittleEndian.Uint32(data[offset+2 : offset+6])}}
					offset += 6
				case byte(Register):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+2)))
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Register, Register: data[offset+2]}}
					offset += 3
				case byte(Offset):
					data = append(data, mem.ReadMemory(*pc+uint32(offset+2)))
					data = append(data, mem.ReadMemoryN(*pc+uint32(offset+3), 4)...)
					operands[i] = Operand{Type: DMem, Value: &DMemOperand{Type: Offset, Register: data[offset+2], Addr: binary.LittleEndian.Uint32(data[offset+3 : offset+7])}}
					offset += 7
				}
			case byte(IMem):
			case byte(Imm):
				data = append(data, mem.ReadMemoryN(*pc+uint32(offset+1), 4)...)
				operands[i] = Operand{Type: Imm, Value: &ImmOperand{Value: binary.LittleEndian.Uint32(data[offset+1 : offset+5])}}
				offset += 5
			}
		}
	}
	*pc += uint32(offset)
	inst.Operands = operands
	return inst
}
