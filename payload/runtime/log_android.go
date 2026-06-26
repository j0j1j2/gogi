//go:build android

package runtime

/*
#include <android/log.h>
#include <stdlib.h>
static void gogi_log(const char* msg) {
	__android_log_write(3, "gogi", msg);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func Logf(format string, args ...any) {
	msg := C.CString(fmt.Sprintf(format, args...))
	defer C.free(unsafe.Pointer(msg))
	C.gogi_log(msg)
}
