package helpers_test

import (
	"testing"
	"unsafe"

	"github.com/thomaspoignant/go-feature-flag/cmd/wasm/helpers"
)

func TestWasmReadBufferFromMemory_ZeroLengthReturnsEmptySlice(t *testing.T) {
	mem := []int32{0x41, 0x42}
	result := helpers.WasmReadBufferFromMemory((*uint32)(unsafe.Pointer(&mem[0])), 0)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}
}

func TestWasmCopyBufferToMemory_ReturnsCorrectPointerAndSize(t *testing.T) {
	buffer := []byte{0x10, 0x20, 0x30}
	result := helpers.WasmCopyBufferToMemory(buffer)
	pos := uint32(result >> 32)
	size := uint32(result & 0xFFFFFFFF)
	if size != uint32(len(buffer)) {
		t.Errorf("Expected size %d, got %d", len(buffer), size)
	}
	if pos == 0 {
		t.Errorf("Expected non-zero pointer position")
	}
}

func TestWasmCopyBufferToMemory_EmptyBufferPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expected panic for empty buffer, but did not panic")
		}
	}()
	helpers.WasmCopyBufferToMemory([]byte{})
}
