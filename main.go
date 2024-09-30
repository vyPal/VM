package main

import (
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"src.vypal.me/vyPal/VM/cpu"
)

func main() {
  c := cpu.NewCPU()

  program := cpu.Program{
    Instructions: []cpu.Instruction{
      &cpu.LD{Register: c.Reg.A, Address: 0x0001}, // Load the value at address 0x0001 into register A
      &cpu.LD{Register: c.Reg.B, Address: 0x0002}, // Load the value at address 0x0002 into register B
      &cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.B}, // Add the value in register B to the value in register A
      &cpu.ST{Register: c.Reg.A, Address: 0x0003}, // Store the value in register A at address 0x0003
			&cpu.LD{Register: c.Reg.B, Address: 0x0004}, // Load the value at address 0x0004 into register B
			&cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.B}, // Add the value in register B to the value in register A
			&cpu.ST{Register: c.Reg.A, Address: 0x7C00}, // Store the value in register A at address 0x0005
      &cpu.HLT{},
    },
  }

  // Print out the encoded program in both hex and binary
  fmt.Printf("%x\n", program.Encode())
  fmt.Printf("%08b\n", program.Encode())

  c.StoreProgram(program.Encode())

  c.Mem.Write(0x0001, 0x01)
  c.Mem.Write(0x0002, 0x02)
	c.Mem.Write(0x0004, 0x30) // 0 in ASCII

  if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Title = "Text-mode video buffer"
	p.SetRect(0, 0, 42, 27)

	ui.Render(p)

  uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Millisecond).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}	
		case <-ticker:
			c.Step()
			p.Text = string(c.Mem.RAM.Data[0x7C00:0x7FFF])
			ui.Render(p)
		}
	}
}
