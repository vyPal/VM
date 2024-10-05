.DATA
  list 1 {1, 2, 3, 4}
  list2 1 {5, 6, 7, 8}
  len 1 4
.TEXT
LOOP:
  LD R0B [R1+list]
  ADD R0B 48
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1
  CMP R1 [len]
  JNE [LOOP]
  JMP [RAMPROG]

ORG 0x00000000
RAMPROG:
  LD R1 0
RLP:
  LD R0B [R1+list2]
  ADD R0B 48
  ST [R1+0xFFFFF004] R0B
  ADD R1 1
  CMP R1 [len]
  JNE [RLP]
  HLT
