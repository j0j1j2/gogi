//go:build android

package mem

/*
#include <sys/mman.h>
static void gogi_clear_cache(char* start, char* end) {
	__builtin___clear_cache(start, end);
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func ApplyProcessPatch(module Module, spec PatchSpec) (AppliedPatch, error) {
	if len(spec.Replace) == 0 {
		return AppliedPatch{}, fmt.Errorf("patch %q replacement is empty", spec.ID)
	}
	addr := ResolveAddress(module, spec.RVA)
	target := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(spec.Replace))
	if len(spec.Expect) > 0 && !bytes.Equal(target[:len(spec.Expect)], spec.Expect) {
		return AppliedPatch{}, fmt.Errorf("patch %q expected bytes mismatch", spec.ID)
	}
	original := append([]byte(nil), target...)
	if err := withWritablePage(addr, len(spec.Replace), func() {
		copy(target, spec.Replace)
	}); err != nil {
		return AppliedPatch{}, err
	}
	C.gogi_clear_cache((*C.char)(unsafe.Pointer(addr)), (*C.char)(unsafe.Pointer(addr+uintptr(len(spec.Replace)))))
	return AppliedPatch{Spec: spec, Address: addr, Original: original, Length: len(spec.Replace)}, nil
}

func RestoreProcessPatch(applied AppliedPatch) error {
	if applied.Address == 0 {
		return fmt.Errorf("patch %q has no process address", applied.Spec.ID)
	}
	target := unsafe.Slice((*byte)(unsafe.Pointer(applied.Address)), len(applied.Original))
	return withWritablePage(applied.Address, len(applied.Original), func() {
		copy(target, applied.Original)
	})
}

func withWritablePage(addr uintptr, length int, fn func()) error {
	pageSize := uintptr(os.Getpagesize())
	pageStart := addr & ^(pageSize - 1)
	pageEnd := (addr + uintptr(length) + pageSize - 1) & ^(pageSize - 1)
	pageLen := pageEnd - pageStart
	if _, _, errno := syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC); errno != 0 {
		return errno
	}
	fn()
	_, _, _ = syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, syscall.PROT_READ|syscall.PROT_EXEC)
	return nil
}
