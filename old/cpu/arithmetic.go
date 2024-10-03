package cpu

import (
	"src.vypal.me/vyPal/VM/opcodes"
)

type ADD struct {
  Register1 *Register
  Register2 *Register
}

func (op *ADD) Execute(cpu *CPU) error {
  op.Register1.Add(op.Register2.Read())
  return nil
}

func (op *ADD) Encode() []byte {
  return []byte{byte(opcodes.ADD), op.Register1.ID, op.Register2.ID}
}

type SUB struct {
  Register1 *Register
  Register2 *Register
}

func (op *SUB) Execute(cpu *CPU) error {
  op.Register1.Subtract(op.Register2.Read())
  return nil
}

func (op *SUB) Encode() []byte {
  return []byte{byte(opcodes.SUB), op.Register1.ID, op.Register2.ID}
}

type MUL struct {
  Register1 *Register
  Register2 *Register
}

func (op *MUL) Execute(cpu *CPU) error {
  op.Register1.Multiply(op.Register2.Read())
  return nil
}

func (op *MUL) Encode() []byte {
  return []byte{byte(opcodes.MUL), op.Register1.ID, op.Register2.ID}
}

type DIV struct {
  Register1 *Register
  Register2 *Register
}

func (op *DIV) Execute(cpu *CPU) error {
  op.Register1.Divide(op.Register2.Read())
  return nil
}

func (op *DIV) Encode() []byte {
  return []byte{byte(opcodes.DIV), op.Register1.ID, op.Register2.ID}
}

type MOD struct {
  Register1 *Register
  Register2 *Register
}

func (op *MOD) Execute(cpu *CPU) error {
  op.Register1.Modulo(op.Register2.Read())
  return nil
}

func (op *MOD) Encode() []byte {
  return []byte{byte(opcodes.MOD), op.Register1.ID, op.Register2.ID}
}
