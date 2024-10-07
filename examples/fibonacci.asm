.TEXT
  LD R0 0
  LD R1 0
  LD R3 0
  LD R4 1
  LD R5 10
  CALL [PRINTR0B]
  LD R0 R4
  CALL [PRINTR0B]

LOOP:
  LD R0 R3
  ADD R0 R4
  LD R3 R4
  LD R4 R0
  CALL [PRINTR0B]

  CMP R5 0
  SUB R5 1
  JNE [LOOP]
  HLT

PRINTR0B:
  CMP R0 0
  JEQ [PRINTR0_PRINT_ZERO]

  ST [0x00000000] R0
  DIV R0 1000000000
  CMP R0 0
  JEQ [PRINTR0_LOOP]
  ADD R0 48
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1

  LD R0 [0x00000000]
  MOD R0 1000000000

PRINTR0_LOOP:
  LD R2 100000000
PRINTR0_NEXT_DIGIT:
  DIV R0 R2
  CMP R0 0
  JEQ [PRINTR0_SKIP_DIGIT]
  ADD R0 48
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1

PRINTR0_SKIP_DIGIT:
  LD R0 [0x00000000]
  MOD R0 R2
  DIV R2 10
  CMP R2 0
  JNE [PRINTR0_NEXT_DIGIT]
  JMP [PRINT_SPACE]

PRINTR0_PRINT_ZERO:
  LD R0 48
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1

PRINT_SPACE:
  LD R0 32
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1
  RET