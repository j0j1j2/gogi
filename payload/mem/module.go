package mem

func FindModule(mods []Module, name string) (Module, bool) {
	for _, mod := range mods {
		if mod.Name == name || mod.Path == name {
			return mod, true
		}
	}
	return Module{}, false
}

func FindModuleContaining(mods []Module, name string, addr uintptr) (Module, bool) {
	for _, mod := range mods {
		if (mod.Name == name || mod.Path == name) && mod.Contains(addr) {
			return mod, true
		}
	}
	return Module{}, false
}

func (m Module) Contains(addr uintptr) bool {
	return addr >= m.Base && addr < m.End
}
