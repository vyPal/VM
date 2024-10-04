.DATA
  list 1 {1, 9, 3, 7}
.TEXT
  LD R0 [list]
  LD R1 0xFFFFF000
  ADD R0 48
  ST [R1] R0B
  HLT
