package apkbuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHelperSmali(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "smali"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := InstallHelperSmali(root); err != nil {
		t.Fatalf("InstallHelperSmali returned error: %v", err)
	}
	for _, rel := range []string{
		"smali/com/gogi/GogiOverlay.smali",
		"smali/com/gogi/GogiOverlay$OverlayConfig.smali",
		"smali/com/gogi/GogiOverlay$OverlayTouchListener.smali",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
}
