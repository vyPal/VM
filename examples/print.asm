; Requires compiling together with std.asm

.DATA
  string DB "Hello, World!\n", 0
.TEXT
  LD R1 string
  CALL [print]
end:
  HLT
