package main

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Parser struct {
	DefaultBaseAddress uint32
	ExplicitStart      bool
	StartAddress       uint32
	Labels             map[string]uint32
	Sectors            []*Sector

	CurrentSection string
	CurrentSector  *Sector
}

type Sector struct {
	BaseAddress  uint32
	Instructions []*Instruction
	Data         []*Data
	Program      []byte
	PostParse    []func()
}

type Data struct {
	Address uint32
	Value   uint32
	Size    uint32
	Name    string
}

func (p *Parser) DataByName(name string) (*Data, bool) {
	for _, sector := range p.Sectors {
		for _, data := range sector.Data {
			if name == data.Name {
				return data, true
			}
		}
	}
	return nil, false
}

func NewParser() *Parser {
	return &Parser{
		DefaultBaseAddress: 0,
		Labels:             make(map[string]uint32),
		Sectors:            []*Sector{},
	}
}

func (p *Parser) AddFile(filename string) {
	if p.CurrentSector == nil || len(p.CurrentSector.Instructions) > 0 || len(p.CurrentSector.Data) > 0 {
		p.CurrentSector = &Sector{BaseAddress: p.DefaultBaseAddress}
		p.Sectors = append(p.Sectors, p.CurrentSector)
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	sectorsToEncode := []*Sector{p.CurrentSector}
	lastSector := p.CurrentSector
	for _, line := range strings.Split(string(contents), "\n") {
		p.ParseLine(line)
		if p.CurrentSector != lastSector {
			lastSector = p.CurrentSector
			sectorsToEncode = append(sectorsToEncode, lastSector)
		}
	}

	for _, sector := range sectorsToEncode {
		for _, data := range sector.Data {
			data.Address = sector.BaseAddress + uint32(len(sector.Program))
			sector.Program = append(sector.Program, EncodeData(data.Value, data.Size)...)
		}
	}

	p.UpdateDefaultBaseAddress()
}

func (p *Parser) UpdateDefaultBaseAddress() {
	var lastRomEnd uint32 = 0x00000000

	for _, sector := range p.Sectors {
		if sector.BaseAddress >= 0x00000000 {
			sectorEnd := sector.BaseAddress + uint32(len(sector.Program))
			if sectorEnd > lastRomEnd {
				lastRomEnd = sectorEnd
			}
		}
	}

	p.DefaultBaseAddress = lastRomEnd + 1
}

func (p *Parser) CheckForOverlappingSectors() error {
	sort.Slice(p.Sectors, func(i, j int) bool {
		return p.Sectors[i].BaseAddress < p.Sectors[j].BaseAddress
	})

	for i := 1; i < len(p.Sectors); i++ {
		prevSector := p.Sectors[i-1]
		currentSector := p.Sectors[i]

		prevEnd := prevSector.BaseAddress + uint32(len(prevSector.Program))

		if currentSector.BaseAddress < prevEnd {
			return fmt.Errorf(
				"Overlap detected between sector starting at 0x%08X and sector starting at 0x%08X",
				prevSector.BaseAddress, currentSector.BaseAddress,
			)
		}
	}

	return nil
}

func (p *Parser) Parse() {
	for _, sector := range p.Sectors {
		for _, postParse := range sector.PostParse {
			postParse()
		}
	}
	for _, sector := range p.Sectors {
		sector.Program = []byte{}
		for _, instruction := range sector.Instructions {
			sector.Program = append(sector.Program, EncodeInstruction(instruction)...)
		}
		for _, data := range sector.Data {
			sector.Program = append(sector.Program, EncodeData(data.Value, data.Size)...)
		}
	}
}

func (p *Parser) ParseLine(line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}
	if line[0] == ';' {
		return
	}
	if line[0] == '.' {
		p.ParseSection(line)
		return
	}
	if line[len(line)-1] == ':' {
		p.ParseLabel(line)
		return
	}
	p.ParseInstruction(line)
}

func (p *Parser) ParseSection(line string) {
	p.CurrentSection = line[1:]
}

func (p *Parser) ParseLabel(line string) {
	label := line[:len(line)-1]
	p.Labels[label] = p.CurrentSector.BaseAddress + uint32(len(p.CurrentSector.Program))
	if label == "_start" {
		if p.ExplicitStart {
			panic("Multiple _start labels found")
		}
		p.ExplicitStart = true
		p.StartAddress = p.CurrentSector.BaseAddress + uint32(len(p.CurrentSector.Program))
	}
}

func (p *Parser) ParseData(line string) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		panic("Invalid data declaration")
	}

	name := parts[0]
	directive := parts[1] // DB, DW, DD
	valueStr := strings.Join(parts[2:], " ")

	switch directive {
	case "DB":
		p.parseByteData(name, valueStr)
	case "DW":
		p.parseWordData(name, valueStr)
	case "DD":
		p.parseDwordData(name, valueStr)
	default:
		panic("Unknown data directive: " + directive)
	}
}

func (p *Parser) parseByteData(name, valueStr string) {
	values := parseMixedValues(valueStr)
	for _, val := range values {
		p.CurrentSector.Data = append(p.CurrentSector.Data, &Data{
			Name:  name,
			Size:  1,
			Value: uint32(val),
		})
	}
}

func parseMixedValues(valueStr string) []uint32 {
	var valueArray []uint32
	var currentBuffer strings.Builder
	inString := false
	escape := false

	for i := 0; i < len(valueStr); i++ {
		char := valueStr[i]

		if escape {
			// Handle special escape sequences like \n, \r, \t, etc.
			switch char {
			case 'n':
				currentBuffer.WriteByte('\n')
			case 'r':
				currentBuffer.WriteByte('\r')
			case 't':
				currentBuffer.WriteByte('\t')
			case '\\':
				currentBuffer.WriteByte('\\')
			default:
				currentBuffer.WriteByte(char)
			}
			escape = false
		} else if char == '\\' {
			escape = true
		} else if char == '"' && !escape {
			inString = !inString
			currentBuffer.WriteByte(char)
		} else if char == ',' && !inString {
			valueArray = append(valueArray, parseSingleValue(currentBuffer.String())...)
			currentBuffer.Reset()
		} else {
			currentBuffer.WriteByte(char)
		}
	}

	if currentBuffer.Len() > 0 {
		valueArray = append(valueArray, parseSingleValue(currentBuffer.String())...)
	}

	return valueArray
}

func parseSingleValue(valueStr string) []uint32 {
	valueStr = strings.TrimSpace(valueStr)

	if isStringLiteral(valueStr) {
		str := parseStringLiteral(valueStr)
		var result []uint32
		for _, char := range str {
			result = append(result, uint32(char))
		}
		return result
	} else if valueStr != "" {
		value, err := strconv.ParseUint(valueStr, 0, 32)
		if err != nil {
			panic("Invalid value in mixed data: " + valueStr)
		}
		return []uint32{uint32(value)}
	}
	return nil
}

func (p *Parser) parseWordData(name, valueStr string) {
	values := parseValues(valueStr)
	for _, val := range values {
		p.CurrentSector.Data = append(p.CurrentSector.Data, &Data{
			Name:  name,
			Size:  2,
			Value: uint32(val),
		})
	}
}

func (p *Parser) parseDwordData(name, valueStr string) {
	values := parseValues(valueStr)
	for _, val := range values {
		p.CurrentSector.Data = append(p.CurrentSector.Data, &Data{
			Name:  name,
			Size:  4,
			Value: uint32(val),
		})
	}
}

func parseValues(valueStr string) []uint32 {
	valueParts := strings.Split(valueStr, ",")
	var valueArray []uint32
	for _, v := range valueParts {
		if strings.TrimSpace(v) == "" {
			continue
		}
		value, err := strconv.ParseUint(strings.TrimSpace(v), 0, 32)
		if err != nil {
			panic("Invalid value: " + v)
		}
		valueArray = append(valueArray, uint32(value))
	}
	return valueArray
}

func isStringLiteral(valueStr string) bool {
	return strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")
}

func parseStringLiteral(valueStr string) string {
	return valueStr[1 : len(valueStr)-1]
}

func EncodeData(value, size uint32) []byte {
	data := make([]byte, size)
	for i := uint32(0); i < size; i++ {
		data[i] = byte(value >> (i * 8))
	}
	return data
}

func (p *Parser) ParseInstruction(line string) {
	if p.CurrentSection == "data" || p.CurrentSection == "DATA" {
		p.ParseData(line)
		return
	} else if p.CurrentSection != "text" && p.CurrentSection != "TEXT" {
		panic("Unknown section: " + p.CurrentSection)
	}
	parts := strings.Split(line, " ")
	opcode := parts[0]
	args := parts[1:]
	if opcode == "ORG" {
		value, err := strconv.ParseUint(args[0], 0, 32)
		if err != nil {
			panic("Invalid value for ORG: " + args[0])
		}
		p.CurrentSector = &Sector{BaseAddress: uint32(value)}
		p.Sectors = append(p.Sectors, p.CurrentSector)
		return
	}
	instruction := GetInstruction(opcode)
	if instruction == nil {
		panic("Unknown instruction: " + opcode)
	}
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg[0] == ';' {
			args = args[:i]
		} else if arg == "" {
			args = args[:i]
		} else if arg[0] == '[' {
			for j := i + 1; j < len(args); j++ {
				if strings.Contains(args[j], "]") {
					oldArgs := args
					args = append(args[:i], strings.Join(args[i:j+1], " "))
					if j+1 < len(oldArgs) {
						args = append(args, oldArgs[j+1:]...)
					}
					break
				}
			}
		}
	}
	if len(args) != len(instruction.Operands) {
		panic("Invalid number of arguments for " + opcode)
	}
	for i, arg := range args {
		p.ParseOperand(arg, &instruction.Operands[i], instruction.Name)
	}
	p.CurrentSector.Instructions = append(p.CurrentSector.Instructions, instruction)
	p.CurrentSector.Program = append(p.CurrentSector.Program, EncodeInstruction(instruction)...)
}

func getRegisterID(name string) (byte, error) {
	id := strings.TrimSuffix(strings.TrimSuffix(name[1:], "B"), "L")
	parsedValue, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return 0, err
	}
	return byte(parsedValue), nil
}

func (p *Parser) ParseOperand(arg string, operand *Operand, opName string) {
	var detectedType OperandType
	if arg[0] == '[' {
		if arg[1] == '[' {
			detectedType = IMem
			toParse := strings.TrimSuffix(arg[2:], "]]")

			// Check if it's a register offset or just a raw address
			if strings.Contains(toParse, "+") {
				parts := strings.Split(toParse, "+")
				registerName := strings.TrimSpace(parts[0])
				offsetStr := strings.TrimSpace(parts[1])
				offset, err := strconv.ParseUint(offsetStr, 0, 32)
				if err != nil {
					p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
						label := toParse
						if val, ok := p.Labels[offsetStr]; ok {
							operand.Value.(*IMemOperand).Addr = val
						} else if val, ok := p.DataByName(offsetStr); ok {
							operand.Value.(*IMemOperand).Addr = val.Address
						} else {
							panic("Unknown label: " + label)
						}
					})
				}

				rid, err := getRegisterID(registerName)
				if err != nil {
					panic("Invalid register for " + opName + ": " + registerName)
				}
				operand.Value = &IMemOperand{
					Type:     Offset,
					Addr:     uint32(offset),
					Register: rid, // Implement this function to map register names to IDs
				}
			} else {
				parsedValue, err := strconv.ParseUint(toParse, 0, 32)
				if err != nil {
					if toParse[0] == 'r' || toParse[0] == 'R' {
						rid, err := getRegisterID(toParse)
						if err != nil {
							operand.Value = &IMemOperand{Type: Address}
							p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
								label := toParse
								if val, ok := p.Labels[toParse]; ok {
									operand.Value.(*IMemOperand).Addr = val
								} else if val, ok := p.DataByName(toParse); ok {
									operand.Value.(*IMemOperand).Addr = val.Address
								} else {
									panic("Unknown label: " + label)
								}
							})
						} else {
							operand.Value = &IMemOperand{Type: Register, Register: rid}
						}
					} else {
						operand.Value = &IMemOperand{Type: Address}
						p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
							label := toParse
							if val, ok := p.Labels[toParse]; ok {
								operand.Value.(*IMemOperand).Addr = val
							} else if val, ok := p.DataByName(toParse); ok {
								operand.Value.(*IMemOperand).Addr = val.Address
							} else {
								panic("Unknown label: " + label)
							}
						})
					}
				} else {
					operand.Value = &IMemOperand{Type: Address, Addr: uint32(parsedValue)}
				}
			}
		} else {
			detectedType = DMem
			toParse := strings.TrimSuffix(arg[1:], "]")

			// Check if it's a register offset or just a raw address
			if strings.Contains(toParse, "+") {
				parts := strings.Split(toParse, "+")
				registerName := strings.TrimSpace(parts[0])
				offsetStr := strings.TrimSpace(parts[1])
				offset, err := strconv.ParseUint(offsetStr, 0, 32)
				if err != nil {
					p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
						label := toParse
						if val, ok := p.Labels[offsetStr]; ok {
							operand.Value.(*DMemOperand).Addr = val
						} else if val, ok := p.DataByName(offsetStr); ok {
							operand.Value.(*DMemOperand).Addr = val.Address
						} else {
							panic("Unknown label: " + label)
						}
					})
				}

				rid, err := getRegisterID(registerName)
				if err != nil {
					panic("Invalid register for " + opName + ": " + registerName)
				}
				operand.Value = &DMemOperand{
					Type:     Offset,
					Addr:     uint32(offset),
					Register: rid, // Implement this function to map register names to IDs
				}
			} else {
				parsedValue, err := strconv.ParseUint(toParse, 0, 32)
				if err != nil {
					if toParse[0] == 'r' || toParse[0] == 'R' {
						rid, err := getRegisterID(toParse)
						if err != nil {
							operand.Value = &DMemOperand{Type: Address}
							p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
								label := toParse
								if val, ok := p.Labels[toParse]; ok {
									operand.Value.(*DMemOperand).Addr = val
								} else if val, ok := p.DataByName(toParse); ok {
									operand.Value.(*DMemOperand).Addr = val.Address
								} else {
									panic("Unknown label: " + label)
								}
							})
						} else {
							operand.Value = &DMemOperand{Type: Register, Register: rid}
						}
					} else {
						operand.Value = &DMemOperand{Type: Address}
						p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
							label := toParse
							if val, ok := p.Labels[toParse]; ok {
								operand.Value.(*DMemOperand).Addr = val
							} else if val, ok := p.DataByName(toParse); ok {
								operand.Value.(*DMemOperand).Addr = val.Address
							} else {
								panic("Unknown label: " + label)
							}
						})
					}
				} else {
					operand.Value = &DMemOperand{Type: Address, Addr: uint32(parsedValue)}
				}
			}
		}
	} else if arg[0] == 'r' || arg[0] == 'R' {
		detectedType = Reg
		id := strings.TrimSuffix(strings.TrimSuffix(arg[1:], "B"), "L")
		parsedValue, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			panic(arg + " is not a valid register for " + opName + ": " + err.Error())
		}
		size := 0x0
		if strings.HasSuffix(arg, "B") {
			size = 0x2
		} else if strings.HasSuffix(arg, "L") {
			size = 0x1
		}
		operand.Value = &RegOperand{RegNum: byte(parsedValue), Size: byte(size)}
	} else {
		detectedType = Imm
		parsedValue, err := strconv.ParseUint(arg, 0, 32)
		if err != nil {
			operand.Value = &ImmOperand{}
			p.CurrentSector.PostParse = append(p.CurrentSector.PostParse, func() {
				if val, ok := p.Labels[arg]; ok {
					operand.Value.(*ImmOperand).Value = val
				} else if val, ok := p.DataByName(arg); ok {
					operand.Value.(*ImmOperand).Value = val.Address
				} else {
					panic("Unknown label: " + arg)
				}
			})
		}
		operand.Value = &ImmOperand{Value: uint32(parsedValue)}
	}
	if len(operand.AllowedTypes) > 0 {
		if !slices.Contains(operand.AllowedTypes, detectedType) {
			panic(arg + " is not an allowed type for " + opName)
		}
		operand.Type = detectedType
	} else {
		if operand.Type != detectedType {
			panic(arg + " is not a valid type for " + opName)
		}
	}
}
