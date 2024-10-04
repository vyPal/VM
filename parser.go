package main

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

type Parser struct {
  Filename string
  Contents string
  BaseAddress uint32
  Labels map[string]uint32
  Data []*Data
  Instructions []*Instruction
  Program []byte
  PostParse []func()

  CurrentSection string
}

type Data struct {
  Address uint32
  Value uint32
  Size uint32
  Name string
}

func (p *Parser) DataByName(name string) (*Data, bool) {
  for _, data := range p.Data {
    if name == data.Name {
      return data, true
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
  p.Data = []*Data{}
  p.Instructions = []*Instruction{}
  p.Program = []byte{}
  p.PostParse = []func(){}
  for _, line := range strings.Split(p.Contents, "\n") {
    p.ParseLine(line)
  }
  for _, data := range p.Data {
    data.Address = p.BaseAddress + uint32(len(p.Program))
    p.Program = append(p.Program, EncodeData(data.Value, data.Size)...)
  }
  for _, f := range p.PostParse {
    f()
  }
  p.Program = []byte{}
  for _, instr := range p.Instructions {
    p.Program = append(p.Program, EncodeInstruction(instr)...)
  }
  for _, data := range p.Data {
    p.Program = append(p.Program, EncodeData(data.Value, data.Size)...)
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
  p.Labels[label] = p.BaseAddress + uint32(len(p.Program))
}

func (p *Parser) ParseData(line string) {
  parts := strings.Split(line, " ")
  name := parts[0]
  size, err := strconv.ParseUint(parts[1], 0, 32)
  if err != nil {
    panic("Invalid size for data: " + parts[1])
  }
  value, err := strconv.ParseUint(parts[2], 0, 32)
  if err != nil {
    panic("Invalid value for data: " + parts[2])
  }
  p.Data = append(p.Data, &Data{
    Name: name,
    Size: uint32(size),
    Value: uint32(value),
  })
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
  p.Instructions = append(p.Instructions, instruction)
  p.Program = append(p.Program, EncodeInstruction(instruction)...)
}

func (p *Parser) ParseOperand(arg string, operand *Operand, opName string) {
  var detectedType OperandType
  if arg[0] == '[' {
    if arg[1] == '[' {
      detectedType = IMem
      toParse := strings.TrimSuffix(arg[2:], "]]")
      parsedValue, err := strconv.ParseUint(toParse, 0, 32)
      if err != nil {
        operand.Value = &IMemOperand{}
        p.PostParse = append(p.PostParse, func() {
          label := toParse
          if val, ok := p.Labels[toParse]; ok {
            operand.Value.(*IMemOperand).Addr = val
          } else if val, ok := p.DataByName(toParse); ok {
            operand.Value.(*IMemOperand).Addr = val.Value
          } else {
            panic("Unknown label: " + label)
          }
        })
      } else {
        operand.Value = &IMemOperand{Addr: uint32(parsedValue)}
      }
    } else {
      detectedType = DMem
      toParse := strings.TrimSuffix(arg[1:], "]")
      parsedValue, err := strconv.ParseUint(toParse, 0, 32)
      if err != nil {
        operand.Value = &DMemOperand{}
        p.PostParse = append(p.PostParse, func() {
          label := toParse
          if val, ok := p.Labels[toParse]; ok {
            operand.Value.(*DMemOperand).Addr = val
          } else if val, ok := p.DataByName(toParse); ok {
            operand.Value.(*DMemOperand).Addr = val.Value
          } else {
            panic("Unknown label: " + label)
          }
        })
      } else {
        operand.Value = &DMemOperand{Addr: uint32(parsedValue)}
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
