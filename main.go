package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	generateBytecode := flag.Bool("bytecode", false, "Generate bytecode")
	outputFilename := flag.String("output", "output.bin", "Output filename")
	flag.Parse()

	var bc *Bytecode
	isAsm := false
	p := NewParser()
	p.DefaultBaseAddress = 0x80000000
	for _, filename := range flag.Args() {
		if strings.HasSuffix(filename, ".asm") {
			isAsm = true
			p.AddFile(filename)

		} else if strings.HasSuffix(flag.Args()[0], ".bin") {
			if isAsm {
				log.Fatalf("cannot mix .asm and .bin files")
			}
			fileContent, err := os.ReadFile(filename)
			if err != nil {
				log.Fatalf("failed to read file: %v", err)
			}
			bc, err = DecodeBytecode(fileContent)
		} else {
			log.Fatalf("unknown file type")
		}
	}

	if isAsm {
		p.Parse()
		err := p.CheckForOverlappingSectors()
		if err != nil {
			log.Fatalf("overlapping sectors: %v", err)
		}
		bc = ProgramToBytecode(p)
	}

	if *generateBytecode {
		fileContent, err := EncodeBytecode(bc)
		if err != nil {
			log.Fatalf("failed to encode bytecode: %v", err)
		}
		err = os.WriteFile(*outputFilename, fileContent, 0644)
		if err != nil {
			log.Fatalf("failed to write file: %v", err)
		}
		return
	}

	c := NewCPU()
	c.LoadProgram(bc)

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
				c.LoadProgram(bc)
			}
		case <-ticker.C:
			if run {
				c.Step()
			}
			video.Text = ""
			for i := 0; i < 37*27; i++ {
				if c.Memory.Read(0xFFFFF000+uint32(i)) == 0 {
					video.Text += " "
					continue
				}
				video.Text += fmt.Sprintf("%c", c.Memory.Read(0xFFFFF000+uint32(i)))
			}

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
	if programCounter > 0xFFFFFFFF-8 {
		linesAfter = 0xFFFFFFFF - int(programCounter)
		linesBefore = 15 - linesAfter
	}

	var memoryWindow string
	for i := programCounter - uint32(linesBefore); i < programCounter+uint32(linesAfter); i++ {
		if !mem.CanRead(i) {
			memoryWindow += fmt.Sprintf(" %08x: ??\n", i)
			continue
		}
		if i == programCounter && instructionSet[mem.Read(i)] != nil {
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
	if lastAccess > 0xFFFFFFFF-8 {
		linesAfter = 0xFFFFFFFF - int(lastAccess)
		linesBefore = 15 - linesAfter
	}

	var memoryWindow string
	for i := lastAccess - uint32(linesBefore); i < lastAccess+uint32(linesAfter); i++ {
		if !mem.CanRead(i) {
			memoryWindow += fmt.Sprintf(" %08x: ??\n", i)
			continue
		}
		if i == lastAccess {
			memoryWindow += fmt.Sprintf(">%08x: %02x\n", i, mem.Read(i))
		} else {
			memoryWindow += fmt.Sprintf(" %08x: %02x\n", i, mem.Read(i))
		}
	}

	return memoryWindow
}
