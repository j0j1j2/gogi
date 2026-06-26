package runtime

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStartServesMenu(t *testing.T) {
	Start(nil)

	var body string
	var lastErr error
	for range 20 {
		resp, err := http.Get(MenuURL())
		if err == nil {
			data, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				t.Fatalf("read menu response: %v", readErr)
			}
			body = string(data)
			break
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}

	if body == "" {
		t.Fatalf("menu server did not respond: %v", lastErr)
	}
	if !strings.Contains(body, "gogi") {
		t.Fatalf("menu response missing gogi marker: %q", body)
	}
}
