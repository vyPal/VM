.DATA
  filename DB "file.bin", 0
.TEXT
  OPEN R1 [filename]
  LD R2 0x1
  WRITE R1 R2 1
  CLOSE R1
  HLT
