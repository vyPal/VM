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

### Generating Bytecode
Internally, the VM uses a custom bytecode format. When an assembly file is passed as an argument, the VM will automatically assemble it into bytecode.

To assemble a file and save the bytecode to a file, use the `-bytecode` flag. If you want to change the name of the output file, use the `-o` flag.
```bash
./VM -bytecode test.asm
```

## Assembly
The VM uses a custom assembly language. The following instructions are supported:
<detailes>
<summary> List of instructions </summary>
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
</details>

### Registers
The VM has 16 general-purpose registers. Each register is 32 bits, but can be accessed as 8, 16, or 32 bits.
- `R0` - `R15`  - General-purpose Registers
- `IP` - Instruction Pointer (Not directly accessible)
- `SP` - Stack Pointer (Not directly accessible)

### Memory
The VM supports up to 4GB of total memory. The memory is split into 3 sections:
- `RAM` - 2GB of general-purpose R/W memory (0x00000000 - 0x7FFFFFFF)
- `ROM` - 128MB of read-only memory (0x80000000 - 0x87FFFFFF)
- `VRAM` - 1KB of video memory (0xFFFFF000 - 0xFFFFFFFF)

The rest of the memory is currently unused and reserved for future use.

### Stack
The VM currently only has a fixed-size stack of 16384 values. The stack is used for function calls and temporary storage.

### Sections
The assembly file supports 2 types of sections:
- `DATA` - Data section for storing constants
- `TEXT` - Text section for storing instructions

These two sections can both be used multiple times in a single file.

### Defining Constants
Constants can be defined in the `DATA` section. The syntax is as follows:
```asm
DATA
    <name> <size> <value> ; Define a constant with a name, size (in bytes), and value
    <name> <size> {<value>, <value>, ...} ; Define an array of constants, the size should be the size of individual elements
```

### Sectors
The assembly language allows the user to split the code into "sectors" by defining starting postitions for the `DATA` and `TEXT` sections. This is useful for creating libraries or splitting the code into multiple files.

This can also be used to load the program at a specific address in memory, including the RAM, and executing from there.
```asm
ORG 0x00000000 ; Any data or instructions after this will be loaded at this address
```

