package runtime

import (
	"sync"
	"unsafe"
)

var startOnce sync.Once
var capturedVM unsafe.Pointer

func Start(vm unsafe.Pointer) {
	startOnce.Do(func() {
		capturedVM = vm
		Logf("gogi runtime started")
		go func() {
			Logf("gogi worker started")
		}()
	})
}

func CapturedVM() unsafe.Pointer {
	return capturedVM
}
