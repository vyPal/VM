package main

import (
	"fmt"
	"os"
)

func main() {
  c := NewCPU()
  fmt.Println("Parsing file", os.Args[1])
  p := &Parser{Filename: os.Args[1]}
  p.Parse()
  c.LoadProgram(p.Program)
  c.PC = 0x80000000
  fmt.Printf("Loaded %d bytes\n", len(p.Program))
  fmt.Printf("Program: %v\n", p.Program)
  fmt.Printf("Value at 0: %v\n", c.Memory.Read(0x0))
  for !c.Halted {
    c.Step()
  }
  fmt.Printf("New value at 0: %v\n", c.Memory.Read(0x0))
}
