package version

import (
	"strings"
	"testing"
)

func TestInfoStringIncludesVersionAndCommit(t *testing.T) {
	info := Info{
		Version:  "v0.0.0-20260626143354-2315612dde57",
		Revision: "2315612dde57aabbcc",
		Time:     "2026-06-26T14:33:54Z",
		Modified: "false",
		Go:       "go1.25",
	}

	out := info.String()

	for _, want := range []string{
		"gogi v0.0.0-20260626143354-2315612dde57",
		"commit 2315612dde57",
		"date 2026-06-26T14:33:54Z",
		"modified false",
		"go go1.25",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("version output missing %q:\n%s", want, out)
		}
	}
}

func TestInfoStringUsesDevFallbacks(t *testing.T) {
	out := (Info{}).String()

	if !strings.Contains(out, "gogi dev") {
		t.Fatalf("version output should use dev fallback:\n%s", out)
	}
	if !strings.Contains(out, "commit unknown") {
		t.Fatalf("version output should use unknown commit fallback:\n%s", out)
	}
}
