package helpers

import "unsafe"

// nolint:gosec, govet
// WasmReadBufferFromMemory reads a buffer from memory and returns it as a byte slice.
// It uses unsafe.Slice to return a view directly into WASM linear memory with no copying.
func WasmReadBufferFromMemory(bufferPosition *uint32, length uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(bufferPosition)), length)
}

// nolint:gosec
// WasmCopyBufferToMemory copies a buffer to memory and returns a pointer to the memory location.
func WasmCopyBufferToMemory(buffer []byte) uint64 {
	bufferPtr := &buffer[0]
	unsafePtr := uintptr(unsafe.Pointer(bufferPtr))
	pos := uint32(unsafePtr)
	size := uint32(len(buffer))
	return (uint64(pos) << uint64(32)) | uint64(size)
}
