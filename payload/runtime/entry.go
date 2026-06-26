package runtime

import (
	"net/http"
	"sync"
	"time"
	"unsafe"

	"gogi/payload/control"
	"gogi/payload/mem"
	"gogi/payload/menu"
)

var startOnce sync.Once
var capturedVM unsafe.Pointer

const menuURL = "http://127.0.0.1:17373/"

func Start(vm unsafe.Pointer) {
	startOnce.Do(func() {
		capturedVM = vm
		Logf("gogi runtime started")
		startMenuServer()
		go attachAutoMenuLoop(vm)
		go func() {
			Logf("gogi worker started")
		}()
	})
}

func CapturedVM() unsafe.Pointer {
	return capturedVM
}

func MenuURL() string {
	return menuURL
}

func startMenuServer() {
	registry := control.NewRegistry()
	registry.SetApplier(mem.NewProcessApplier())
	registry.Register(demoPatchSpec())
	server := menu.NewServer(registry)
	go func() {
		if err := http.ListenAndServe("127.0.0.1:17373", server.Handler()); err != nil {
			Logf("gogi menu server stopped: %v", err)
		}
	}()
}

func attachAutoMenu(vm unsafe.Pointer) {
	if vm == nil {
		Logf("gogi auto menu skipped: no JavaVM")
		return
	}
	if err := menu.AttachAuto(vm, MenuURL()); err != nil {
		Logf("gogi auto menu attach failed: %v", err)
		return
	}
	Logf("gogi auto menu attached")
}

func attachAutoMenuLoop(vm unsafe.Pointer) {
	for attempt := 1; attempt <= 20; attempt++ {
		if vm == nil {
			Logf("gogi auto menu skipped: no JavaVM")
			return
		}
		if err := menu.AttachAuto(vm, MenuURL()); err != nil {
			Logf("gogi auto menu attach attempt %d failed: %v", attempt, err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		Logf("gogi auto menu attached")
		return
	}
}
