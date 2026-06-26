package mem

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

type Module struct {
	Name  string
	Base  uintptr
	End   uintptr
	Path  string
	Perms string
}

func ParseMaps(r io.Reader) ([]Module, error) {
	var mods []Module
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}
		bounds := strings.Split(fields[0], "-")
		if len(bounds) != 2 {
			continue
		}
		base, err := strconv.ParseUint(bounds[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse base %q: %w", bounds[0], err)
		}
		end, err := strconv.ParseUint(bounds[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse end %q: %w", bounds[1], err)
		}
		path := fields[5]
		mods = append(mods, Module{
			Name:  filepath.Base(path),
			Base:  uintptr(base),
			End:   uintptr(end),
			Path:  path,
			Perms: fields[1],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan maps: %w", err)
	}
	return mods, nil
}
