package opcodes

type OPCode byte

const (
  LD OPCode = iota
  ST
  ADD
  SUB
  MUL
  DIV
  MOD
  AND
  OR
  XOR
  NOT
  SHL
  SHR
  JMP
  JZ
  JNZ
  JG
  JGE
  JL
  JLE
  CALL
  RET
  HLT
)
