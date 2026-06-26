package project

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func ParseHexBytes(input string) ([]byte, error) {
	clean := strings.ReplaceAll(strings.TrimSpace(input), " ", "")
	if clean == "" {
		return nil, nil
	}
	if len(clean)%2 != 0 {
		return nil, fmt.Errorf("hex byte string has odd length")
	}
	out, err := hex.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("decode hex bytes: %w", err)
	}
	return out, nil
}
