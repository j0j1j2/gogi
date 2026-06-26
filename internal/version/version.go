package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type Info struct {
	Version  string
	Revision string
	Time     string
	Modified string
	Go       string
}

func Current() Info {
	info := Info{
		Version: "dev",
		Go:      runtime.Version(),
	}
	if build, ok := debug.ReadBuildInfo(); ok {
		if build.Main.Version != "" && build.Main.Version != "(devel)" {
			info.Version = build.Main.Version
		}
		for _, setting := range build.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.Revision = setting.Value
			case "vcs.time":
				info.Time = setting.Value
			case "vcs.modified":
				info.Modified = setting.Value
			}
		}
	}
	return info
}

func (i Info) String() string {
	version := fallback(i.Version, "dev")
	revision := fallback(i.Revision, pseudoCommit(version))
	if len(revision) > 12 {
		revision = revision[:12]
	}
	buildTime := fallback(i.Time, pseudoTime(version))
	modified := fallback(i.Modified, "unknown")
	goVersion := fallback(i.Go, runtime.Version())
	return fmt.Sprintf("gogi %s\ncommit %s\ndate %s\nmodified %s\ngo %s\n", version, revision, buildTime, modified, goVersion)
}

func fallback(value string, defaultValue string) string {
	if value == "" {
		if defaultValue == "" {
			return "unknown"
		}
		return defaultValue
	}
	return value
}

func pseudoCommit(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-1]
}

func pseudoTime(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) < 3 {
		return ""
	}
	parsed, err := time.Parse("20060102150405", parts[len(parts)-2])
	if err != nil {
		return ""
	}
	return parsed.UTC().Format(time.RFC3339)
}
