.DATA
  list 1 {1, 9, 3, 7}
  len 1 4
.TEXT
LOOP:
  LD R0 [R1+list]
  ADD R0 48
  ADD R1 1
  CMP R1 [len]
  JNE [LOOP]
  HLT
