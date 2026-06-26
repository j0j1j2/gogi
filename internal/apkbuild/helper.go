package apkbuild

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/smali/com/gogi/*.smali
var helperSmali embed.FS

func InstallHelperSmali(decoded string) error {
	smaliRoot := filepath.Join(decoded, "smali")
	if _, err := os.Stat(smaliRoot); err != nil {
		smaliRoot = filepath.Join(decoded, "smali_classes2")
		if err := os.MkdirAll(smaliRoot, 0o755); err != nil {
			return err
		}
	}
	return fs.WalkDir(helperSmali, "assets/smali", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "assets/smali/")
		target := filepath.Join(smaliRoot, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := helperSmali.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
