package runtime

import (
	"io/fs"

	"github.com/j0j1j2/gogi/payload/menu"
)

var frontendAssets *menu.Assets

func SetFrontendAssets(fsys fs.FS, root string) {
	frontendAssets = &menu.Assets{
		FS:    fsys,
		Root:  root,
		Index: "index.html",
		CSS:   "style.css",
		JS:    "main.js",
	}
}

func menuAssets() *menu.Assets {
	return frontendAssets
}
