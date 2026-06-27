package template

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitProject(root string, name string) error {
	files := map[string]string{
		"go.mod":              goModTemplate(name),
		"gogi.toml":           manifestTemplate(name),
		"frontend/index.html": frontendIndexTemplate(),
		"frontend/style.css":  frontendStyleTemplate(),
		"frontend/main.js":    frontendScriptTemplate(),
		"backend/main.go":     backendMainTemplate(),
	}

	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create directory for %s: %w", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
	}
	return nil
}

func goModTemplate(name string) string {
	return fmt.Sprintf(`module %s

go 1.25
`, name)
}

func manifestTemplate(name string) string {
	return fmt.Sprintf(`name = %q

[build]
package = "com.example.target"
abis = ["arm64-v8a"]
min_sdk = 24

[overlay]
enabled = true
mode = "webview"
width = 320
height = 420
collapsed_size = 56
draggable = true

[frontend]
entry = "frontend/index.html"

[backend]
entry = "backend"
`, name)
}

func frontendIndexTemplate() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="style.css">
  <title>gogi menu</title>
</head>
<body>
  <main id="app">
    <h1>gogi</h1>
    <button id="give-coins" type="button">Give coins</button>
    <p id="status" class="status">Ready</p>
  </main>
  <script src="/gogi.js"></script>
  <script src="main.js"></script>
</body>
</html>
`
}

func frontendStyleTemplate() string {
	return `:root {
  color-scheme: dark;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  margin: 0;
  background: rgba(18, 20, 22, 0.9);
  color: #f4f7f8;
}

#app {
  padding: 14px;
}

h1 {
  margin: 0 0 12px;
  font-size: 18px;
  font-weight: 700;
}

button {
  min-height: 44px;
  border: 0;
  border-radius: 6px;
  padding: 0 12px;
  background: #2f7d68;
  color: #ffffff;
  font: inherit;
}

button:disabled {
  opacity: 0.65;
}

.status {
  min-height: 20px;
  margin: 12px 0 0;
  color: #b8c4bd;
  font-size: 13px;
}
`
}

func frontendScriptTemplate() string {
	return `const status = document.getElementById("status");
const giveCoins = document.getElementById("give-coins");

giveCoins.addEventListener("click", async () => {
  giveCoins.disabled = true;
  status.textContent = "Sending...";
  try {
    await gogi.action("give_coins", {amount: 10});
    status.textContent = "Sent";
  } catch (error) {
    status.textContent = error.message;
  } finally {
    giveCoins.disabled = false;
  }
});
`
}

func backendMainTemplate() string {
	return `package backend

import "github.com/j0j1j2/gogi/sdk"

func Init(ctx *sdk.Context) {
	ctx.Logf("backend initialized")

	ctx.Action("give_coins", func(req sdk.ActionRequest) (any, error) {
		ctx.Logf("give_coins called with %s", string(req.Payload))
		return map[string]any{"ok": true}, nil
	})

	// Register memory patches in Go so editor completion and type checking work.
	// ctx.RegisterPatch(sdk.Patch{
	// 	ID:      "example_patch",
	// 	Library: "libtarget.so",
	// 	RVA:     0x1234,
	// 	Expect:  []byte{0x00},
	// 	Replace: []byte{0x01},
	// })
}
`
}
