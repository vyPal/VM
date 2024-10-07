; Original code by @vh8t
.DATA
  string DB "Hello, World!\n", 0
.TEXT
; R1 - arg1 string
; R2 - char
; R3 - counter
  LD R1 string
  CALL [strlen]
  HLT
strlen:
  LD R3 0
strlen_counter:
  LD R2B [R1]
  CMP R2B 0
  JEQ [strlen_end]
  ADD R3 1
  ADD R1 1
  JMP [strlen_counter]
strlen_end:
  LD R0 R3
  RET
