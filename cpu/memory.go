package cpu

import (
	"src.vypal.me/vyPal/VM/opcodes"
)

type LD struct {
  Register *Register
  Address uint16
}

func (op *LD) Execute(cpu *CPU) error {
  op.Register.Write(cpu.Mem.Read(op.Address))
  return nil
}

func (op LD) Encode() []byte {
  return []byte{byte(opcodes.LD), op.Register.ID, byte(op.Address >> 8), byte(op.Address)}
}

type ST struct {
  Register *Register
  Address uint16
}

func (op *ST) Execute(cpu *CPU) error {
  cpu.Mem.Write(op.Address, op.Register.Read())
  return nil
}

func (op *ST) Encode() []byte {
  return []byte{byte(opcodes.ST), op.Register.ID, byte(op.Address >> 8), byte(op.Address)}
}
