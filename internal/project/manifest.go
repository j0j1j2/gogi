package project

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Name     string         `toml:"name"`
	Build    BuildConfig    `toml:"build"`
	Overlay  OverlayConfig  `toml:"overlay"`
	Frontend FrontendConfig `toml:"frontend"`
	Backend  BackendConfig  `toml:"backend"`
}

type BuildConfig struct {
	Package string   `toml:"package"`
	ABIs    []string `toml:"abis"`
	MinSDK  int      `toml:"min_sdk"`
}

type OverlayConfig struct {
	Enabled       bool   `toml:"enabled"`
	Mode          string `toml:"mode"`
	Width         int    `toml:"width"`
	Height        int    `toml:"height"`
	CollapsedSize int    `toml:"collapsed_size"`
	Draggable     bool   `toml:"draggable"`
}

type FrontendConfig struct {
	Entry string `toml:"entry"`
}

type BackendConfig struct {
	Entry string `toml:"entry"`
}

func LoadManifest(path string) (*Manifest, error) {
	var manifest Manifest
	if _, err := toml.DecodeFile(path, &manifest); err != nil {
		return nil, fmt.Errorf("load manifest %s: %w", path, err)
	}
	return &manifest, nil
}

func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(m.Build.Package) == "" {
		return fmt.Errorf("build package is required")
	}
	if len(m.Build.ABIs) == 0 {
		return fmt.Errorf("at least one ABI is required")
	}
	if m.Build.MinSDK == 0 {
		return fmt.Errorf("build min_sdk is required")
	}
	if m.Overlay.Enabled && strings.TrimSpace(m.Overlay.Mode) == "" {
		return fmt.Errorf("overlay mode is required")
	}
	if strings.TrimSpace(m.Frontend.Entry) == "" {
		return fmt.Errorf("frontend entry is required")
	}
	if strings.TrimSpace(m.Backend.Entry) == "" {
		return fmt.Errorf("backend entry is required")
	}

	return nil
}
