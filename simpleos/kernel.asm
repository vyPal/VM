.DATA
  loading DB "Loading system...", 0
.TEXT
ORG 0x80000000
  LD R1 loading
  CALL [print]
  HLT
