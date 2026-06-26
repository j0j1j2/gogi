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
	if err := withWritablePage(addr, len(spec.Replace), module.Perms, func() {
		copy(target, spec.Replace)
	}); err != nil {
		return AppliedPatch{}, err
	}
	C.gogi_clear_cache((*C.char)(unsafe.Pointer(addr)), (*C.char)(unsafe.Pointer(addr+uintptr(len(spec.Replace)))))
	return AppliedPatch{Spec: spec, Address: addr, Original: original, Length: len(spec.Replace), Perms: module.Perms}, nil
}

func RestoreProcessPatch(applied AppliedPatch) error {
	if applied.Address == 0 {
		return fmt.Errorf("patch %q has no process address", applied.Spec.ID)
	}
	target := unsafe.Slice((*byte)(unsafe.Pointer(applied.Address)), len(applied.Original))
	return withWritablePage(applied.Address, len(applied.Original), applied.Perms, func() {
		copy(target, applied.Original)
	})
}

func withWritablePage(addr uintptr, length int, perms string, fn func()) error {
	pageSize := uintptr(os.Getpagesize())
	pageStart := addr & ^(pageSize - 1)
	pageEnd := (addr + uintptr(length) + pageSize - 1) & ^(pageSize - 1)
	pageLen := pageEnd - pageStart
	writeProt, err := protFromPerms(perms, true)
	if err != nil {
		return err
	}
	restoreProt, err := protFromPerms(perms, false)
	if err != nil {
		return err
	}
	if _, _, errno := syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, writeProt); errno != 0 {
		return errno
	}
	fn()
	_, _, _ = syscall.RawSyscall(syscall.SYS_MPROTECT, pageStart, pageLen, restoreProt)
	return nil
}
