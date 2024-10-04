package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
  c := NewCPU()
  fmt.Println("Parsing file", os.Args[1])
  p := &Parser{Filename: os.Args[1]}
  p.BaseAddress = 0x80000000
  p.Parse()
  c.LoadProgram(p.Program)
  
  simulationDelay := 100 // ms per instruction

  if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	video := widgets.NewParagraph()
  video.Title = "Text-mode video buffer"
	video.SetRect(0, 0, 42, 27)

	regDump := widgets.NewParagraph()
	regDump.Title = "Reg"
	regDump.SetRect(42, 0, 57, 10)

	simInfo := widgets.NewParagraph()
	simInfo.Title = "Sim Info"
	simInfo.Text = fmt.Sprintf("Frequency: %d Hz", 1000/simulationDelay)
	simInfo.SetRect(57, 0, 77, 10)

	memoryWindow := widgets.NewParagraph()
	memoryWindow.Title = "Program"
	memoryWindow.SetRect(42, 10, 61, 27)

	accessWindow := widgets.NewParagraph()
	accessWindow.Title = "Access"
	accessWindow.SetRect(61, 10, 77, 27)

	stackWindow := widgets.NewParagraph()
	stackWindow.Title = "Stack"
	stackWindow.SetRect(77, 0, 95, 27)

	ui.Render(video, regDump, simInfo, memoryWindow, accessWindow, stackWindow)

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
				c.LoadProgram(p.Program)
			}	
		case <-ticker.C:
			if run {
				c.Step()
			}
			video.Text = string(c.Memory.VRAM.mem[:])

			regDump.Text = ""
			for i, v := range c.Registers {
				regDump.Text += fmt.Sprintf("R%d: %08x\n", i, v)
			}

			simInfo.Text = fmt.Sprintf("Frequency: %d Hz\nHalted: %t\nRunning: %t\n\nStep: <s>\nRun: <r>\nPause: <p>\nClear: <c>", 1000/simulationDelay, c.Halted, run)

			memoryWindow.Text = drawMemoryWindow(c.Memory, c.PC)
			accessWindow.Text = drawAccessWindow(c.Memory, c.LastAccessedAddress)
			stackWindow.Text = ""
			for _, v := range slices.Backward(c.Stack.Stack[:c.Stack.SP]) {
				stackWindow.Text += fmt.Sprintf("%04x\n", v)
			}
			ui.Render(video, regDump, simInfo, memoryWindow, accessWindow, stackWindow)
		}
	}
}

func drawMemoryWindow(mem *Memory, programCounter uint32) string {
	linesBefore := 7
	linesAfter := 7

	if programCounter < 8 {
		linesBefore = int(programCounter)
		linesAfter = 15 - linesBefore
	}
	if programCounter > 0x88000000 - 8 {
		linesAfter = 0x88000000 - int(programCounter)
		linesBefore = 15 - linesAfter
	}

	var memoryWindow string
	for i := programCounter - uint32(linesBefore); i < programCounter + uint32(linesAfter); i++ {
		if i == programCounter {
			memoryWindow += fmt.Sprintf(">%08x: %02x %s\n", i, mem.Read(i), instructionSet[mem.Read(i)].Name)
		} else {
			memoryWindow += fmt.Sprintf(" %08x: %02x\n", i, mem.Read(i))
		}
	}

	return memoryWindow
}

func drawAccessWindow(mem *Memory, lastAccess uint32) string {
	linesBefore := 7
	linesAfter := 7

	if lastAccess < 8 {
		linesBefore = int(lastAccess)
		linesAfter = 15 - linesBefore
	}
	if lastAccess > 0x88000000 - 8 {
		linesAfter = 0x88000000 - int(lastAccess)
		linesBefore = 15 - linesAfter
	}

	var memoryWindow string
	for i := lastAccess - uint32(linesBefore); i < lastAccess + uint32(linesAfter); i++ {
		if i == lastAccess {
			memoryWindow += fmt.Sprintf(">%08x: %02x\n", i, mem.Read(i))
		} else {
			memoryWindow += fmt.Sprintf(" %08x: %02x\n", i, mem.Read(i))
		}
	}

	return memoryWindow
}
