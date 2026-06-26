package buildenv

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type AndroidConfig struct {
	NDKHome string
	ABI     string
	API     int
	HostTag string
	GoOS    string
	GoArch  string
	CC      string
}

func ResolveAndroid(env map[string]string, abi string, api int, hostTag string) (AndroidConfig, error) {
	ndk := env["ANDROID_NDK_HOME"]
	if ndk == "" {
		ndk = env["ANDROID_NDK_ROOT"]
	}
	if ndk == "" {
		ndk = findLatestNDK(env["ANDROID_HOME"])
	}
	if ndk == "" {
		ndk = findLatestNDK(env["ANDROID_SDK_ROOT"])
	}
	if ndk == "" {
		return AndroidConfig{}, fmt.Errorf("ANDROID_NDK_HOME or ANDROID_NDK_ROOT is required")
	}
	if api <= 0 {
		return AndroidConfig{}, fmt.Errorf("android api must be positive")
	}

	var goarch string
	var clang string
	switch abi {
	case "arm64-v8a":
		goarch = "arm64"
		clang = fmt.Sprintf("aarch64-linux-android%d-clang", api)
	default:
		return AndroidConfig{}, fmt.Errorf("unsupported abi %q", abi)
	}

	resolvedHostTag := hostTag
	cc := filepath.Join(ndk, "toolchains", "llvm", "prebuilt", resolvedHostTag, "bin", clang)
	if _, err := os.Stat(cc); err != nil {
		if fallbackHostTag, fallbackCC, ok := findInstalledClang(ndk, clang); ok {
			resolvedHostTag = fallbackHostTag
			cc = fallbackCC
		}
	}
	return AndroidConfig{
		NDKHome: ndk,
		ABI:     abi,
		API:     api,
		HostTag: resolvedHostTag,
		GoOS:    "android",
		GoArch:  goarch,
		CC:      cc,
	}, nil
}

func findLatestNDK(sdk string) string {
	if sdk == "" {
		return ""
	}
	entries, err := os.ReadDir(filepath.Join(sdk, "ndk"))
	if err != nil {
		return ""
	}
	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	if len(versions) == 0 {
		return ""
	}
	sort.Strings(versions)
	return filepath.Join(sdk, "ndk", versions[len(versions)-1])
}

func findInstalledClang(ndk string, clang string) (string, string, bool) {
	prebuiltRoot := filepath.Join(ndk, "toolchains", "llvm", "prebuilt")
	entries, err := os.ReadDir(prebuiltRoot)
	if err != nil {
		return "", "", false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cc := filepath.Join(prebuiltRoot, entry.Name(), "bin", clang)
		if _, err := os.Stat(cc); err == nil {
			return entry.Name(), cc, true
		}
	}
	return "", "", false
}
