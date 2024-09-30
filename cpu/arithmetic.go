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
