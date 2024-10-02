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

	constantBlock := &cpu.ListDataBlock{
		Data: []byte{
			10, 23, 0x30, 10, 100,
		},
	}

  program := cpu.Program{
		DataBlock: constantBlock,
    Instructions: []cpu.Instruction{
      &cpu.LD{Register: c.Reg.A, Address: constantBlock.GetAddr(0)}, // Load the value at address 0x0001 into register A
      &cpu.LD{Register: c.Reg.B, Address: constantBlock.GetAddr(1)}, // Load the value at address 0x0002 into register B
      &cpu.SUB{Register1: c.Reg.A, Register2: c.Reg.B}, // Add the value in register B to the value in register A
			&cpu.ST{Register: c.Reg.A, Address: 0x0000}, // Store the result in register A at address 0x0000
			// Get first digit of the result (max 3 digits) and display into video buffer
			&cpu.LD{Register: c.Reg.B, Address: constantBlock.GetAddr(4)},
			&cpu.DIV{Register1: c.Reg.A, Register2: c.Reg.B},
			&cpu.LD{Register: c.Reg.C, Address: constantBlock.GetAddr(2)},
			&cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.C},
			&cpu.ST{Register: c.Reg.A, Address: 0x7C00},
			// Get second digit of the result (max 3 digits) and display into video buffer
			&cpu.LD{Register: c.Reg.A, Address: 0x0000},
			&cpu.MOD{Register1: c.Reg.A, Register2: c.Reg.B},
			&cpu.LD{Register: c.Reg.D, Address: constantBlock.GetAddr(3)},
			&cpu.DIV{Register1: c.Reg.A, Register2: c.Reg.D},
			&cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.C},
			&cpu.ST{Register: c.Reg.A, Address: 0x7C01},
			// Get third digit of the result (max 3 digits) and display into video buffer
			&cpu.LD{Register: c.Reg.A, Address: 0x0000},
			&cpu.MOD{Register1: c.Reg.A, Register2: c.Reg.D},
			&cpu.ADD{Register1: c.Reg.A, Register2: c.Reg.C},
			&cpu.ST{Register: c.Reg.A, Address: 0x7C02},
      &cpu.HLT{},
    },
  }

  c.StoreProgram(program.Encode())

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

	memoryWindow := widgets.NewParagraph()
	memoryWindow.Title = "Memory"
	memoryWindow.SetRect(42, 10, 72, 27)

	ui.Render(p, regDump, simInfo, memoryWindow)

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
				run = false
				c.Reset()
				c.StoreProgram(program.Encode())
			}	
		case <-ticker.C:
			if run {
				c.Step()
			}
			p.Text = string(c.Mem.RAM.Data[0x7C00:0x7FFF])

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

			simInfo.Text = fmt.Sprintf("Frequency: %d Hz\nHalted: %t\nRunning: %t\n\nStep: <s>\nRun: <r>\nPause: <p>\nClear: <c>", 1000/simulationDelay, c.Halt, run)

			memoryWindow.Text = drawMemoryWindow(c.Mem, c.PC.Read())
			ui.Render(p, regDump, simInfo, memoryWindow)
		}
	}
}

func drawMemoryWindow(mem *cpu.Memory, programCounter uint16) string {
	// 0x0000 - 0x7BFF: RAM
	// 0x7C00 - 0x7FFF: Video buffer
	// 0x8000 - 0xFFFF: ROM

	// If it is possible, show the memory around the program counter. Add splitters to show the different memory regions.
	// Show the memory in hex format, with the program counter highlighted.

	linesBefore := 7
	linesAfter := 7

	if programCounter < 8 {
		linesBefore = int(programCounter)
	}
	if programCounter > 0x8000 + uint16(len(mem.ROM.Data)) - 8 {
		linesAfter = 0x8000 + len(mem.ROM.Data) - int(programCounter)
	}

	var memoryWindow string
	for i := programCounter - uint16(linesBefore); i < programCounter + uint16(linesAfter); i++ {
		if i == programCounter {
			memoryWindow += fmt.Sprintf(">%04x: %02x\n", i, mem.Read(i))
		} else {
			memoryWindow += fmt.Sprintf(" %04x: %02x\n", i, mem.Read(i))
		}
	}

	return memoryWindow
}
