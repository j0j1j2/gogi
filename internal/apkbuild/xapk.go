package apkbuild

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type XAPKOptions struct {
	XAPKPath     string
	OutPath      string
	ABI          string
	LibraryPath  string
	WorkDir      string
	DebugKeyPath string
	DebugKeyPass string
	Runner       Runner
	Stdout       io.Writer
	Stderr       io.Writer
}

type zipAPKEntry struct {
	Name string
	Size uint64
}

func BuildXAPK(opts XAPKOptions) error {
	workDir := opts.WorkDir
	if workDir == "" {
		dir, err := os.MkdirTemp("", "gogi-xapk-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		workDir = dir
	}
	reader, err := zip.OpenReader(opts.XAPKPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	var apkEntries []zipAPKEntry
	for _, file := range reader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".apk") {
			apkEntries = append(apkEntries, zipAPKEntry{Name: file.Name, Size: file.UncompressedSize64})
		}
	}
	base, ok := selectBaseAPK(apkEntries)
	if !ok {
		return fmt.Errorf("xapk has no apk entries")
	}
	baseIn := filepath.Join(workDir, "base-in.apk")
	baseOut := filepath.Join(workDir, "base-out.apk")
	if err := extractZipFile(&reader.Reader, base.Name, baseIn); err != nil {
		return err
	}
	if err := BuildAPK(APKOptions{
		APKPath:      baseIn,
		OutPath:      baseOut,
		ABI:          opts.ABI,
		LibraryPath:  opts.LibraryPath,
		WorkDir:      filepath.Join(workDir, "apk"),
		DebugKeyPath: opts.DebugKeyPath,
		DebugKeyPass: opts.DebugKeyPass,
		Runner:       opts.Runner,
		Stdout:       opts.Stdout,
		Stderr:       opts.Stderr,
	}); err != nil {
		return err
	}
	return rewriteXAPK(&reader.Reader, base.Name, baseOut, opts.OutPath)
}

func selectBaseAPK(entries []zipAPKEntry) (zipAPKEntry, bool) {
	if len(entries) == 0 {
		return zipAPKEntry{}, false
	}
	for _, entry := range entries {
		if filepath.Base(entry.Name) == "base.apk" {
			return entry, true
		}
	}
	selected := entries[0]
	for _, entry := range entries[1:] {
		if entry.Size > selected.Size {
			selected = entry
		}
	}
	return selected, true
}

func extractZipFile(reader *zip.Reader, name string, outPath string) error {
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}
	return fmt.Errorf("zip entry %q not found", name)
}

func rewriteXAPK(reader *zip.Reader, replaceName string, replacementPath string, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := zip.NewWriter(outFile)
	defer writer.Close()
	for _, file := range reader.File {
		header := file.FileHeader
		header.Method = zip.Deflate
		out, err := writer.CreateHeader(&header)
		if err != nil {
			return err
		}
		if file.Name == replaceName {
			replacement, err := os.Open(replacementPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, replacement); err != nil {
				_ = replacement.Close()
				return err
			}
			_ = replacement.Close()
			continue
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = in.Close()
			return err
		}
		_ = in.Close()
	}
	return nil
}
