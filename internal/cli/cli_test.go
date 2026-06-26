package cli

import (
	"bytes"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"help"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("gogi init <name>")) {
		t.Fatalf("help output missing init usage: %q", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"missing"}, &out, &errOut)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("unknown command")) {
		t.Fatalf("stderr missing unknown command message: %q", errOut.String())
	}
}
