package mem

func FindModule(mods []Module, name string) (Module, bool) {
	for _, mod := range mods {
		if mod.Name == name || mod.Path == name {
			return mod, true
		}
	}
	return Module{}, false
}

func (m Module) Contains(addr uintptr) bool {
	return addr >= m.Base && addr < m.End
}
