package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Bytecode struct {
	MagicNumber uint32
	SectorCount uint8
	StartAddress uint32
	Sectors     []BCSector
}

type BCSector struct {
	StartAddress uint32
	Length       uint32
	Bytecode		 []byte
}

func NewBytecode(magicNumber uint32) *Bytecode {
	return &Bytecode{
		MagicNumber: magicNumber,
		SectorCount: 0,
		Sectors:     []BCSector{},
	}
}

func ProgramToBytecode(program *Parser) *Bytecode {
	bytecode := NewBytecode(0x736F6265)
	firstSectorWithInstructions := -1

	for _, sector := range program.Sectors {
		if len(sector.Instructions) > 0 && firstSectorWithInstructions == -1 {
			firstSectorWithInstructions = len(bytecode.Sectors)
			bytecode.StartAddress = sector.BaseAddress
		}
		bcSector := BCSector{
			StartAddress: sector.BaseAddress,
			Length:       uint32(len(sector.Program)),
			Bytecode:     sector.Program,
		}
		bytecode.Sectors = append(bytecode.Sectors, bcSector)
		bytecode.SectorCount++
	}

	return bytecode
}

func EncodeBytecode(bc *Bytecode) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, bc.MagicNumber)
	if err != nil {
		return nil, err
	}

	err = buffer.WriteByte(bc.SectorCount)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, bc.StartAddress)
	if err != nil {
		return nil, err
	}

	for _, sector := range bc.Sectors {
		err := binary.Write(buffer, binary.LittleEndian, sector.StartAddress)
		if err != nil {
			return nil, err
		}

		err = binary.Write(buffer, binary.LittleEndian, sector.Length)
		if err != nil {
			return nil, err
		}

		_, err = buffer.Write(sector.Bytecode)
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func DecodeBytecode(data []byte) (*Bytecode, error) {
	buffer := bytes.NewReader(data)
	bc := &Bytecode{}

	err := binary.Read(buffer, binary.LittleEndian, &bc.MagicNumber)
	if err != nil {
		return nil, err
	}

	if bc.MagicNumber != 0x736F6265 {
		return nil, fmt.Errorf("Invalid magic number")
	}

	sectorCount, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	bc.SectorCount = sectorCount

	err = binary.Read(buffer, binary.LittleEndian, &bc.StartAddress)
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(bc.SectorCount); i++ {
		sector := BCSector{}

		err := binary.Read(buffer, binary.LittleEndian, &sector.StartAddress)
		if err != nil {
			return nil, err
		}

		err = binary.Read(buffer, binary.LittleEndian, &sector.Length)
		if err != nil {
			return nil, err
		}

		sector.Bytecode = make([]byte, sector.Length)
		_, err = buffer.Read(sector.Bytecode)
		if err != nil {
			return nil, err
		}

		bc.Sectors = append(bc.Sectors, sector)
	}

	return bc, nil
}


