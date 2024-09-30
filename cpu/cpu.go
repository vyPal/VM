package cpu

import (
	"src.vypal.me/vyPal/VM/memory"
	"src.vypal.me/vyPal/VM/opcodes"
)

type CPU struct {
  Reg *Registers
  Mem *Memory
  ACC *LargeRegister
  PC *LargeRegister
  Halt bool
  ShouldIncrement bool
}

type Memory struct {
  RAM *memory.RandomAccessMemory
  ROM *memory.ReadOnlyMemory
}

type Instruction interface {
  Execute(cpu *CPU) error
  Encode() []byte
}

type Program struct {
  Instructions []Instruction
}

func (p *Program) Encode() []byte {
  var encoded []byte
  for _, inst := range p.Instructions {
    encoded = append(encoded, inst.Encode()...)
  }
  return encoded
}

func (mem *Memory) Read(addr uint16) byte {
  if addr < 0x8000 {
    return mem.RAM.Read(addr)
  } else {
    return mem.ROM.Read(addr - 0x8000)
  }
}

func (mem *Memory) Write(addr uint16, val byte) {
  if addr < 0x8000 {
    mem.RAM.Write(addr, val)
  } else {
    panic("Cannot write to ROM")
  }
}

func NewCPU() *CPU {
  return &CPU{NewRegisters(), &Memory{memory.NewRandomAccessMemory(0x400), memory.NewReadOnlyMemory(nil)}, &LargeRegister{}, &LargeRegister{}, false, true}
}

func (cpu *CPU) StoreProgram(program []byte) {
  cpu.Mem.ROM.Data = program
  cpu.PC.Write(0x8000)
}

func (cpu *CPU) StoreProgramInRAM(program []byte) {
  for i, b := range program {
    cpu.Mem.RAM.Write(uint16(i), b)
  }
  cpu.PC.Write(0x0000)
}

func (cpu *CPU) Step() {
  inst := decode(cpu)
  if inst == nil {
    panic("Unknown instruction")
  }
  inst.Execute(cpu)
  if cpu.ShouldIncrement {
    cpu.PC.Increment()
  }
}

func (cpu *CPU) Run() {
  for !cpu.Halt {
    cpu.Step()
  }
}

func decode(cpu *CPU) Instruction {
  firstByte := cpu.Mem.Read(cpu.PC.Read())
  switch firstByte {
    case byte(opcodes.LD):
      return &LD{
        Register: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.ST):
      return &ST{
        Register: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JMP):
      return &JMP{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JZ):
      return &JZ{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JNZ):
      return &JNZ{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JG):
      return &JG{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JGE):
      return &JGE{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JL):
      return &JL{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.JLE):
      return &JLE{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.AND):
      return &AND{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.OR):
      return &OR{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.XOR):
      return &XOR{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.NOT):
      return &NOT{
        Register: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.SHL):
      return &SHL{
        Register: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.SHR):
      return &SHR{
        Register: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.ADD):
      return &ADD{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.SUB):
      return &SUB{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.HLT):
      return &HLT{}
  }
  return nil
}
