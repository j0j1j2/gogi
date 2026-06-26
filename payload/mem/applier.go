package mem

import (
	"bytes"
	"fmt"
)

type AppliedPatch struct {
	Spec     PatchSpec
	Address  uintptr
	Offset   int
	Original []byte
	Length   int
	Perms    string
}

func ApplyToSlice(buf []byte, offset int, spec PatchSpec) (AppliedPatch, error) {
	if offset < 0 || offset+len(spec.Replace) > len(buf) {
		return AppliedPatch{}, fmt.Errorf("patch %q range is outside target buffer", spec.ID)
	}
	current := buf[offset : offset+len(spec.Replace)]
	if len(spec.Expect) > 0 && !bytes.Equal(current[:len(spec.Expect)], spec.Expect) {
		return AppliedPatch{}, fmt.Errorf("patch %q expected bytes mismatch", spec.ID)
	}
	original := append([]byte(nil), current...)
	copy(current, spec.Replace)
	return AppliedPatch{
		Spec:     spec,
		Offset:   offset,
		Original: original,
		Length:   len(spec.Replace),
	}, nil
}

func RestoreSlice(buf []byte, applied AppliedPatch) error {
	if len(applied.Original) == 0 {
		return fmt.Errorf("patch %q has no original bytes", applied.Spec.ID)
	}
	if applied.Offset < 0 || applied.Offset+len(applied.Original) > len(buf) {
		return fmt.Errorf("patch %q restore buffer too small", applied.Spec.ID)
	}
	copy(buf[applied.Offset:applied.Offset+len(applied.Original)], applied.Original)
	return nil
}
