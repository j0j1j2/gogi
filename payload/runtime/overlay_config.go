package runtime

import (
	"encoding/json"
	"strconv"
)

var overlayWidth = "320"
var overlayHeight = "420"
var overlayCollapsedSize = "56"
var overlayDraggable = "true"

type overlayConfig struct {
	Width         int  `json:"width"`
	Height        int  `json:"height"`
	CollapsedSize int  `json:"collapsed_size"`
	Draggable     bool `json:"draggable"`
}

func overlayConfigJSON() string {
	config := overlayConfig{
		Width:         parseIntDefault(overlayWidth, 320),
		Height:        parseIntDefault(overlayHeight, 420),
		CollapsedSize: parseIntDefault(overlayCollapsedSize, 56),
		Draggable:     parseBoolDefault(overlayDraggable, true),
	}
	data, err := json.Marshal(config)
	if err != nil {
		return `{"width":320,"height":420,"collapsed_size":56,"draggable":true}`
	}
	return string(data)
}

func parseIntDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseBoolDefault(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
