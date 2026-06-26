package apkbuild

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ManifestInfo struct {
	Package        string
	Application    string
	LaunchActivity string
}

type manifestXML struct {
	Package     string         `xml:"package,attr"`
	Application applicationXML `xml:"application"`
}

type applicationXML struct {
	Name       string        `xml:"http://schemas.android.com/apk/res/android name,attr"`
	Activities []activityXML `xml:"activity"`
}

type activityXML struct {
	Name          string            `xml:"http://schemas.android.com/apk/res/android name,attr"`
	IntentFilters []intentFilterXML `xml:"intent-filter"`
}

type intentFilterXML struct {
	Actions    []nameXML `xml:"action"`
	Categories []nameXML `xml:"category"`
}

type nameXML struct {
	Name string `xml:"http://schemas.android.com/apk/res/android name,attr"`
}

func ParseManifest(data []byte) (ManifestInfo, error) {
	var parsed manifestXML
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return ManifestInfo{}, err
	}
	info := ManifestInfo{
		Package:     parsed.Package,
		Application: ResolveAndroidClassName(parsed.Package, parsed.Application.Name),
	}
	for _, activity := range parsed.Application.Activities {
		if isLauncher(activity) {
			info.LaunchActivity = ResolveAndroidClassName(parsed.Package, activity.Name)
			break
		}
	}
	return info, nil
}

func isLauncher(activity activityXML) bool {
	for _, filter := range activity.IntentFilters {
		hasMain := false
		hasLauncher := false
		for _, action := range filter.Actions {
			if action.Name == "android.intent.action.MAIN" {
				hasMain = true
			}
		}
		for _, category := range filter.Categories {
			if category.Name == "android.intent.category.LAUNCHER" {
				hasLauncher = true
			}
		}
		if hasMain && hasLauncher {
			return true
		}
	}
	return false
}

func ResolveAndroidClassName(pkg string, name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, ".") {
		return pkg + name
	}
	if strings.Contains(name, ".") {
		return name
	}
	return pkg + "." + name
}

func FindSmaliFile(root string, className string) (string, error) {
	rel := filepath.Join(strings.Split(className, ".")...) + ".smali"
	var found string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base != "smali" && !strings.HasPrefix(base, "smali_classes") {
			return nil
		}
		candidate := filepath.Join(path, rel)
		if _, err := os.Stat(candidate); err == nil {
			found = candidate
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("smali class %q not found", className)
	}
	return found, nil
}

func InjectLoadLibrary(input string) (string, bool, error) {
	if strings.Contains(input, `loadLibrary(Ljava/lang/String;)V`) && strings.Contains(input, `"gogi"`) {
		return input, false, nil
	}
	lines := strings.SplitAfter(input, "\n")
	inOnCreate := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ".method ") && strings.Contains(trimmed, " onCreate(Landroid/os/Bundle;)V") {
			inOnCreate = true
			continue
		}
		if inOnCreate && strings.HasPrefix(trimmed, ".locals ") {
			indent := leadingWhitespace(line)
			injection := indent + `const-string v0, "gogi"` + "\n" +
				indent + `invoke-static {v0}, Ljava/lang/System;->loadLibrary(Ljava/lang/String;)V` + "\n"
			var out bytes.Buffer
			for _, before := range lines[:i+1] {
				out.WriteString(before)
			}
			out.WriteString(injection)
			for _, after := range lines[i+1:] {
				out.WriteString(after)
			}
			return out.String(), true, nil
		}
		if inOnCreate && strings.HasPrefix(trimmed, ".end method") {
			return "", false, fmt.Errorf("onCreate has no .locals directive")
		}
	}
	return "", false, fmt.Errorf("onCreate(Landroid/os/Bundle;)V not found")
}

func leadingWhitespace(line string) string {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}
