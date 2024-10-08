package cpu

import (
	"src.vypal.me/vyPal/VM/memory"
	"src.vypal.me/vyPal/VM/opcodes"
)

type CPU struct {
  Reg *Registers
  Mem *Memory
  PC *LargeRegister
  Stack *LiFo
  Halt bool
  ShouldIncrement bool
  LastAccessedAddress uint16
}

type LiFo struct {
  Data []uint16
}

func (l *LiFo) Push(val uint16) {
  l.Data = append(l.Data, val)
}

func (l *LiFo) Pop() uint16 {
  val := l.Data[len(l.Data) - 1]
  l.Data = l.Data[:len(l.Data) - 1]
  return val
}

type Memory struct {
  RAM *memory.RandomAccessMemory
  ROM *memory.ReadOnlyMemory
}

type Instruction interface {
  Execute(cpu *CPU) error
  Encode() []byte
}

type DataBlock interface {
  Encode() []byte
}

type ListDataBlock struct {
  Data []byte
}

func (l *ListDataBlock) Encode() []byte {
  return l.Data
}

func (l *ListDataBlock) GetAddr(id int) uint16 {
  return 0x8000 + 0x3 + uint16(id)
}

type Program struct {
  DataBlock DataBlock
  Instructions []Instruction
}

func (p *Program) Encode() []byte {
  var encoded []byte
  if p.DataBlock != nil {
    jmp := JMP{Address: uint16(0x8000 + 0x3 + len(p.DataBlock.Encode()))}
    encoded = append(encoded, jmp.Encode()...)
    encoded = append(encoded, p.DataBlock.Encode()...)
  }
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
  return &CPU{NewRegisters(), &Memory{memory.NewRandomAccessMemory(0x8000), memory.NewReadOnlyMemory(nil)}, &LargeRegister{}, &LiFo{}, false, true, 0}
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

func (cpu *CPU) Reset() {
  cpu.Reg.Reset()
  cpu.PC.Write(0)
  cpu.Mem.RAM.Data = make([]byte, 0x8000)
  cpu.Mem.ROM = memory.NewReadOnlyMemory(nil)
  cpu.Halt = false
  cpu.ShouldIncrement = true
}

func (cpu *CPU) Step() {
  if cpu.Halt {
    return
  }
  inst := decode(cpu)
  if inst == nil {
    panic("Unknown instruction")
  }
  cpu.ShouldIncrement = true
  inst.Execute(cpu)
  if cpu.ShouldIncrement {
    cpu.PC.Increment()
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
    case byte(opcodes.MUL):
      return &MUL{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.DIV):
      return &DIV{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.MOD):
      return &MOD{
        Register1: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
        Register2: cpu.Reg.Get(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.CALL):
      return &CALL{
        Address: uint16(cpu.Mem.Read(cpu.PC.Increment())) << 8 | uint16(cpu.Mem.Read(cpu.PC.Increment())),
      }
    case byte(opcodes.RET):
      return &RET{}
    case byte(opcodes.HLT):
      return &HLT{}
  }
  return nil
}
