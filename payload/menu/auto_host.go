//go:build !android

package menu

import (
	"errors"
	"unsafe"
)

var ErrAutoAttachUnavailable = errors.New("auto_attach_unavailable")

func AttachAuto(vm unsafe.Pointer, url string, configJSON string) error {
	return ErrAutoAttachUnavailable
}
