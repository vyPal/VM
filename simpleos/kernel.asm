.DATA
  loading DB "Loading system...", 0
  ramprogram DB "runtime.bin", 0
  rt2 DB "rt2.bin", 0
.TEXT
ORG 0x80000000
_start:
  LD R1 loading
  CALL [print]

  OPEN R1 [ramprogram]
  LOADBIN R1 R2
  CLOSE R1
  CALL [R2]

  OPEN R1 [rt2]
  LOADBIN R1 R2
  CLOSE R1
  CALL [R2]

  HLT
