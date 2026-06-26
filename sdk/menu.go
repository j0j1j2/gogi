package sdk

type Menu struct{}

type ToggleOptions struct {
	Label     string
	OnEnable  func() error
	OnDisable func() error
}

func (m *Menu) Toggle(id string, options ToggleOptions) {
}
