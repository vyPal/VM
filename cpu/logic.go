package cpu

import (
	"src.vypal.me/vyPal/VM/opcodes"
)

type AND struct {
  Register1 *Register
  Register2 *Register
}

func (op *AND) Execute(cpu *CPU) error {
  op.Register1.And(op.Register2.Read())
  return nil
}

func (op *AND) Encode() []byte {
  return []byte{byte(opcodes.AND), op.Register1.ID, op.Register2.ID}
}

type OR struct {
  Register1 *Register
  Register2 *Register
}

func (op *OR) Execute(cpu *CPU) error {
  op.Register1.Or(op.Register2.Read())
  return nil
}

func (op *OR) Encode() []byte {
  return []byte{byte(opcodes.OR), op.Register1.ID, op.Register2.ID}
}

type XOR struct {
  Register1 *Register
  Register2 *Register
}

func (op *XOR) Execute(cpu *CPU) error {
  op.Register1.Xor(op.Register2.Read())
  return nil
}

func (op *XOR) Encode() []byte {
  return []byte{byte(opcodes.XOR), op.Register1.ID, op.Register2.ID}
}

type NOT struct {
  Register *Register
}

func (op *NOT) Execute(cpu *CPU) error {
  op.Register.Not()
  return nil
}

func (op *NOT) Encode() []byte {
  return []byte{byte(opcodes.NOT), op.Register.ID}
}

type SHL struct {
  Register *Register
}

func (op *SHL) Execute(cpu *CPU) error {
  op.Register.ShiftLeft()
  return nil
}

func (op *SHL) Encode() []byte {
  return []byte{byte(opcodes.SHL), op.Register.ID}
}

type SHR struct {
  Register *Register
}

func (op *SHR) Execute(cpu *CPU) error {
  op.Register.ShiftRight()
  return nil
}

func (op *SHR) Encode() []byte {
  return []byte{byte(opcodes.SHR), op.Register.ID}
}
