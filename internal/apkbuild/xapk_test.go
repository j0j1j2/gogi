package apkbuild

import "testing"

func TestSelectBaseAPKPrefersBaseAPK(t *testing.T) {
	files := []zipAPKEntry{
		{Name: "split_config.arm64_v8a.apk", Size: 300},
		{Name: "base.apk", Size: 100},
	}
	got, ok := selectBaseAPK(files)
	if !ok {
		t.Fatal("expected base apk")
	}
	if got.Name != "base.apk" {
		t.Fatalf("selected %q", got.Name)
	}
}

func TestSelectBaseAPKFallsBackToLargestAPK(t *testing.T) {
	files := []zipAPKEntry{
		{Name: "config.apk", Size: 100},
		{Name: "main.apk", Size: 500},
	}
	got, ok := selectBaseAPK(files)
	if !ok {
		t.Fatal("expected apk")
	}
	if got.Name != "main.apk" {
		t.Fatalf("selected %q", got.Name)
	}
}
