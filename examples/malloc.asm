.TEXT
  ; Allocate 32 bits
  MALLOC 32 R0
  ; Store something to those bits
  LD R1 420
  ST [R0] R1
  ; Load it into another register
  LD R2 [R0]
  ; Free the memory
  FREE R0 32
  ; Halt
  HLT
