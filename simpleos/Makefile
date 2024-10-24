BINARY_NAME=VM
FS_ROOT=vmdata

all: run

.PHONY: build
build:
	@echo "Building..."
	go build -o ../${BINARY_NAME} ..

.PHONY: install
install:
	@echo "Installing..."
	go install ..

.PHONY: run
run: install makefs kernel.bin
	@echo "Running..."
	${BINARY_NAME} -root ${FS_ROOT} kernel.bin

.PHONY: makefs
makefs: install runtime.asm rt2.asm
	@echo "Creating filesystem..."
	${BINARY_NAME} -bytecode -output ${FS_ROOT}/runtime.bin runtime.asm
	${BINARY_NAME} -bytecode -output ${FS_ROOT}/rt2.bin rt2.asm

kernel.bin: install kernel.asm std.asm
	@echo "Assembling kernel..."
	${BINARY_NAME} -bytecode -output kernel.bin kernel.asm std.asm
