package cpu

type Register struct {
  value byte
  ID byte
}

func (r *Register) Read() byte {
  return r.value
}

func (r *Register) Write(val byte) {
  r.value = val
}

func (r *Register) Increment() {
  r.value++
}

func (r *Register) Decrement() {
  r.value--
}

func (r *Register) Add(val byte) {
  r.value += val
}

func (r *Register) Subtract(val byte) {
  r.value -= val
}

func (r *Register) And(val byte) {
  r.value &= val
}

func (r *Register) Or(val byte) {
  r.value |= val
}

func (r *Register) Xor(val byte) {
  r.value ^= val
}

func (r *Register) Not() {
  r.value = ^r.value
}

func (r *Register) ShiftLeft() {
  r.value <<= 1
}

func (r *Register) ShiftRight() {
  r.value >>= 1
}

type LargeRegister struct {
  value uint16
}

func (r *LargeRegister) Read() uint16 {
  return r.value
}

func (r *LargeRegister) Write(val uint16) {
  r.value = val
}

func (r *LargeRegister) Increment() uint16 {
  r.value++
  return r.value
}

func (r *LargeRegister) Decrement() uint16 {
  r.value--
  return r.value
}

type Registers struct {
  A *Register
  B *Register
  C *Register
  D *Register
  E *Register
  H *Register
  L *Register
}

func NewRegisters() *Registers {
  return &Registers{&Register{ID: 0}, &Register{ID: 1}, &Register{ID: 2}, &Register{ID: 3}, &Register{ID: 4}, &Register{ID: 5}, &Register{ID: 6}}
}

func (r *Registers) Reset() {
  r.A.Write(0)
  r.B.Write(0)
  r.C.Write(0)
  r.D.Write(0)
  r.E.Write(0)
  r.H.Write(0)
  r.L.Write(0)
}

func (r *Registers) Get(id byte) *Register {
  switch id {
    case 0:
      return r.A
    case 1:
      return r.B
    case 2:
      return r.C
    case 3:
      return r.D
    case 4:
      return r.E
    case 5:
      return r.H
    case 6:
      return r.L
    default:
      return nil
  }
}
