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
  Size byte // 0x0 = 32-bit, 0x1 = 16-bit, 0x2 = 8-bit
}

type DMemOperand struct {
  Addr uint32
}

type IMemOperand struct {
  Addr uint32
}

type ImmOperand struct {
  Value uint32
}

type Operand struct {
  Type OperandType
  AllowedTypes []OperandType
  Value interface{}
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
    Name: "NOP",
    Execute: func(cpu *CPU, operands []Operand) {
      return
    },
  },
  0x01: {
    Opcode: 0x01,
    Name: "LD",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] = cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] = uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] = uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] = cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] = uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] = uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "ST",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        r := operands[1].Value.(*RegOperand)
        if r.Size == 0x0 {
          cpu.Memory.WriteDWord(operands[0].Value.(*DMemOperand).Addr, cpu.Registers[r.RegNum])
        } else if r.Size == 0x1 {
          cpu.Memory.WriteWord(operands[0].Value.(*DMemOperand).Addr, uint16(cpu.Registers[r.RegNum]))
        } else if r.Size == 0x2 {
          cpu.Memory.Write(operands[0].Value.(*DMemOperand).Addr, uint8(cpu.Registers[r.RegNum]))
        }
      case IMem:
        r := operands[1].Value.(*RegOperand)
        if r.Size == 0x0 {
          cpu.Memory.WriteDWord(operands[0].Value.(*IMemOperand).Addr, cpu.Registers[r.RegNum])
        } else if r.Size == 0x1 {
          cpu.Memory.WriteWord(operands[0].Value.(*IMemOperand).Addr, uint16(cpu.Registers[r.RegNum]))
        } else if r.Size == 0x2 {
          cpu.Memory.Write(operands[0].Value.(*IMemOperand).Addr, uint8(cpu.Registers[r.RegNum]))
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
    Name: "ADD",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] += cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] += uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] += uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] += cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] += uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] += uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "SUB",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] -= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] -= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] -= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] -= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] -= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] -= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "MUL",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] *= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] *= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] *= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] *= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] *= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] *= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "DIV",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] /= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] /= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] /= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] /= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] /= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] /= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "MOD",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] %= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] %= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] %= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] %= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] %= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] %= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "AND",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] &= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] &= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] &= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] &= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] &= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] &= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "OR",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] |= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] |= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] |= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] |= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] |= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] |= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "XOR",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] ^= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] ^= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] ^= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] ^= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] ^= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] ^= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "NOT",
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
    Name: "SHL",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] <<= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] <<= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] <<= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] <<= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] <<= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] <<= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "SHR",
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
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] >>= cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] >>= uint32(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] >>= uint32(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr))
        }
      case IMem:
        if r.Size == 0x0 {
          cpu.Registers[r.RegNum] >>= cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)
        } else if r.Size == 0x1 {
          cpu.Registers[r.RegNum] >>= uint32(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr))
        } else if r.Size == 0x2 {
          cpu.Registers[r.RegNum] >>= uint32(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr))
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
    Name: "CMP",
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
        if r.Size == 0x0 {
          if cpu.Registers[r.RegNum] == cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr) {
            cpu.Registers[0xF] = 0x0
          } else if cpu.Registers[r.RegNum] > uint32(cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if cpu.Registers[r.RegNum] < uint32(cpu.Memory.ReadDWord(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x2
          }
        } else if r.Size == 0x1 {
          if uint16(cpu.Registers[r.RegNum]) == uint16(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x0
          } else if uint16(cpu.Registers[r.RegNum]) > uint16(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if uint16(cpu.Registers[r.RegNum]) < uint16(cpu.Memory.ReadWord(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x2
          }
        } else if r.Size == 0x2 {
          if uint8(cpu.Registers[r.RegNum]) == uint8(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x0
          } else if uint8(cpu.Registers[r.RegNum]) > uint8(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if uint8(cpu.Registers[r.RegNum]) < uint8(cpu.Memory.Read(operands[1].Value.(*DMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x2
          }
        }
      case IMem:
        if r.Size == 0x0 {
          if cpu.Registers[r.RegNum] == cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr) {
            cpu.Registers[0xF] = 0x0
          } else if cpu.Registers[r.RegNum] > uint32(cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if cpu.Registers[r.RegNum] < uint32(cpu.Memory.ReadDWord(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x2
          }
        } else if r.Size == 0x1 {
          if uint16(cpu.Registers[r.RegNum]) == uint16(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x0
          } else if uint16(cpu.Registers[r.RegNum]) > uint16(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if uint16(cpu.Registers[r.RegNum]) < uint16(cpu.Memory.ReadWord(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x2
          }
        } else if r.Size == 0x2 {
          if uint8(cpu.Registers[r.RegNum]) == uint8(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x0
          } else if uint8(cpu.Registers[r.RegNum]) > uint8(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr)) {
            cpu.Registers[0xF] = 0x1
          } else if uint8(cpu.Registers[r.RegNum]) < uint8(cpu.Memory.Read(operands[1].Value.(*IMemOperand).Addr)) {
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
    Name: "JMP",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        cpu.Registers[0xF] = cpu.PC
        cpu.PC = operands[0].Value.(*DMemOperand).Addr
      case IMem:
        cpu.Registers[0xF] = cpu.PC
        cpu.PC = operands[0].Value.(*IMemOperand).Addr
      case Imm:
        cpu.Registers[0xF] = cpu.PC
        cpu.PC = operands[0].Value.(*ImmOperand).Value
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x10: {
    Opcode: 0x10,
    Name: "JEQ",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] == 0x0 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] == 0x0 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] == 0x0 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x11: {
    Opcode: 0x11,
    Name: "JNE",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] != 0x0 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] != 0x0 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] != 0x0 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x12: {
    Opcode: 0x12,
    Name: "JGT",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x13: {
    Opcode: 0x13,
    Name: "JLT",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x14: {
    Opcode: 0x14,
    Name: "JGE",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x1 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x15: {
    Opcode: 0x15,
    Name: "JLE",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*DMemOperand).Addr
        }
      case IMem:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*IMemOperand).Addr
        }
      case Imm:
        if cpu.Registers[0xF] == 0x0 || cpu.Registers[0xF] == 0x2 {
          cpu.PC = operands[0].Value.(*ImmOperand).Value
        }
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x16: {
    Opcode: 0x16,
    Name: "CALL",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case DMem:
        cpu.Stack.Push(cpu.PC)
        cpu.PC = operands[0].Value.(*DMemOperand).Addr
      case IMem:
        cpu.Stack.Push(cpu.PC)
        cpu.PC = operands[0].Value.(*IMemOperand).Addr
      case Imm:
        cpu.Stack.Push(cpu.PC)
        cpu.PC = operands[0].Value.(*ImmOperand).Value
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{DMem, IMem, Imm}}, // A - Dest
    },
  },
  0x17: {
    Opcode: 0x17,
    Name: "RET",
    Execute: func(cpu *CPU, operands []Operand) {
      cpu.PC = cpu.Stack.Pop()
    },
  },
  0x18: {
    Opcode: 0x18,
    Name: "PUSH",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case Reg:
        cpu.Stack.Push(cpu.Registers[operands[0].Value.(*RegOperand).RegNum])
      case DMem:
        cpu.Stack.Push(cpu.Memory.ReadDWord(operands[0].Value.(*DMemOperand).Addr))
      case IMem:
        cpu.Stack.Push(uint32(cpu.Memory.ReadWord(operands[0].Value.(*IMemOperand).Addr)))
      case Imm:
        cpu.Stack.Push(operands[0].Value.(*ImmOperand).Value)
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{Reg, DMem, IMem, Imm}}, // A - Source
    },
  },
  0x19: {
    Opcode: 0x19,
    Name: "POP",
    Execute: func(cpu *CPU, operands []Operand) {
      switch operands[0].Type {
      case Reg:
        cpu.Registers[operands[0].Value.(*RegOperand).RegNum] = cpu.Stack.Pop()
      case DMem:
        cpu.Memory.WriteDWord(operands[0].Value.(*DMemOperand).Addr, cpu.Stack.Pop())
      case IMem:
        cpu.Memory.WriteWord(operands[0].Value.(*IMemOperand).Addr, uint16(cpu.Stack.Pop()))
      }
    },
    Operands: []Operand{
      {AllowedTypes: []OperandType{Reg, DMem, IMem}}, // A - Dest
    },
  },
  0x1A: {
    Opcode: 0x1A,
    Name: "HLT",
    Execute: func(cpu *CPU, operands []Operand) {
      cpu.Halted = true
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
        binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
      case IMem:
        binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
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
        binary.Write(&buf, binary.LittleEndian, operand.Value.(*DMemOperand).Addr)
      case IMem:
        buf.WriteByte(byte(IMem))
        binary.Write(&buf, binary.LittleEndian, operand.Value.(*IMemOperand).Addr)
      case Imm:
        buf.WriteByte(byte(Imm))
        binary.Write(&buf, binary.LittleEndian, operand.Value.(*ImmOperand).Value)
      }
    }
  }
  return buf.Bytes()
}

func DecodeInstruction(mem *Memory, pc *uint32) *Instruction {
  data := []byte{mem.Read(*pc)}
  inst := GetInstructionByOpcode(data[0])
  operands := make([]Operand, len(inst.Operands))
  offset := 1
  for i, operand := range inst.Operands {
    if len(operand.AllowedTypes) == 0 {
      switch operand.Type {
      case Reg:
        data = append(data, mem.Read(*pc+uint32(offset)))
        operands[i] = Operand{Type: Reg, Value: &RegOperand{RegNum: data[offset] & 0xF, Size: data[offset] >> 4}}
        offset++
      case DMem:
        data = append(data, mem.ReadN(*pc+uint32(offset), 4)...)
        operands[i] = Operand{Type: DMem, Value: &DMemOperand{Addr: binary.LittleEndian.Uint32(data[offset:offset+4])}}
        offset += 4
      case IMem:
        data = append(data, mem.ReadN(*pc+uint32(offset), 4)...)
        operands[i] = Operand{Type: IMem, Value: &IMemOperand{Addr: binary.LittleEndian.Uint32(data[offset:offset+4])}}
        offset += 4
      case Imm:
        data = append(data, mem.ReadN(*pc+uint32(offset), 4)...)
        operands[i] = Operand{Type: Imm, Value: &ImmOperand{Value: binary.LittleEndian.Uint32(data[offset:offset+4])}}
        offset += 4
      }
    } else {
      data = append(data, mem.Read(*pc+uint32(offset)))
      switch data[offset] {
      case byte(Reg):
        data = append(data, mem.Read(*pc+uint32(offset+1)))
        operands[i] = Operand{Type: Reg, Value: &RegOperand{RegNum: data[offset+1] & 0xF, Size: data[offset+1] >> 4}}
        offset += 2
      case byte(DMem):
        data = append(data, mem.ReadN(*pc+uint32(offset+1), 4)...)
        operands[i] = Operand{Type: DMem, Value: &DMemOperand{Addr: binary.LittleEndian.Uint32(data[offset+1:offset+5])}}
        offset += 5
      case byte(IMem):
        data = append(data, mem.ReadN(*pc+uint32(offset+1), 4)...)
        operands[i] = Operand{Type: IMem, Value: &IMemOperand{Addr: binary.LittleEndian.Uint32(data[offset+1:offset+5])}}
        offset += 5
      case byte(Imm):
        data = append(data, mem.ReadN(*pc+uint32(offset+1), 4)...)
        operands[i] = Operand{Type: Imm, Value: &ImmOperand{Value: binary.LittleEndian.Uint32(data[offset+1:offset+5])}}
        offset += 5
      }
    }
  }
  *pc += uint32(offset)
  inst.Operands = operands
  return inst
}
