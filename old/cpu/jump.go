package cpu

import (
	"src.vypal.me/vyPal/VM/opcodes"
)

type JMP struct {
  Address uint16
}

func (op *JMP) Execute(cpu *CPU) error {
  cpu.PC.Write(op.Address)
  cpu.ShouldIncrement = false
  return nil
}

func (op *JMP) Encode() []byte {
  return []byte{byte(opcodes.JMP), byte(op.Address >> 8), byte(op.Address)}
}

type JZ struct {
  Address uint16
}

func (op *JZ) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() == 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JZ) Encode() []byte {
  return []byte{byte(opcodes.JZ), byte(op.Address >> 8), byte(op.Address)}
}

type JNZ struct {
  Address uint16
}

func (op *JNZ) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() != 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JNZ) Encode() []byte {
  return []byte{byte(opcodes.JNZ), byte(op.Address >> 8), byte(op.Address)}
}

type JG struct {
  Address uint16
}

func (op *JG) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() > 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JG) Encode() []byte {
  return []byte{byte(opcodes.JG), byte(op.Address >> 8), byte(op.Address)}
}

type JGE struct {
  Address uint16
}

func (op *JGE) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() >= 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JGE) Encode() []byte {
  return []byte{byte(opcodes.JGE), byte(op.Address >> 8), byte(op.Address)}
}

type JL struct {
  Address uint16
}

func (op *JL) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() < 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JL) Encode() []byte {
  return []byte{byte(opcodes.JL), byte(op.Address >> 8), byte(op.Address)}
}

type JLE struct {
  Address uint16
}

func (op *JLE) Execute(cpu *CPU) error {
  if cpu.Reg.A.Read() <= 0 {
    cpu.PC.Write(op.Address)
    cpu.ShouldIncrement = false
  }
  return nil
}

func (op *JLE) Encode() []byte {
  return []byte{byte(opcodes.JLE), byte(op.Address >> 8), byte(op.Address)}
}

type CALL struct {
  Address uint16
}

func (op *CALL) Execute(cpu *CPU) error {
  cpu.Stack.Push(cpu.PC.Read())
  cpu.PC.Write(op.Address)
  cpu.ShouldIncrement = false
  return nil
}

func (op *CALL) Encode() []byte {
  return []byte{byte(opcodes.CALL), byte(op.Address >> 8), byte(op.Address)}
}

type RET struct {
}

func (op *RET) Execute(cpu *CPU) error {
  cpu.PC.Write(cpu.Stack.Pop())
  return nil
}

func (op *RET) Encode() []byte {
  return []byte{byte(opcodes.RET)}
}

type HLT struct {
}

func (op *HLT) Execute(cpu *CPU) error {
  cpu.Halt = true
  cpu.ShouldIncrement = false
  return nil
}

func (op *HLT) Encode() []byte {
  return []byte{byte(opcodes.HLT)}
}
