package runtime

import (
	"strconv"
	"strings"

	"github.com/j0j1j2/gogi/payload/mem"
)

var demoTargetValueRVAHex = ""

func demoPatchSpec() (mem.PatchSpec, bool) {
	rva := parseDemoRVA()
	if rva == 0 {
		return mem.PatchSpec{}, false
	}
	return mem.PatchSpec{
		ID:      "target_value_42",
		Library: "libtarget.so",
		RVA:     rva,
		Expect:  []byte{0x07, 0x00, 0x00, 0x00},
		Replace: []byte{0x2a, 0x00, 0x00, 0x00},
	}, true
}

func parseDemoRVA() uintptr {
	value := strings.TrimPrefix(strings.TrimSpace(demoTargetValueRVAHex), "0x")
	parsed, err := strconv.ParseUint(value, 16, 64)
	if err != nil {
		return 0
	}
	return uintptr(parsed)
}
