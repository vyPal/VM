package main

import (
	"fmt"

	"src.vypal.me/vyPal/VM/cpu"
)

func main() {
  c := cpu.NewCPU()

  program := cpu.Program{
    Instructions: []cpu.Instruction{
      &cpu.LD{Register: c.Reg.A, Address: 0x0001},
      &cpu.LD{Register: c.Reg.B, Address: 0x0002},
      &cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.B},
      &cpu.ST{Register: c.Reg.A, Address: 0x0003},
      &cpu.HLT{},
    },
  }

  c.StoreProgram(program.Encode())

  c.Mem.Write(0x0001, 0x01)
  c.Mem.Write(0x0002, 0x02)

  c.Run()

  fmt.Println(c.Mem.Read(0x0003))
}
