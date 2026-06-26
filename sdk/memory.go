package sdk

type Memory struct{}

func (m *Memory) PatchRVA(id string, library string, rva uintptr, expect []byte, replace []byte) error {
	return nil
}

func (m *Memory) Restore(id string) error {
	return nil
}
