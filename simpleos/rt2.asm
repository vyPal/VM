.DATA
  prompt DB " >", 0

.TEXT
  MALLOC 0x400 R3
  LD R4 0
  HANDLE 1 keyboard_event
  CALL [clear]
  LD R1 prompt
  CALL [print]
  LD R6 R4
loop:
  CMP R6 R4
  JEQ [loop]
  LD R1 R3
  LD R2 2
  CALL [print_loop]
  LD R6 R4
  JMP [loop]

clear:
  LD R0 0
  LD R1 0
clear_loop:
  CMP R0 0x12
  JEQ [return]
  ST [R0 + 0xFFFFF000] R1B
  INC R0
  JMP [clear_loop]

print:
  LD R2 0
print_loop:
  LD R0B [R1]
  CMP R0B 0
  JEQ [return]
  ST [R2 + 0xFFFFF000] R0B
  ADD R1 1
  ADD R2 1
  JMP [print_loop]
return:
  RET

keyboard_event:
  AND R15 0xFF
  LD R5 R4
  ADD R5 R3
  ST [R5] R15B
  LD R12 R15B
  INC R4
  RET
