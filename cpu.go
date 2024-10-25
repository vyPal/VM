package main

import (
	"fmt"
	"strings"
)

const (
	KeyDown  uint32 = 0x01 << 24
	KeyPress uint32 = 0x02 << 24
	KeyUp    uint32 = 0x03 << 24
)

type CPU struct {
	MemoryManager       *MemoryManager
	Registers           [19]uint32 // 0-15: General purpose (15 can be overwritten by interrupts), 16: Instruction register, 17: Stack pointer, 18: Heap pointer
	Halted              bool
	LastAccessedAddress uint32
	FileSystem          VFS
	FileTable           map[uint32]interface{}
	NextFD              uint32
	InputQueue          chan string
	InterruptPending    bool
	InterruptProcessing bool
	OriginalPC          uint32
	InterruptVector     uint32
	InterruptData       uint32
	InterruptReturned   chan bool
}

func NewCPU() *CPU {
	cpu := &CPU{
		Registers:         [19]uint32{},
		Halted:            false,
		FileTable:         make(map[uint32]interface{}),
		NextFD:            0,
		InputQueue:        make(chan string),
		InterruptReturned: make(chan bool),
	}
	cpu.MemoryManager = NewMemoryManager(cpu, NewMemory())
	go cpu.KeyboardInputLoop()
	return cpu
}

func (c *CPU) handleKeyCombo(key string) []string {
	if strings.HasPrefix(key, "<") && strings.HasSuffix(key, ">") {
		keyCombo := strings.Trim(key, "<>")
		parts := strings.Split(keyCombo, "-")

		if len(parts) == 1 {
			switch parts[0] {
			case "Space":
				return []string{"KeyPress "}
			case "Enter":
				return []string{"KeyPress\n"}
			case "Tab":
				return []string{"KeyPress\t"}
			}
		}

		var keyEvents []string

		for _, part := range parts[:len(parts)-1] {
			switch part {
			case "C":
				keyEvents = append(keyEvents, "KeyDownControl")
			}
		}

		keyEvents = append(keyEvents, "KeyPress"+parts[len(parts)-1])

		for i := len(parts) - 2; i >= 0; i-- {
			switch parts[i] {
			case "C":
				keyEvents = append(keyEvents, "KeyUpControl")
			}
		}

		return keyEvents
	}
	return []string{"KeyPress" + key}
}

func (c *CPU) KeyboardInputLoop() {
	var inputBuffer []string

	for {
		select {
		case key := <-c.InputQueue:
			keyEvents := c.handleKeyCombo(key)

			inputBuffer = append(inputBuffer, keyEvents...)

		case <-c.InterruptReturned:
			if len(inputBuffer) > 0 {
				bufferedKey := inputBuffer[0]
				inputBuffer = inputBuffer[1:]

				c.processKeyEvent(bufferedKey)
			}
		}
	}
}

func (c *CPU) processKeyEvent(event string) {
	var eventType uint32
	var asciiCode uint32

	switch event {
	case "KeyDownControl":
		eventType = KeyDown
		asciiCode = 17

	case "KeyUpControl":
		eventType = KeyUp
		asciiCode = 17

	default:
		if len(event) > 8 && event[:8] == "KeyPress" {
			eventType = KeyPress
			key := rune(event[8])
			asciiCode = uint32(int(key))
		}
	}

	c.InterruptPending = true
	c.InterruptVector = 1
	c.InterruptData = eventType | asciiCode
}

func (c *CPU) Reset() {
	c.Registers = [19]uint32{}
	c.MemoryManager = NewMemoryManager(c, NewMemory())
	c.Halted = false
	for _, v := range c.FileTable {
		if v == nil {
			continue
		}
		c.FileSystem.Close(v)
	}
	c.FileTable = make(map[uint32]interface{})
	c.NextFD = 0
}

func (c *CPU) Step() {
	if c.Halted {
		return
	}
	if c.InterruptPending {
		c.InterruptPending = false
		c.InterruptProcessing = true
		c.OriginalPC = c.Registers[16]
		c.MemoryManager.Push(c.Registers[16])
		if c.MemoryManager.ReadMemoryDWord(0x88000000+c.InterruptVector*4) == 0 {
			c.Registers[16] = c.MemoryManager.Pop()
			c.InterruptProcessing = false
		} else {
			c.Registers[16] = c.MemoryManager.ReadMemoryDWord(0x88000000 + c.InterruptVector*4)
			c.Registers[10] = c.InterruptData
		}
	} else if !c.InterruptProcessing {
		c.InterruptReturned <- true
	}
	instr := DecodeInstruction(c.MemoryManager, &c.Registers[16])
	instr.Execute(c, instr.Operands)
	if c.InterruptProcessing {
		if c.Registers[16] == c.OriginalPC {
			c.InterruptProcessing = false
			c.InterruptReturned <- true
		}
	}
}

func (c *CPU) LoadProgram(program *Bytecode) {
	var p *ProgramInfo
	for _, sector := range program.Sectors {
		if sector.Bytecode != nil {
			if sector.StartAddress != 0x0 && sector.StartAddress < 0x80000000 {
				panic(fmt.Sprintf("Specifying a explicit start address for a sector to be stored in RAM is not allowed as it may interfere with dynamic memory allocation."))
			} else if sector.StartAddress == 0x0 {
				if p == nil {
					p = c.MemoryManager.NewProgram()
				}
				c.MemoryManager.AddSector(p, sector.StartAddress, sector.Bytecode, program.StartAddress == sector.StartAddress)
			} else {
				c.MemoryManager.Memory.LoadProgram(sector.StartAddress, sector.Bytecode)
			}
		}
	}

	c.Registers[16] = program.StartAddress

	if p != nil {
		start := c.MemoryManager.LoadProgram(p)
		if c.Registers[16] < 0x80000000 {
			c.Registers[16] = start
		}
	}
}
