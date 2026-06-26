package sdk

type Context struct {
	Menu   *Menu
	Memory *Memory
	Logf   func(format string, args ...any)
}

func NewContext() *Context {
	return &Context{
		Menu:   &Menu{},
		Memory: &Memory{},
		Logf:   func(format string, args ...any) {},
	}
}
