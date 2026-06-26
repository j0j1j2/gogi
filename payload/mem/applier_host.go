//go:build !android

package mem

import "fmt"

func ApplyProcessPatch(module Module, spec PatchSpec) (AppliedPatch, error) {
	return AppliedPatch{}, fmt.Errorf("process patching requires android build")
}

func RestoreProcessPatch(applied AppliedPatch) error {
	return fmt.Errorf("process patch restore requires android build")
}
