.DATA
  prompt DB " >", 0
  not_found DB "Binary not found.", 0

.TEXT
  MALLOC 0x400 R3
  LD R4 0
start:
  HANDLE 0x1 keyboard_event
  CALL [clear]
  LD R1 prompt
  LD R2 0
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

failed_to_load:
  LD R1 not_found
  LD R2 41
  CALL [print_loop]
  POP R3
  POP R2
  POP R1
  JMP [start]

exec:
  PUSH R1
  PUSH R2
  PUSH R3
  OPEN R1 [R3]
  CMP R1 0xFFFFFFFF
  JEQ [failed_to_load]
  LOADBIN R1 R2
  CLOSE R1
  CALL [R2]
  POP R3
  POP R2
  POP R1
  CALL [clear_buf]
  JMP [start]

clear:
  LD R0 0
  LD R1 0
clear_loop:
  CMP R0 0x3A
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

clear_buf:
  LD R5 R4
  ADD R5 R3
  LD [R5] 0
  CMP R4 0
  JEQ [return]
  DEC R4
  JMP [clear_buf]

keyboard_event:
  AND R10 0xFF
  CMP R10 10
  JEQ [exec]
  LD R5 R4
  ADD R5 R3
  ST [R5] R10B
  INC R4
  RET
