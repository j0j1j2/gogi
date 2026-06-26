package project

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Name        string     `toml:"name"`
	Package     string     `toml:"package"`
	ABIs        []string   `toml:"abis"`
	API         int        `toml:"api"`
	Entry       []string   `toml:"entry"`
	MenuBackend string     `toml:"menu_backend"`
	Patches     []Patch    `toml:"patches"`
	Menu        MenuConfig `toml:"menu"`
}

type MenuConfig struct {
	Toggles []MenuToggle `toml:"toggles"`
}

type Patch struct {
	ID      string `toml:"id"`
	Library string `toml:"library"`
	RVA     string `toml:"rva"`
	Expect  string `toml:"expect"`
	Replace string `toml:"replace"`
	Startup bool   `toml:"startup"`
}

type MenuToggle struct {
	ID      string `toml:"id"`
	Label   string `toml:"label"`
	Patch   string `toml:"patch"`
	Initial bool   `toml:"initial"`
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
	if strings.TrimSpace(m.Package) == "" {
		return fmt.Errorf("package is required")
	}
	if len(m.ABIs) == 0 {
		return fmt.Errorf("at least one ABI is required")
	}
	if m.API == 0 {
		return fmt.Errorf("api is required")
	}

	patchIDs := map[string]bool{}
	for _, patch := range m.Patches {
		if patch.ID == "" {
			return fmt.Errorf("patch id is required")
		}
		if patchIDs[patch.ID] {
			return fmt.Errorf("duplicate patch id %q", patch.ID)
		}
		patchIDs[patch.ID] = true
		if patch.Library == "" {
			return fmt.Errorf("patch %q library is required", patch.ID)
		}
		if patch.RVA == "" {
			return fmt.Errorf("patch %q rva is required", patch.ID)
		}
		expect, err := ParseHexBytes(patch.Expect)
		if err != nil {
			return fmt.Errorf("patch %q expect: %w", patch.ID, err)
		}
		replace, err := ParseHexBytes(patch.Replace)
		if err != nil {
			return fmt.Errorf("patch %q replace: %w", patch.ID, err)
		}
		if len(expect) != 0 && len(replace) != 0 && len(expect) != len(replace) {
			return fmt.Errorf("patch %q expect and replace lengths differ", patch.ID)
		}
	}

	for _, toggle := range m.Menu.Toggles {
		if toggle.ID == "" {
			return fmt.Errorf("menu toggle id is required")
		}
		if toggle.Label == "" {
			return fmt.Errorf("menu toggle %q label is required", toggle.ID)
		}
		if !patchIDs[toggle.Patch] {
			return fmt.Errorf("menu toggle %q references unknown patch %q", toggle.ID, toggle.Patch)
		}
	}

	return nil
}
