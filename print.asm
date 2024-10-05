.DATA
  string DB "Hello, World!"
.TEXT
print_loop:
  LD R0B [R1 + string]
  CMP R0B 0
  JEQ [end]
  ST [R1 + 0xFFFFF000] R0B
  ADD R1 1
  JMP [print_loop]
end:
  HLT
