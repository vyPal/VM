package main

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	PageSize  = 4096
	PageCount = 0x80000
	RAMStart  = 0x00000000
	RAMEnd    = 0x7FFFFFFF
	ROMStart  = 0x80000000
	ROMEnd    = 0x87FFFFFF
	VRAMStart = 0xFFFFF000
	VRAMEnd   = 0xFFFFFFFF
)

type ProgramInfo struct {
	StartAddress uint32
	Size         uint32
	Sectors      []ProgramInfoSector
}

type ProgramInfoSector struct {
	StartAddress uint32
	Bytecode     []byte
	IsStart      bool
}

type MemoryManager struct {
	Memory           *Memory
	cpu              *CPU
	Programs         []*ProgramInfo
	PageTable        map[uint32]uint32
	FreeFrames       []uint32
	VirtualStackEnd  uint32
	VirtualHeapStart uint32
}

func NewMemoryManager(cpu *CPU, memory *Memory) *MemoryManager {
	mm := &MemoryManager{
		Memory:           memory,
		cpu:              cpu,
		Programs:         []*ProgramInfo{},
		PageTable:        make(map[uint32]uint32),
		FreeFrames:       []uint32{},
		VirtualStackEnd:  0x7FFFFFFF,
		VirtualHeapStart: 0x00000000,
	}

	mm.cpu.Registers[17] = mm.VirtualStackEnd
	mm.cpu.Registers[18] = mm.VirtualHeapStart

	for i := uint32(0); i < PageCount; i++ {
		mm.FreeFrames = append(mm.FreeFrames, i)
	}

	return mm
}

func (mm *MemoryManager) AllocateFrame() (uint32, error) {
	if len(mm.FreeFrames) == 0 {
		return 0, fmt.Errorf("out of memory")
	}

	frame := mm.FreeFrames[0]
	mm.FreeFrames = mm.FreeFrames[1:]

	return frame, nil
}

func (mm *MemoryManager) FreeFrame(frame uint32) {
	mm.FreeFrames = append(mm.FreeFrames, frame)
}

func (mm *MemoryManager) MapVirtualToPhysical(virtualAddr uint32) error {
	virtualPageNum := virtualAddr / PageSize
	if _, exists := mm.PageTable[virtualPageNum]; !exists {
		physicalPageIndex, err := mm.AllocateFrame()
		if err != nil {
			return err
		}
		mm.PageTable[virtualPageNum] = physicalPageIndex
	}
	return nil
}

func (mm *MemoryManager) TranslateAddress(virtualAddr uint32) (uint32, error) {
	if virtualAddr >= ROMStart || virtualAddr <= ROMEnd {
		return virtualAddr, nil
	} else if virtualAddr >= VRAMStart || virtualAddr <= VRAMEnd {
		return virtualAddr, nil
	}
	virtualPageNum := virtualAddr / PageSize
	offset := virtualAddr % PageSize

	if physicalPageIndex, exists := mm.PageTable[virtualPageNum]; exists {
		return uint32(physicalPageIndex)*PageSize + offset, nil
	}

	return 0, errors.New("unmapped memory access")
}

func (mm *MemoryManager) CanRead(addr uint32) bool {
	addr, err := mm.TranslateAddress(addr)
	if err != nil {
		return false
	}
	return mm.Memory.CanRead(addr)
}

func (mm *MemoryManager) ReadMemory(addr uint32) byte {
	physAddr, err := mm.TranslateAddress(addr)
	if err != nil {
		panic(err)
	}
	return mm.Memory.Read(physAddr)
}

func (mm *MemoryManager) ReadMemoryWord(addr uint32) uint16 {
	data, err := mm.ReadNMemory(addr, 2)
	if err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint16(data)
}

func (mm *MemoryManager) ReadMemoryDWord(addr uint32) uint32 {
	data, err := mm.ReadNMemory(addr, 4)
	if err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(data)
}

func (mm *MemoryManager) ReadMemoryString(addr uint32) string {
	var str []byte
	for {
		ch := mm.ReadMemory(addr)
		if ch == 0 {
			break
		}
		str = append(str, ch)
		addr++
	}
	return string(str)
}

func (mm *MemoryManager) ReadMemoryN(addr uint32, n int) []byte {
	data, err := mm.ReadNMemory(addr, n)
	if err != nil {
		panic(err)
	}
	return data
}

func (mm *MemoryManager) ReadNMemory(addr uint32, n int) ([]byte, error) {
	data := make([]byte, n)
	for i := 0; i < n; i++ {
		physAddr, err := mm.TranslateAddress(addr + uint32(i))
		if err != nil {
			return nil, err
		}
		data[i] = mm.Memory.Read(physAddr)
	}
	return data, nil
}

func (mm *MemoryManager) WriteMemory(addr uint32, value byte) {
	physAddr, err := mm.TranslateAddress(addr)
	if err != nil {
		panic(err)
	}
	mm.Memory.Write(physAddr, value)
}

func (mm *MemoryManager) WriteMemoryWord(addr uint32, value uint16) {
	valueBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(valueBytes, value)
	err := mm.WriteNMemory(addr, valueBytes)
	if err != nil {
		panic(err)
	}
}

func (mm *MemoryManager) WriteMemoryDWord(addr uint32, value uint32) {
	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, value)
	err := mm.WriteNMemory(addr, valueBytes)
	if err != nil {
		panic(err)
	}
}

func (mm *MemoryManager) WriteNMemory(addr uint32, data []byte) error {
	for i, value := range data {
		physAddr, err := mm.TranslateAddress(addr + uint32(i))
		if err != nil {
			return err
		}
		mm.Memory.Write(physAddr, value)
	}
	return nil
}

func (mm *MemoryManager) Push(value uint32) {
	if mm.cpu.Registers[17]-4 < mm.cpu.Registers[18] {
		if err := mm.GrowStack(); err != nil {
			panic(err)
		}
	}

	mm.cpu.Registers[17] -= 4

	if err := mm.MapVirtualToPhysical(mm.cpu.Registers[17]); err != nil {
		panic(err)
	}

	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, value)
	err := mm.WriteNMemory(mm.cpu.Registers[17], valueBytes)
	if err != nil {
		panic(err)
	}
}

func (mm *MemoryManager) Pop() uint32 {
	if mm.cpu.Registers[17] >= mm.VirtualStackEnd {
		panic("stack underflow")
	}

	valueBytes, err := mm.ReadNMemory(mm.cpu.Registers[17], 4)
	if err != nil {
		panic(err)
	}

	value := binary.LittleEndian.Uint32(valueBytes)

	mm.cpu.Registers[17] += 4

	mm.TryShrinkStack()

	return value
}

func (mm *MemoryManager) Malloc(size uint32) (uint32, error) {
	alignedSize := (size + 3) & ^uint32(3)
	startAddr := mm.cpu.Registers[18]
	endAddr := startAddr + alignedSize

	for endAddr > mm.cpu.Registers[18] {
		if err := mm.GrowHeap(); err != nil {
			return 0, err
		}
	}

	return startAddr, nil
}

func (mm *MemoryManager) Free(addr uint32, size uint32) {
	alignedSize := (size + 3) & ^uint32(3)
	startAddr := addr
	endAddr := startAddr + alignedSize

	for freeAddr := startAddr; freeAddr < endAddr; freeAddr += PageSize {
		pageNum := freeAddr / PageSize
		if physicalPageIndex, exists := mm.PageTable[pageNum]; exists {
			mm.WriteNMemory(freeAddr, make([]byte, PageSize))
			delete(mm.PageTable, pageNum)
			mm.FreeFrame(physicalPageIndex)
		}
	}

	mm.TryShrinkHeap()
}

func (mm *MemoryManager) GrowStack() error {
	newStackPtr := mm.cpu.Registers[17] - PageSize
	if newStackPtr <= mm.cpu.Registers[18] {
		return errors.New("cannot grow stack: collision with heap")
	}

	if err := mm.MapVirtualToPhysical(newStackPtr); err != nil {
		return err
	}

	mm.cpu.Registers[17] = newStackPtr
	return nil
}

func (mm *MemoryManager) TryShrinkStack() {
	if (mm.cpu.Registers[17]%PageSize == 0) && (mm.cpu.Registers[17] < mm.VirtualStackEnd) {
		topPageStart := mm.cpu.Registers[17]
		topPageEnd := topPageStart + PageSize

		isEmpty := true
		for addr := topPageStart; addr < topPageEnd; addr += 4 {
			value, err := mm.ReadNMemory(addr, 4)
			if err != nil || binary.LittleEndian.Uint32(value) != 0 {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			mm.UnmapPage(topPageStart)
			mm.cpu.Registers[17] += PageSize
		}
	}
}

func (mm *MemoryManager) GrowHeap() error {
	newHeapPtr := mm.cpu.Registers[18] + PageSize
	if newHeapPtr >= mm.cpu.Registers[17] {
		return errors.New("cannot grow heap: collision with stack")
	}

	if err := mm.MapVirtualToPhysical(mm.cpu.Registers[18]); err != nil {
		return err
	}

	mm.cpu.Registers[18] = newHeapPtr
	return nil
}

func (mm *MemoryManager) TryShrinkHeap() {
	if (mm.cpu.Registers[18]%PageSize == 0) && (mm.cpu.Registers[18] > mm.VirtualHeapStart) {
		topPageStart := mm.cpu.Registers[18] - PageSize
		topPageEnd := mm.cpu.Registers[18]

		isEmpty := true
		for addr := topPageStart; addr < topPageEnd; addr += 4 {
			value, err := mm.ReadNMemory(addr, 4)
			if err != nil || binary.LittleEndian.Uint32(value) != 0 {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			mm.UnmapPage(topPageStart)
			mm.cpu.Registers[18] -= PageSize
		}
	}
}

func (mm *MemoryManager) UnmapPage(addr uint32) {
	pageNum := addr / PageSize
	if physicalPageIndex, exists := mm.PageTable[pageNum]; exists {
		delete(mm.PageTable, pageNum)
		mm.FreeFrame(physicalPageIndex)
	}
}

func (mm *MemoryManager) ExecuteJump(currentPC uint32, jumpAddr uint32) uint32 {
	for _, info := range mm.Programs {
		if currentPC >= info.StartAddress && currentPC < info.StartAddress+info.Size {
			translatedAddr, err := mm.TranslateAddress(info.StartAddress + jumpAddr)
			if err != nil {
				panic(err)
			}
			if translatedAddr == 0xFFFFFFFF {
				panic("invalid jump address")
			}
			return translatedAddr
		}
	}
	return jumpAddr
}

func (mm *MemoryManager) NewProgram() *ProgramInfo {
	program := &ProgramInfo{
		Sectors: []ProgramInfoSector{},
	}
	mm.Programs = append(mm.Programs, program)
	return program
}

func (mm *MemoryManager) AddSector(programInfo *ProgramInfo, baseAddress uint32, program []byte, isStart bool) {
	sector := ProgramInfoSector{
		StartAddress: baseAddress,
		Bytecode:     program,
		IsStart:      isStart,
	}
	programInfo.Sectors = append(programInfo.Sectors, sector)
}

func (mm *MemoryManager) LoadProgram(programInfo *ProgramInfo) uint32 {
	var totalSize uint32
	for _, sector := range programInfo.Sectors {
		totalSize += uint32(len(sector.Bytecode))
	}

	startAddr, err := mm.Malloc(totalSize)
	if err != nil {
		panic(err)
	}

	programInfo.StartAddress = startAddr

	for _, sector := range programInfo.Sectors {
		if sector.IsStart {
			programInfo.StartAddress = startAddr
		}
		mm.WriteNMemory(startAddr, sector.Bytecode)
		programInfo.Size += uint32(len(sector.Bytecode))
		startAddr += uint32(len(sector.Bytecode))
	}

	return programInfo.StartAddress
}

func (mm *MemoryManager) UnloadProgram(startAddr uint32) error {
	programInfo := mm.Programs[startAddr]
	if programInfo == nil {
		return errors.New("program not found")
	}

	mm.Free(programInfo.StartAddress, programInfo.Size)

	mm.Programs[startAddr] = nil

	return nil
}
