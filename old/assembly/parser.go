package assembly

import (
	"strconv"
	"strings"

	"src.vypal.me/vyPal/VM/cpu"
)

type Data struct {
	name string
	value byte
}

func ParseString(s string) cpu.Program {
	lines := strings.Split(s, "\n")
	currentSection := ""
	vals := []Data{}
	instr := []cpu.Instruction{}
	labels := map[string]uint16{}

	afterDone := []func(){}

	consts := &cpu.ListDataBlock{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ".") {
			if currentSection == ".text" && line == ".data" {
				panic("Data section must be before text section")
			}
			currentSection = line
		} else if strings.HasPrefix(line, "#") {
			// Comment
			continue
		} else {
			if currentSection == "" {
				panic("No section defined")
			} else if currentSection == ".data" {
				data := strings.Split(line, " ")
				if len(data) != 2 {
					panic("Invalid data")
				}
				i, err := strconv.ParseInt(data[1], 0, 32)
				if err != nil {
					panic("Invalid data")
				}
				vals = append(vals, Data{name: data[0], value: byte(i)})
				consts.Data = append(consts.Data, byte(i))
			} else if currentSection == ".text" {
				data := strings.Split(line, " ")
				// INST PARAM...
				if len(data) < 1 {
					panic("Invalid instruction")
				}
				inst := data[0]
				params := data[1:]

				if strings.HasPrefix(inst, "[") && strings.HasSuffix(inst, "]") {
					label := inst[1:len(inst)-1]
					tl := 0
					for _, v := range instr {
						tl += len(v.Encode())
					}
					labels[label] = uint16(0x8000 + 0x3 + len(consts.Encode()) + tl)
					continue
				}
				switch inst {
				case "AND":
					if len(params) != 2 {
						panic("Invalid AND")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid AND")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid AND")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.AND{Register1: &r1, Register2: &r2})
				case "OR":
					if len(params) != 2 {
						panic("Invalid OR")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid OR")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid OR")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.OR{Register1: &r1, Register2: &r2})
				case "XOR":
					if len(params) != 2 {
						panic("Invalid XOR")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid XOR")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid XOR")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.XOR{Register1: &r1, Register2: &r2})
				case "NOT":
					if len(params) != 1 {
						panic("Invalid NOT")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid NOT")
					}
					r1 := cpu.Register{ID: byte(i1)}
					instr = append(instr, &cpu.NOT{Register: &r1})
				case "SHL":
					if len(params) != 1 {
						panic("Invalid SHL")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid SHL")
					}
					r1 := cpu.Register{ID: byte(i1)}
					instr = append(instr, &cpu.SHL{Register: &r1})
				case "SHR":
					if len(params) != 1 {
						panic("Invalid SHR")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid SHR")
					}
					r1 := cpu.Register{ID: byte(i1)}
					instr = append(instr, &cpu.SHR{Register: &r1})
				case "ADD":
					if len(params) != 2 {
						panic("Invalid ADD")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid ADD")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid ADD")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.ADD{Register1: &r1, Register2: &r2})
				case "SUB":
					if len(params) != 2 {
						panic("Invalid SUB")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid SUB")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid SUB")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.SUB{Register1: &r1, Register2: &r2})
				case "MUL":
					if len(params) != 2 {
						panic("Invalid MUL")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid MUL")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid MUL")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.MUL{Register1: &r1, Register2: &r2})
				case "DIV":
					if len(params) != 2 {
						panic("Invalid DIV")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid DIV")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid DIV")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.DIV{Register1: &r1, Register2: &r2})
				case "MOD":
					if len(params) != 2 {
						panic("Invalid MOD")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid MOD")
					}
					i2, err := strconv.Atoi(params[1])
					if err != nil {
						panic("Invalid MOD")
					}
					r1 := cpu.Register{ID: byte(i1)}
					r2 := cpu.Register{ID: byte(i2)}
					instr = append(instr, &cpu.MOD{Register1: &r1, Register2: &r2})
				case "HLT":
					instr = append(instr, &cpu.HLT{})
				case "LD":
					if len(params) != 2 {
						panic("Invalid LD")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid LD")
					}
					if strings.HasPrefix(params[1], "0x") {
						i2, err := strconv.ParseInt(params[1][2:], 16, 32)
						if err != nil {
							panic("Invalid LD")
						}
						r1 := cpu.Register{ID: byte(i1)}
						instr = append(instr, &cpu.LD{Register: &r1, Address: uint16(i2)})
					} else {
						r1 := cpu.Register{ID: byte(i1)}
						index := -1
						for i, v := range vals {
							if v.name == params[1] {
								index = i
								break
							}
						}
						if index == -1 {
							panic("Invalid LD")
						}
						instr = append(instr, &cpu.LD{Register: &r1, Address: consts.GetAddr(index)})
					}
				case "ST":
					if len(params) != 2 {
						panic("Invalid ST")
					}
					i1, err := strconv.Atoi(params[0])
					if err != nil {
						panic("Invalid ST")
					}
					if strings.HasPrefix(params[1], "0x") {
						i2, err := strconv.ParseInt(params[1][2:], 16, 32)
						if err != nil {
							panic("Invalid ST")
						}
						r1 := cpu.Register{ID: byte(i1)}
						instr = append(instr, &cpu.ST{Register: &r1, Address: uint16(i2)})
					} else {
						r1 := cpu.Register{ID: byte(i1)}
						index := -1
						for i, v := range vals {
							if v.name == params[1] {
								index = i
								break
							}
						}
						if index == -1 {
							panic("Invalid ST")
						}
						instr = append(instr, &cpu.ST{Register: &r1, Address: consts.GetAddr(index)})
					}
				case "JMP":
					if len(params) != 1 {
						panic("Invalid JMP")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JMP{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JMP{Address: uint16(i1)})
					}
				case "JZ":
					if len(params) != 1 {
						panic("Invalid JZ")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JZ{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JZ{Address: uint16(i1)})
					}
				case "JNZ":
					if len(params) != 1 {
						panic("Invalid JNZ")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JNZ{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JNZ{Address: uint16(i1)})
					}
				case "JG":
					if len(params) != 1 {
						panic("Invalid JG")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JG{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JG{Address: uint16(i1)})
					}
				case "JGE":
					if len(params) != 1 {
						panic("Invalid JGE")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JGE{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JGE{Address: uint16(i1)})
					}
				case "JL":
					if len(params) != 1 {
						panic("Invalid JL")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JL{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JL{Address: uint16(i1)})
					}
				case "JLE":
					if len(params) != 1 {
						panic("Invalid JLE")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.JLE{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.JLE{Address: uint16(i1)})
					}
				case "CALL":
					if len(params) != 1 {
						panic("Invalid CALL")
					}
					i1, err := strconv.ParseInt(params[0], 0, 32)
					if err != nil {
						inst := &cpu.CALL{}
						afterDone = append(afterDone, func() {
							inst.Address = labels[params[0]]
						})
						instr = append(instr, inst)
					} else {
						instr = append(instr, &cpu.CALL{Address: uint16(i1)})
					}
				case "RET":
					instr = append(instr, &cpu.RET{})
				default:
					panic("Unknown instruction "+inst)
			  }
			} else {
				panic("Unknown section")
			}
		}
	}

	for _, v := range afterDone {
		v()
	}

	return cpu.Program{
		DataBlock: consts,
		Instructions: instr,
	}
}
