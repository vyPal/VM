# vyPal/VM
A very basic 32bit virtual machine written in Go.

## Building
```bash
go build .
```

## Usage
To run a program, simply pass the path to the binary or assembly file as an argument.
```bash
./VM test.bin
```

### VFS (Virtual File System)
The VM has support for a virtual file system. The VFS can be used to read and write files from the host system.

There is currently only one driver available, the `folder` driver. This driver allows the VM to read and write files from a folder on the host system.

The VFS is enabled by default and will create a folder named `vmdata` in the current working directory. This folder will be used as the root directory for the VFS.
You can change the root directory by using the `-root` flag.
```bash
./VM -root /path/to/folder test.bin
```

### Generating Bytecode
Internally, the VM uses a custom bytecode format. When an assembly file is passed as an argument, the VM will automatically assemble it into bytecode.

To assemble a file and save the bytecode to a file, use the `-bytecode` flag. If you want to change the name of the output file, use the `-output` flag.
```bash
./VM -bytecode -output test.bin test.asm
```

### Printing calltable for assembly file(s)
To print the calltable for an assembly file, use the `-calltable` flag.
```bash
./VM -calltable test.asm
```

## Assembly
The VM uses a custom assembly language. The following instructions are supported:
<details>

<summary> List of instructions </summary>

(`r` - Register, `im` - Indirect Memory, `dm` - Direct Memory, `i` - Immediate)

- `NOP` - No operation
- `LD <r> <r/im/dm/i>` - Load a value into a register
- `ST <dm/im> <r>` - Store a value from a register into memory
- `ADD <r> <r/im/dm/i>` - Add two values and store the result in a register
- `SUB <r> <r/im/dm/i>` - Subtract two values and store the result in a register
- `MUL <r> <r/im/dm/i>` - Multiply two values and store the result in a register
- `DIV <r> <r/im/dm/i>` - Divide two values and store the result in a register
- `MOD <r> <r/im/dm/i>` - Modulo two values and store the result in a register
- `AND <r> <r/im/dm/i>` - Bitwise AND two values and store the result in a register
- `OR <r> <r/im/dm/i>` - Bitwise OR two values and store the result in a register
- `XOR <r> <r/im/dm/i>` - Bitwise XOR two values and store the result in a register
- `NOT <r>` - Bitwise NOT a value and store the result in a register
- `SHL <r> <r/im/dm/i>` - Shift a value left and store the result in a register
- `SHR <r> <r/im/dm/i>` - Shift a value right and store the result in a register
- `CMP <r> <r/im/dm/i>` - Compare two values
- `JMP <dm/im/i>` - Jump to an address
- `JEQ <dm/im/i>` - Jump to an address if the previous comparison was equal
- `JNE <dm/im/i>` - Jump to an address if the previous comparison was not equal
- `JGT <dm/im/i>` - Jump to an address if the previous comparison was greater
- `JLT <dm/im/i>` - Jump to an address if the previous comparison was less
- `JGE <dm/im/i>` - Jump to an address if the previous comparison was greater or equal
- `JLE <dm/im/i>` - Jump to an address if the previous comparison was less or equal
- `CALL <dm/im/i>` - Call a function
- `RET` - Return from a function
- `PUSH <r/im/dm/i>` - Push a value onto the stack
- `POP <r/im/dm>` - Pop a value from the stack
- `HLT` - Halt the program
- `INC <r>` - Increment a value
- `DEC <r>` - Decrement a value
- `OPEN <r> <dm/im>` - Open a file and store the file descriptor in a register
- `READ <r> <r/im/dm> <r/im/dm/i>` - Read from a file
- `WRITE <r> <r/im/dm> <r/im/dm/i>` - Write to a file
- `SEEK <r> <r/im/dm> <i>` - Seek to a position in a file
- `LOADBIN <r> <r>` - Load a binary file into memory
- `CLOSE <r>` - Close a file
- `MALLOC <r/im/dm/i> <r>` - Allocate memory on heap, takes size and register to store the address
- `FREE <r/dm/im> <r/dm/im/i>` - Free memory on heap, takes start address and size
- `INT <i>` - Call an interrupt

</details>

### Registers
The VM has 19 general-purpose registers. Each register is 32 bits, but can be accessed as 8, 16, or 32 bits.
- `R0` - `R15`  - General-purpose Registers
- `R16 (PC)` - Program Counter
- `R17 (SP)` - Stack Pointer
- `R18 (HP)` - Heap Pointer

### Memory
The VM supports up to 4GB of total memory. The memory is split into 3 sections:
- `RAM` - 2GB of general-purpose R/W memory (0x00000000 - 0x7FFFFFFF)
- `ROM` - 128MB of read-only memory (0x80000000 - 0x87FFFFFF)
- `IVT` - 1KB for interrupt vector table (0x88000000 - 0x880003FF)
- `VRAM` - 1KB of video memory (0xFFFFF000 - 0xFFFFFFFF)

The rest of the memory is currently unused and reserved for future use.

### Operands
Operands can be registers, immediate values, direct memory addresses, or indirect memory addresses.

#### Registers
Registers can be accessed as 8, 16, or 32 bits.
You should use `R` with the appropriate number and size prefix, or you can use the special name (e.g. `PC`, `SP`, `HP`).
```asm
R0 ; 32-bit register
R0W ; 16-bit register
R0B ; 8-bit register
```
#### Immediate Values
Immediate values are constants that are directly encoded into the instruction.
```asm
0x1234 ; 32-bit immediate value
0x12 ; 8-bit immediate value
labelname ; Label address (calculated at assembly time)
```

#### Direct Memory Addresses
Direct memory addresses are used to access memory directly.
```asm
[0x12345678] ; Use immediate value as memory address
[R0] ; Read from memory address stored in register
[R0+0x10] ; Read from memory address stored in register + immediate value (offset)
```

#### Indirect Memory Addresses
Indirect memory addresses are used to access memory indirectly.
```asm
[[0x12345678]] ; Read from memory address that is stored at the immediate value
[[R0]] ; Read from memory address that is stored at the memory address stored in register
[[R0+0x10]] ; Read from memory address that is stored at the memory address stored in register + immediate value (offset)
```

### Labels
Labels can be defined in the `.TEXT` section to make the code more readable. Labels can be used as jump targets or as addresses for memory operations.

Label addresses are calculated at assembly time and are not stored in the bytecode.
```asm
.TEXT
<label>: ; Define a label
    JMP [<label>] ; Jump to a label
    LD R0 [<label>] ; Load the address of a label into a register
```

There is also a special label called `_start` that can be used to indicate the starting point of the program. There can only be one `_start` label in the program.
```asm
.TEXT
_start:
    ; Program code
```

### Stack
The Stack is not directly accessible, but can be used with the `PUSH` and `POP` instructions. The stack grows downwards from the top of the RAM.
```asm
PUSH R0 ; Push the value of a register onto the Stack
POP R0 ; Pop a value from the Stack into a register
```

### Sections
The assembly file supports 2 types of sections:
- `.DATA` - Data section for storing constants
- `.TEXT` - Text section for storing instructions

These two sections can both be used multiple times in a single file.

### Defining Constants
Constants can be defined in the `DATA` section. The syntax is as follows:
```asm
.DATA
    <name> <size> <value> ; Define a constant with a name, size (in bytes), and value
    <name> <size> <value>, <value>, ... ; Define an array of constants, the size should be the size of individual elements

.TEXT
    LD R0 [<name>] ; Load the value of a constant into a register
    LD R0 <name> ; Load the address in memory of the constant into a register
```

The possible values for size are:
- `DB` - Byte
- `DW` - Word (16 bits)
- `DD` - Double-word (32 bits)

### Sectors
The assembly language allows the user to split the code into "sectors" by defining starting postitions for the `DATA` and `TEXT` sections. This is useful for creating libraries or splitting the code into multiple files.

This can also be used to load the program at a specific address in memory, including the RAM, and executing from there, however this is not recommended and in some case not allowed.

Specifically, this should only be used when creating a bootloader or a kernel, where the program needs to be loaded at a specific address in memory.
```asm
ORG 0x00000000 ; Any data or instructions after this will be loaded at this address
```

### Interrupts
The VM supports software interrupts. To define a interrupt handler, define it as a label in the `.TEXT` section, and store it in the IVT.
The IVT acts as another section of memory, and the address of the interrupt handler should be stored at the appropriate index in the IVT.
Since each address needs to be stored as a 32-bit value, the IVT can store up to 256 interrupt handlers, each 4 bytes long.
```asm
.TEXT
int0:
    ; Interrupt handler code
    RET ; Return from interrupt

_start:
    LD R0 int0 ; Load the address of the interrupt handler into a register
    ST [0x88000000] R0 ; Store the address in the IVT

    INT 0 ; Call the interrupt
```

## Encoding instructions
Each instruction is encoded as an array of bytes. The first byte is the opcode, followed by the operands.

### Encoding Operands
Each operand can either be a specific type, or can support multiple types. The operand types are as follows:
- `0x00` - Reg
- `0x01` - DMem
- `0x02` - IMem
- `0x03` - Imm

Both DMem and IMem operands support some additional types:
- `0x00` - Address
- `0x01` - Register
- `0x02` - Offset (Register + Immediate)

If an operand supports multiple types, the operand type is encoded as a separate byte before the operand value.
If the operand only supports one type, the operand value is directly encoded, and the type byte is omitted.

#### Reg (Register)
`(RegNum | RegSize << 6)`
- `RegNum` - Register number (0-63)
- `RegSize` - Register size (0-2)

#### DMem (Direct Memory)
Before the value itself, the type of the operand is encoded as a separate byte.

Depending on the type, the operand value is encoded differently:
- `Address` - 32-bit memory address (encoded as 4 bytes in little-endian)
- `Register` - Register number (0-63)
- `Offset` - Register number (0-15) and 32-bit immediate value (register number encoded first, then 4 bytes in little-endian)

#### IMem (Indirect Memory)
Same as DMem, the type of the operand is encoded as a separate byte before the value.

#### Imm (Immediate)
The immediate value is encoded as 4 bytes in little-endian.

## Bytecode Format
The bytecode format is a simple binary format that encodes the individual sectors of the program.

### Header
The bytecode file starts with a header that contains the following information:
- `Magic` - 4 bytes (0x736F6265)
- `Version` - 4 bytes (Version of the bytecode format)
- `SectorCount` - 4 bytes (Number of sectors in the file)
- `StartAddress` - 4 bytes (Address to use as the initial instruction pointer)

### Sectors
Each sector is encoded as follows:
- `StartAddress` - 4 bytes (Address to load the sector at)
- `Size` - 4 bytes (Size of the sector in bytes)
- `Data` - `Size` bytes (Sector data)


