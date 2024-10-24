package main

import "fmt"

type Memory struct {
	RAM  *RAM
	ROM  *ROM
	IVT  *IVT
	VRAM *VRAM
}

func NewMemory() *Memory {
	return &Memory{
		RAM:  NewRAM(),
		ROM:  NewROM(),
		IVT:  NewIVT(),
		VRAM: NewVRAM(),
	}
}

func (m *Memory) LoadProgram(startAddr uint32, program []byte) {
	switch {
	case startAddr < 0x80000000:
		for i, v := range program {
			m.RAM.Write(startAddr+uint32(i), v)
		}
	case startAddr < 0x88000000:
		for i, v := range program {
			m.ROM.Write(startAddr-0x80000000+uint32(i), v)
		}
	case startAddr >= 0xFFFFF000:
		for i, v := range program {
			m.VRAM.Write(startAddr-0xFFFFF000+uint32(i), v)
		}
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) CanRead(addr uint32) bool {
	switch {
	case addr < 0x80000000:
		return true
	case addr < 0x88000000:
		return true
	case addr < 0x88000400:
		return true
	case addr >= 0xFFFFF000:
		return true
	default:
		return false
	}
}

func (m *Memory) Read(addr uint32) uint8 {
	switch {
	case addr < 0x80000000:
		return m.RAM.Read(addr)
	case addr < 0x88000000:
		return m.ROM.Read(addr - 0x80000000)
	case addr < 0x88000400:
		return m.IVT.Read(addr - 0x88000000)
	case addr >= 0xFFFFF000:
		return m.VRAM.Read(addr - 0xFFFFF000)
	default:
		panic("Addressing unuseable memory: " + fmt.Sprintf("%x", addr))
	}
}

func (m *Memory) ReadWord(addr uint32) uint16 {
	switch {
	case addr < 0x80000000:
		return m.RAM.ReadWord(addr)
	case addr < 0x88000000:
		return m.ROM.ReadWord(addr - 0x80000000)
	case addr < 0x88000400:
		return m.IVT.ReadWord(addr - 0x88000000)
	case addr >= 0xFFFFF000:
		panic("VRAM ReadWord")
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) ReadDWord(addr uint32) uint32 {
	switch {
	case addr < 0x80000000:
		return m.RAM.ReadDWord(addr)
	case addr < 0x88000000:
		return m.ROM.ReadDWord(addr - 0x80000000)
	case addr < 0x88000400:
		return m.IVT.ReadDWord(addr - 0x88000000)
	case addr >= 0xFFFFF000:
		panic("VRAM ReadDWord")
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) ReadN(addr uint32, n uint32) []uint8 {
	data := make([]uint8, n)
	for i := uint32(0); i < n; i++ {
		data[i] = m.Read(addr + i)
	}
	return data
}

func (m *Memory) ReadString(addr uint32) string {
	var data []uint8
	for {
		d := m.Read(addr)
		if d == 0 {
			break
		}
		data = append(data, d)
		addr++
	}
	return string(data)
}

func (m *Memory) Write(addr uint32, data uint8) {
	switch {
	case addr < 0x80000000:
		m.RAM.Write(addr, data)
	case addr < 0x88000000:
		panic("ROM Write")
	case addr < 0x88000400:
		m.IVT.Write(addr-0x88000000, data)
	case addr >= 0xFFFFF000:
		m.VRAM.Write(addr-0xFFFFF000, data)
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) WriteWord(addr uint32, data uint16) {
	switch {
	case addr < 0x80000000:
		m.RAM.WriteWord(addr, data)
	case addr < 0x88000000:
		panic("ROM WriteWord")
	case addr < 0x88000400:
		m.IVT.WriteWord(addr-0x88000000, data)
	case addr >= 0xFFFFF000:
		panic("VRAM WriteWord")
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) WriteDWord(addr uint32, data uint32) {
	switch {
	case addr < 0x80000000:
		m.RAM.WriteDWord(addr, data)
	case addr < 0x88000000:
		panic("ROM WriteDWord")
	case addr < 0x88000400:
		m.IVT.WriteDWord(addr-0x88000000, data)
	case addr >= 0xFFFFF000:
		panic("VRAM WriteDWord")
	default:
		panic("Addressing unuseable memory")
	}
}

func (m *Memory) Clear() {
	m.RAM.Clear()
	m.ROM.Clear()
	m.IVT.Clear()
	m.VRAM.Clear()
}

type RAM struct {
	mem [0x80000000]uint8
}

func (r *RAM) Read(addr uint32) uint8 {
	return r.mem[addr]
}

func (r *RAM) ReadWord(addr uint32) uint16 {
	return uint16(r.mem[addr]) | uint16(r.mem[addr+1])<<8
}

func (r *RAM) ReadDWord(addr uint32) uint32 {
	return uint32(r.mem[addr]) | uint32(r.mem[addr+1])<<8 | uint32(r.mem[addr+2])<<16 | uint32(r.mem[addr+3])<<24
}

func (r *RAM) Write(addr uint32, data uint8) {
	r.mem[addr] = data
}

func (r *RAM) WriteWord(addr uint32, data uint16) {
	r.mem[addr] = uint8(data)
	r.mem[addr+1] = uint8(data >> 8)
}

func (r *RAM) WriteDWord(addr uint32, data uint32) {
	r.mem[addr] = uint8(data)
	r.mem[addr+1] = uint8(data >> 8)
	r.mem[addr+2] = uint8(data >> 16)
	r.mem[addr+3] = uint8(data >> 24)
}

func (r *RAM) Clear() {
	for i := range r.mem {
		r.mem[i] = 0
	}
}

func NewRAM() *RAM {
	return &RAM{}
}

type IVT struct {
	mem [0x400]uint8
}

func (r *IVT) Read(addr uint32) uint8 {
	return r.mem[addr]
}

func (r *IVT) ReadWord(addr uint32) uint16 {
	return uint16(r.mem[addr]) | uint16(r.mem[addr+1])<<8
}

func (r *IVT) ReadDWord(addr uint32) uint32 {
	return uint32(r.mem[addr]) | uint32(r.mem[addr+1])<<8 | uint32(r.mem[addr+2])<<16 | uint32(r.mem[addr+3])<<24
}

func (r *IVT) Write(addr uint32, data uint8) {
	r.mem[addr] = data
}

func (r *IVT) WriteWord(addr uint32, data uint16) {
	r.mem[addr] = uint8(data)
	r.mem[addr+1] = uint8(data >> 8)
}

func (r *IVT) WriteDWord(addr uint32, data uint32) {
	r.mem[addr] = uint8(data)
	r.mem[addr+1] = uint8(data >> 8)
	r.mem[addr+2] = uint8(data >> 16)
	r.mem[addr+3] = uint8(data >> 24)
}

func (r *IVT) Clear() {
	for i := range r.mem {
		r.mem[i] = 0
	}
}

func NewIVT() *IVT {
	return &IVT{}
}

type ROM struct {
	mem [0x8000000]uint8
}

func (r *ROM) Read(addr uint32) uint8 {
	return r.mem[addr]
}

func (r *ROM) ReadWord(addr uint32) uint16 {
	return uint16(r.mem[addr]) | uint16(r.mem[addr+1])<<8
}

func (r *ROM) ReadDWord(addr uint32) uint32 {
	return uint32(r.mem[addr]) | uint32(r.mem[addr+1])<<8 | uint32(r.mem[addr+2])<<16 | uint32(r.mem[addr+3])<<24
}

func (r *ROM) Write(addr uint32, data uint8) {
	r.mem[addr] = data
}

func (r *ROM) Clear() {
	for i := range r.mem {
		r.mem[i] = 0
	}
}

func NewROM() *ROM {
	return &ROM{}
}

type VRAM struct {
	mem [0x400]uint8
}

func (v *VRAM) Read(addr uint32) uint8 {
	return v.mem[addr]
}

func (v *VRAM) Write(addr uint32, data uint8) {
	v.mem[addr] = data
}

func (v *VRAM) Clear() {
	for i := range v.mem {
		v.mem[i] = 0
	}
}

func NewVRAM() *VRAM {
	return &VRAM{}
}
