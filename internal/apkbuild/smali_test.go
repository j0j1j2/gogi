package apkbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAndroidClassName(t *testing.T) {
	tests := []struct {
		pkg  string
		name string
		want string
	}{
		{pkg: "com.example.app", name: ".MainActivity", want: "com.example.app.MainActivity"},
		{pkg: "com.example.app", name: "MainActivity", want: "com.example.app.MainActivity"},
		{pkg: "com.example.app", name: "com.other.MainActivity", want: "com.other.MainActivity"},
	}
	for _, tt := range tests {
		if got := ResolveAndroidClassName(tt.pkg, tt.name); got != tt.want {
			t.Fatalf("ResolveAndroidClassName(%q, %q) = %q, want %q", tt.pkg, tt.name, got, tt.want)
		}
	}
}

func TestFindLaunchActivity(t *testing.T) {
	manifest := []byte(`<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.app">
  <application>
    <activity android:name=".MainActivity" android:exported="true">
      <intent-filter>
        <action android:name="android.intent.action.MAIN" />
        <category android:name="android.intent.category.LAUNCHER" />
      </intent-filter>
    </activity>
  </application>
</manifest>`)
	info, err := ParseManifest(manifest)
	if err != nil {
		t.Fatalf("ParseManifest returned error: %v", err)
	}
	if info.Package != "com.example.app" {
		t.Fatalf("package = %q", info.Package)
	}
	if info.LaunchActivity != "com.example.app.MainActivity" {
		t.Fatalf("launch activity = %q", info.LaunchActivity)
	}
}

func TestInjectLoadLibraryIntoOnCreate(t *testing.T) {
	input := `.class public Lcom/example/app/MainActivity;
.super Landroid/app/Activity;

.method protected onCreate(Landroid/os/Bundle;)V
    .locals 1

    invoke-super {p0, p1}, Landroid/app/Activity;->onCreate(Landroid/os/Bundle;)V
    return-void
.end method
`
	output, changed, err := InjectLoadLibrary(input)
	if err != nil {
		t.Fatalf("InjectLoadLibrary returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed")
	}
	if !strings.Contains(output, `const-string v0, "gogi"`) {
		t.Fatalf("output missing const-string: %s", output)
	}
	if !strings.Contains(output, `invoke-static {v0}, Ljava/lang/System;->loadLibrary(Ljava/lang/String;)V`) {
		t.Fatalf("output missing loadLibrary: %s", output)
	}
	if strings.Count(output, `loadLibrary`) != 1 {
		t.Fatalf("expected one loadLibrary call: %s", output)
	}
}

func TestInjectLoadLibraryIsIdempotent(t *testing.T) {
	input := `.class public Lcom/example/app/MainActivity;
.super Landroid/app/Activity;

.method protected onCreate(Landroid/os/Bundle;)V
    .locals 1
    const-string v0, "gogi"
    invoke-static {v0}, Ljava/lang/System;->loadLibrary(Ljava/lang/String;)V
    return-void
.end method
`
	_, changed, err := InjectLoadLibrary(input)
	if err != nil {
		t.Fatalf("InjectLoadLibrary returned error: %v", err)
	}
	if changed {
		t.Fatal("expected no change")
	}
}

func TestFindSmaliFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "smali_classes2", "com", "example", "app", "MainActivity.smali")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(".class public Lcom/example/app/MainActivity;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := FindSmaliFile(root, "com.example.app.MainActivity")
	if err != nil {
		t.Fatalf("FindSmaliFile returned error: %v", err)
	}
	if got != path {
		t.Fatalf("path = %q, want %q", got, path)
	}
}
