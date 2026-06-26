package mem

type PatchSpec struct {
	ID      string
	Library string
	RVA     uintptr
	Expect  []byte
	Replace []byte
	Startup bool
}

func ResolveAddress(module Module, rva uintptr) uintptr {
	return module.Base + rva
}
