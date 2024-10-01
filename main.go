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
			&cpu.LD{Register: c.Reg.B, Address: 0x0003}, // Load the value at address 0x0004 into register B
			&cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.B}, // Add the value in register B to the value in register A
			&cpu.ST{Register: c.Reg.A, Address: 0x7C00}, // Store the value in register A at address 0x0005
      &cpu.HLT{},
    },
  }

  c.StoreProgram(program.Encode())

  c.Mem.Write(0x0001, 0x04)
  c.Mem.Write(0x0002, 0x03)
	c.Mem.Write(0x0003, 0x30) // 0 in ASCII

	simulationDelay := 1000 // ms per instruction

  if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Title = "Text-mode video buffer"
	p.SetRect(0, 0, 42, 27)

	regDump := widgets.NewParagraph()
	regDump.Title = "Reg"
	regDump.SetRect(42, 0, 52, 10)

	simInfo := widgets.NewParagraph()
	simInfo.Title = "Sim Info"
	simInfo.Text = fmt.Sprintf("Frequency: %d Hz", 1000/simulationDelay)
	simInfo.SetRect(52, 0, 72, 10)

	ui.Render(p, regDump, simInfo)

	run := false

  uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Millisecond * time.Duration(simulationDelay))
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "+":
				simulationDelay /= 10
				ticker.Reset(time.Millisecond * time.Duration(simulationDelay))
			case "-":
				simulationDelay *= 10
				ticker.Reset(time.Millisecond * time.Duration(simulationDelay))
			case "s":
				c.Step()
			case "r":
				run = true
			case "p":
				run = false
			case "c":
				c.Reset()
				c.StoreProgram(program.Encode())
			}	
		case <-ticker.C:
			if run {
				c.Step()
			}
			p.Text = string(c.Mem.RAM.Data[0x7C00:0x7FFF])
			ui.Render(p)

			regDump.Text = fmt.Sprintf("A: %02x\nB: %02x\nC: %02x\nD: %02x\nE: %02x\nH: %02x\nL: %02x\nPC: %04x\n",
				c.Reg.A.Read(),
				c.Reg.B.Read(),
				c.Reg.C.Read(),
				c.Reg.D.Read(),
				c.Reg.E.Read(),
				c.Reg.H.Read(),
				c.Reg.L.Read(),
				c.PC.Read(),
				)
			ui.Render(regDump)

			simInfo.Text = fmt.Sprintf("Frequency: %d Hz\nHalted: %t\nRunning: %t\n\nStep: <s>\nRun: <r>\nPause: <p>\nClear: <c>", 1000/simulationDelay, c.Halt, run)
			ui.Render(simInfo)
		}
	}
}
