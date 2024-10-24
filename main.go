package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	generateBytecode := flag.Bool("bytecode", false, "Generate bytecode")
	outputFilename := flag.String("output", "output.bin", "Output filename")
	fsType := flag.String("fs", "folder", "Filesystem type")
	fsRoot := flag.String("root", "./vmdata", "Root folder")
	callTable := flag.Bool("calltable", false, "Generate call table")
	flag.Parse()

	var fs VFS
	if *fsType == "folder" {
		if _, err := os.Stat(*fsRoot); os.IsNotExist(err) {
			err := os.Mkdir(*fsRoot, 0755)
			if err != nil {
				log.Fatalf("failed to create root folder: %v", err)
			}
		}
		fs = &FolderBasedVFS{Root: *fsRoot}
	} else {
		log.Fatalf("unknown filesystem type")
	}

	var bc *Bytecode
	isAsm := false
	p := NewParser()
	p.DefaultBaseAddress = 0x00000000
	for _, filename := range flag.Args() {
		if strings.HasSuffix(filename, ".asm") {
			isAsm = true
			p.AddFile(filename)

		} else if strings.HasSuffix(flag.Args()[0], ".bin") {
			if isAsm {
				log.Fatalf("cannot mix .asm and .bin files")
			}
			if bc != nil {
				log.Fatalf("cannot load multiple .bin files")
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

		if *callTable {
			for k, v := range p.Labels {
				fmt.Printf("%s: %08x\n", k, v)
			}
			return
		}
	}

	if *callTable {
		log.Fatalln("Only able to generate call table from .asm files")
		return
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
	c.FileSystem = fs
	c.LoadProgram(bc)

	simulationDelay := time.Millisecond * 100

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	video := widgets.NewParagraph()
	video.Title = "Text-mode video buffer"
	video.SetRect(0, 0, 42, 27)

	regDump := widgets.NewParagraph()
	regDump.Title = "Reg"
	regDump.SetRect(42, 0, 72, 10)

	simInfo := widgets.NewParagraph()
	simInfo.Title = "Sim Info"
	simInfo.SetRect(72, 0, 91, 10)

	memoryWindow := widgets.NewParagraph()
	memoryWindow.Title = "Program"
	memoryWindow.SetRect(42, 10, 74, 27)

	accessWindow := widgets.NewParagraph()
	accessWindow.Title = "Access"
	accessWindow.SetRect(74, 10, 91, 27)

	stackWindow := widgets.NewParagraph()
	stackWindow.Title = "Stack"
	stackWindow.SetRect(91, 0, 101, 27)

	heapWindow := widgets.NewParagraph()
	heapWindow.Title = "Heap"
	heapWindow.SetRect(101, 0, 111, 27)

	ui.Render(video, regDump, simInfo, memoryWindow, accessWindow, stackWindow, heapWindow)

	run := false

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(simulationDelay)
	isEscaped := true

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-<Backspace>>":
				isEscaped = !isEscaped
			case "q", "<C-c>":
				if isEscaped {
					return
				} else {
					c.InputQueue <- e.ID
				}

			default:
				if isEscaped {
					switch e.ID {
					case "+":
						simulationDelay /= 10
						ticker.Reset(simulationDelay)
					case "-":
						simulationDelay *= 10
						ticker.Reset(simulationDelay)
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
				} else {
					c.InputQueue <- e.ID
				}
			}
		case <-ticker.C:
			if run {
				c.Step()
			}
			video.Text = ""
			for i := 0; i < 37*27; i++ {
				if c.MemoryManager.ReadMemory(0xFFFFF000+uint32(i)) == 0 {
					video.Text += " "
					continue
				}
				video.Text += fmt.Sprintf("%c", c.MemoryManager.ReadMemory(0xFFFFF000+uint32(i)))
			}

			regDump.Text = ""
			for i, v := range c.Registers[:8] {
				regDump.Text += fmt.Sprintf("R%d: %08x | R%d: %08x\n", i, v, i+8, c.Registers[i+8])
			}

			simInfo.Text = fmt.Sprintf("Frequency: %s\nHalted: %t\nRunning: %t\nEscaped: %t\n\nPC: %08x\nSP: %08x\nHP: %08x", DurationToFrequency(simulationDelay), c.Halted, run, isEscaped, c.Registers[16], c.Registers[17], c.Registers[18])

			memoryWindow.Text = drawMemoryWindow(c.MemoryManager, c.Registers[16])
			accessWindow.Text = drawAccessWindow(c.MemoryManager, c.LastAccessedAddress)
			stackWindow.Text = ""
			stackMemory := c.MemoryManager.ReadMemoryN(c.Registers[17], int(c.MemoryManager.VirtualStackEnd-c.Registers[17]))
			for i := 0; i < len(stackMemory); i += 4 {
				if i+4 <= len(stackMemory) {
					v := binary.LittleEndian.Uint32(stackMemory[i : i+4])
					stackWindow.Text += fmt.Sprintf("%08x\n", v)
				}
			}

			heapWindow.Text = ""
			heapMemory := c.MemoryManager.ReadMemoryN(c.MemoryManager.VirtualHeapStart, int(c.Registers[18]))
			for i := 0; i < len(heapMemory); i += 4 {
				if i+4 <= len(heapMemory) {
					v := binary.LittleEndian.Uint32(heapMemory[i : i+4])
					heapWindow.Text += fmt.Sprintf("%08x\n", v)
				}
			}
			ui.Render(video, regDump, simInfo, memoryWindow, accessWindow, stackWindow, heapWindow)
		}
	}
}

func DurationToFrequency(d time.Duration) string {
	if d <= time.Nanosecond {
		return fmt.Sprintf("%d GHz", time.Nanosecond/d)
	} else if d <= time.Microsecond {
		return fmt.Sprintf("%d MHz", time.Microsecond/d)
	} else if d <= time.Millisecond {
		return fmt.Sprintf("%d KHz", time.Millisecond/d)
	}
	return fmt.Sprintf("%d Hz", time.Second/d)
}

func drawMemoryWindow(mem *MemoryManager, programCounter uint32) string {
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
		if i == programCounter && instructionSet[mem.ReadMemory(i)] != nil {
			memoryWindow += fmt.Sprintf(">%08x: %02x %s\n", i, mem.ReadMemory(i), instructionSet[mem.ReadMemory(i)].Name)
		} else {
			memoryWindow += fmt.Sprintf(" %08x: %02x\n", i, mem.ReadMemory(i))
		}
	}

	return memoryWindow
}

func drawAccessWindow(mem *MemoryManager, lastAccess uint32) string {
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
			memoryWindow += fmt.Sprintf(">%08x: %02x\n", i, mem.ReadMemory(i))
		} else {
			memoryWindow += fmt.Sprintf(" %08x: %02x\n", i, mem.ReadMemory(i))
		}
	}

	return memoryWindow
}
