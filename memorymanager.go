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
}

type MemoryManager struct {
	Memory           *Memory
	Programs         []*ProgramInfo
	PageTable        map[uint32]uint32
	FreeFrames       []uint32
	VirtualStackEnd  uint32
	VirtualStackPtr  uint32
	VirtualHeapStart uint32
	VirtualHeapPtr   uint32
}

func NewMemoryManager(memory *Memory) *MemoryManager {
	mm := &MemoryManager{
		Memory:           memory,
		Programs:         []*ProgramInfo{},
		PageTable:        make(map[uint32]uint32),
		FreeFrames:       []uint32{},
		VirtualStackEnd:  0x7FFFFFFF,
		VirtualStackPtr:  0x7FFFFFFF,
		VirtualHeapStart: 0x00000000,
		VirtualHeapPtr:   0x00000000,
	}

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
	if mm.VirtualStackPtr-4 < mm.VirtualHeapPtr {
		if err := mm.GrowStack(); err != nil {
			panic(err)
		}
	}

	mm.VirtualStackPtr -= 4

	if err := mm.MapVirtualToPhysical(mm.VirtualStackPtr); err != nil {
		panic(err)
	}

	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, value)
	err := mm.WriteNMemory(mm.VirtualStackPtr, valueBytes)
	if err != nil {
		panic(err)
	}
}

func (mm *MemoryManager) Pop() uint32 {
	if mm.VirtualStackPtr >= mm.VirtualStackEnd {
		panic("stack underflow")
	}

	valueBytes, err := mm.ReadNMemory(mm.VirtualStackPtr, 4)
	if err != nil {
		panic(err)
	}

	value := binary.LittleEndian.Uint32(valueBytes)

	mm.VirtualStackPtr += 4

	mm.TryShrinkStack()

	return value
}

func (mm *MemoryManager) Malloc(size uint32) (uint32, error) {
	alignedSize := (size + 3) & ^uint32(3)
	startAddr := mm.VirtualHeapPtr
	endAddr := startAddr + alignedSize

	for endAddr > mm.VirtualStackPtr {
		if err := mm.GrowHeap(); err != nil {
			return 0, err
		}
	}

	for addr := startAddr; addr < endAddr; addr += PageSize {
		if err := mm.MapVirtualToPhysical(addr); err != nil {
			for freeAddr := startAddr; freeAddr < addr; freeAddr += PageSize {
				mm.Free(freeAddr)
			}
			return 0, err
		}
	}

	mm.VirtualHeapPtr = endAddr
	return startAddr, nil
}

func (mm *MemoryManager) Free(addr uint32) {
	pageNum := addr / PageSize
	if physicalPageIndex, exists := mm.PageTable[pageNum]; exists {
		delete(mm.PageTable, pageNum)
		mm.FreeFrame(physicalPageIndex)
	}

	if addr == mm.VirtualHeapPtr-PageSize {
		for mm.VirtualHeapPtr > mm.VirtualHeapStart &&
			mm.PageTable[mm.VirtualHeapPtr/PageSize-1] == 0 {
			mm.VirtualHeapPtr -= PageSize
		}
	}

	mm.TryShrinkHeap()
}

func (mm *MemoryManager) GrowStack() error {
	newStackPtr := mm.VirtualStackPtr - PageSize
	if newStackPtr <= mm.VirtualHeapPtr {
		return errors.New("cannot grow stack: collision with heap")
	}

	if err := mm.MapVirtualToPhysical(newStackPtr); err != nil {
		return err
	}

	mm.VirtualStackPtr = newStackPtr
	return nil
}

func (mm *MemoryManager) TryShrinkStack() {
	if (mm.VirtualStackPtr%PageSize == 0) && (mm.VirtualStackPtr < mm.VirtualStackEnd) {
		topPageStart := mm.VirtualStackPtr
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
			mm.VirtualStackPtr += PageSize
		}
	}
}

func (mm *MemoryManager) GrowHeap() error {
	newHeapPtr := mm.VirtualHeapPtr + PageSize
	if newHeapPtr >= mm.VirtualStackPtr {
		return errors.New("cannot grow heap: collision with stack")
	}

	if err := mm.MapVirtualToPhysical(mm.VirtualHeapPtr); err != nil {
		return err
	}

	mm.VirtualHeapPtr = newHeapPtr
	return nil
}

func (mm *MemoryManager) TryShrinkHeap() {
	if (mm.VirtualHeapPtr%PageSize == 0) && (mm.VirtualHeapPtr > mm.VirtualHeapStart) {
		topPageStart := mm.VirtualHeapPtr - PageSize
		topPageEnd := mm.VirtualHeapPtr

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
			mm.VirtualHeapPtr -= PageSize
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

func (mm *MemoryManager) AddSector(programInfo *ProgramInfo, baseAddress uint32, program []byte) {
	sector := ProgramInfoSector{
		StartAddress: baseAddress,
		Bytecode:     program,
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

	for addr := startAddr; addr < startAddr+programInfo.Size; addr += PageSize {
		mm.Free(addr)
	}

	mm.Programs[startAddr] = nil

	return nil
}
