.DATA
  text DB "Nejaky program", 0
.TEXT
  LD R5 0x69
  RTA text R1
  INT 0x2
  RET
