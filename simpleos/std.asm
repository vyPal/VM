.TEXT
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
