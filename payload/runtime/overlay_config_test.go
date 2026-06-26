package runtime

import (
	"encoding/json"
	"testing"
)

func TestOverlayConfigJSONUsesInjectedValues(t *testing.T) {
	oldWidth := overlayWidth
	oldHeight := overlayHeight
	oldCollapsedSize := overlayCollapsedSize
	oldDraggable := overlayDraggable
	t.Cleanup(func() {
		overlayWidth = oldWidth
		overlayHeight = oldHeight
		overlayCollapsedSize = oldCollapsedSize
		overlayDraggable = oldDraggable
	})

	overlayWidth = "480"
	overlayHeight = "360"
	overlayCollapsedSize = "48"
	overlayDraggable = "false"

	var got map[string]any
	if err := json.Unmarshal([]byte(overlayConfigJSON()), &got); err != nil {
		t.Fatalf("overlayConfigJSON returned invalid JSON: %v", err)
	}
	if got["width"] != float64(480) {
		t.Fatalf("width = %#v", got["width"])
	}
	if got["height"] != float64(360) {
		t.Fatalf("height = %#v", got["height"])
	}
	if got["collapsed_size"] != float64(48) {
		t.Fatalf("collapsed_size = %#v", got["collapsed_size"])
	}
	if got["draggable"] != false {
		t.Fatalf("draggable = %#v", got["draggable"])
	}
}
