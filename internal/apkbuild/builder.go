package apkbuild

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Runner func(name string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error

type APKOptions struct {
	APKPath      string
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

func BuildAPK(opts APKOptions) error {
	if opts.Runner == nil {
		return fmt.Errorf("runner is required")
	}
	if opts.ABI == "" {
		opts.ABI = "arm64-v8a"
	}
	if opts.DebugKeyPath == "" {
		home, _ := os.UserHomeDir()
		opts.DebugKeyPath = filepath.Join(home, ".android", "debug.keystore")
	}
	if opts.DebugKeyPass == "" {
		opts.DebugKeyPass = "android"
	}
	workDir := opts.WorkDir
	if workDir == "" {
		dir, err := os.MkdirTemp("", "gogi-apk-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		workDir = dir
	}
	decoded := filepath.Join(workDir, "decoded")
	unsigned := filepath.Join(workDir, "unsigned.apk")
	aligned := filepath.Join(workDir, "aligned.apk")

	if err := opts.Runner("apktool", []string{"d", "-f", "-o", decoded, opts.APKPath}, nil, opts.Stdout, opts.Stderr); err != nil {
		return err
	}
	if err := installLibrary(decoded, opts.ABI, opts.LibraryPath); err != nil {
		return err
	}
	if err := InstallHelperSmali(decoded); err != nil {
		return err
	}
	if err := injectDecoded(decoded); err != nil {
		return err
	}
	if err := opts.Runner("apktool", []string{"b", "-f", "-o", unsigned, decoded}, nil, opts.Stdout, opts.Stderr); err != nil {
		return err
	}
	if err := opts.Runner("zipalign", []string{"-f", "4", unsigned, aligned}, nil, opts.Stdout, opts.Stderr); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0o755); err != nil {
		return err
	}
	return opts.Runner("apksigner", []string{
		"sign",
		"--ks", opts.DebugKeyPath,
		"--ks-pass", "pass:" + opts.DebugKeyPass,
		"--key-pass", "pass:" + opts.DebugKeyPass,
		"--out", opts.OutPath,
		aligned,
	}, nil, opts.Stdout, opts.Stderr)
}

func installLibrary(decoded string, abi string, libPath string) error {
	target := filepath.Join(decoded, "lib", abi, "libgogi.so")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(libPath)
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644)
}

func injectDecoded(decoded string) error {
	data, err := os.ReadFile(filepath.Join(decoded, "AndroidManifest.xml"))
	if err != nil {
		return err
	}
	info, err := ParseManifest(data)
	if err != nil {
		return err
	}
	targetClass := info.Application
	if targetClass == "" {
		targetClass = info.LaunchActivity
	}
	if targetClass == "" {
		return fmt.Errorf("manifest has no application class or launch activity")
	}
	smaliPath, err := FindSmaliFile(decoded, targetClass)
	if err != nil && info.LaunchActivity != "" && targetClass != info.LaunchActivity {
		smaliPath, err = FindSmaliFile(decoded, info.LaunchActivity)
	}
	if err != nil {
		return err
	}
	source, err := os.ReadFile(smaliPath)
	if err != nil {
		return err
	}
	updated, changed, err := InjectLoadLibrary(string(source))
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return os.WriteFile(smaliPath, []byte(updated), 0o644)
}
