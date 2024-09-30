package memory

type Memory interface {
  Read(addr uint16) byte
  Write(addr uint16, val byte)
}

type RandomAccessMemory struct {
  Data []byte
}

type ReadOnlyMemory struct {
  Data []byte
}

func (ram *RandomAccessMemory) Read(addr uint16) byte {
  return ram.Data[addr]
}

func (ram *RandomAccessMemory) Write(addr uint16, val byte) {
  ram.Data[addr] = val
}

func (rom *ReadOnlyMemory) Read(addr uint16) byte {
  return rom.Data[addr]
}

func (rom *ReadOnlyMemory) Write(addr uint16, val byte) {
  panic("Cannot write to read-only memory")  
}

func NewRandomAccessMemory(size uint16) *RandomAccessMemory {
  return &RandomAccessMemory{make([]byte, size)}
}

func NewReadOnlyMemory(data []byte) *ReadOnlyMemory {
  return &ReadOnlyMemory{data}
}
