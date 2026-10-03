package helpers

import "unsafe"

// nolint:gosec, govet
// WasmReadBufferFromMemory reads a buffer from memory and returns it as a byte slice.
func WasmReadBufferFromMemory(bufferPosition *uint32, length uint32) []byte {
	subjectBuffer := make([]byte, length)
	if length == 0 {
		return subjectBuffer
	}
	src := unsafe.Slice((*byte)(unsafe.Pointer(bufferPosition)), length)
	copy(subjectBuffer, src)
	return subjectBuffer
}

// lastOutput keeps the most recent output buffer reachable: the raw pointer
// returned to the host is not a reference the garbage collector knows about,
// so without it the buffer could be reclaimed before the host reads it.
// The module is single-threaded per instance, so a single package variable is
// enough: the buffer stays valid until the next call into the module.
//
//nolint:unused // intentionally write-only, it exists to pin the buffer for the GC
var lastOutput []byte

// nolint:gosec
// WasmCopyBufferToMemory copies a buffer to memory and returns a pointer to
// the memory location (high 32 bits) packed with its size (low 32 bits).
// It returns 0 for an empty buffer. The buffer stays valid only until the
// next call into the module.
func WasmCopyBufferToMemory(buffer []byte) uint64 {
	if len(buffer) == 0 {
		return 0
	}
	lastOutput = buffer
	bufferPtr := &buffer[0]
	unsafePtr := uintptr(unsafe.Pointer(bufferPtr))
	pos := uint32(unsafePtr)
	size := uint32(len(buffer))
	return (uint64(pos) << uint64(32)) | uint64(size)
}
