package buildenv

import (
	"fmt"
	"path/filepath"
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

	cc := filepath.Join(ndk, "toolchains", "llvm", "prebuilt", hostTag, "bin", clang)
	return AndroidConfig{
		NDKHome: ndk,
		ABI:     abi,
		API:     api,
		HostTag: hostTag,
		GoOS:    "android",
		GoArch:  goarch,
		CC:      cc,
	}, nil
}
