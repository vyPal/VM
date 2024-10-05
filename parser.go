package main

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

type Parser struct {
	Filename           string
	Contents           string
	DefaultBaseAddress uint32
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

func (p *Parser) Parse() {
	contents, err := os.ReadFile(p.Filename)
	if err != nil {
		panic(err)
	}
	p.Contents = string(contents)
	p.Labels = make(map[string]uint32)
	p.Sectors = []*Sector{}
	p.CurrentSector = &Sector{BaseAddress: p.DefaultBaseAddress}
	p.Sectors = append(p.Sectors, p.CurrentSector)
	for _, line := range strings.Split(p.Contents, "\n") {
		p.ParseLine(line)
	}

	for i := 0; i < len(p.Sectors); i++ {
		if len(p.Sectors[i].Instructions) == 0 && len(p.Sectors[i].Data) == 0 {
			p.Sectors = append(p.Sectors[:i], p.Sectors[i+1:]...)
			i--
		}
	}

	for _, sector := range p.Sectors {
		for _, data := range sector.Data {
			data.Address = sector.BaseAddress + uint32(len(sector.Program))
			sector.Program = append(sector.Program, EncodeData(data.Value, data.Size)...)
		}
	}
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
}

func (p *Parser) ParseData(line string) {
	parts := strings.Split(line, " ")
	name := parts[0]
	size, err := strconv.ParseUint(parts[1], 0, 32)
	if err != nil {
		panic("Invalid size for data: " + parts[1])
	}

	rest := strings.Join(parts[2:], " ")

	if strings.HasPrefix(rest, "{") && strings.HasSuffix(rest, "}") {
		values := strings.Trim(rest, "{}")
		valueParts := strings.Split(values, ",")
		var valueArray []uint32
		for _, v := range valueParts {
			if strings.TrimSpace(v) == "" {
				continue
			}
			value, err := strconv.ParseUint(strings.TrimSpace(v), 0, 32)
			if err != nil {
				panic("Invalid value in array: " + v)
			}
			valueArray = append(valueArray, uint32(value))
		}

		for _, v := range valueArray {
			p.CurrentSector.Data = append(p.CurrentSector.Data, &Data{
				Name:  name,
				Size:  uint32(size),
				Value: v,
			})
		}
	} else {
		value, err := strconv.ParseUint(rest, 0, 32)
		if err != nil {
			panic("Invalid value for data: " + rest)
		}
		p.CurrentSector.Data = append(p.CurrentSector.Data, &Data{
			Name:  name,
			Size:  uint32(size),
			Value: uint32(value),
		})
	}
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
		if arg[0] == ';' {
			args = args[:i]
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
			panic(arg + " is not a valid immediate value for " + opName + ": " + err.Error())
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
