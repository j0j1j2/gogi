package runtime

import (
	"testing"
	"unsafe"
)

func TestStartCapturesVM(t *testing.T) {
	vm := unsafe.Pointer(uintptr(0x1234))

	Start(vm)

	if CapturedVM() != vm {
		t.Fatalf("CapturedVM() = %p, want %p", CapturedVM(), vm)
	}
}
