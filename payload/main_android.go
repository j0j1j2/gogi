//go:build android

package main

/*
#include <jni.h>
*/
import "C"

import (
	"unsafe"

	gogiruntime "gogi/payload/runtime"
)

//export ModInit
func ModInit() {
	gogiruntime.Start(nil)
}

//export JNI_OnLoad
func JNI_OnLoad(vm *C.JavaVM, reserved unsafe.Pointer) C.jint {
	gogiruntime.Start(unsafe.Pointer(vm))
	return C.JNI_VERSION_1_6
}

func main() {}
