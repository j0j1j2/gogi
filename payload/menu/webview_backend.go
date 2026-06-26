package menu

import (
	"errors"
	"unsafe"
)

var ErrContextRequired = errors.New("context_required")

type WebViewBackend struct{}

func NewWebViewBackend() *WebViewBackend {
	return &WebViewBackend{}
}

func (b *WebViewBackend) Attach(ctx unsafe.Pointer, url string) error {
	if ctx == nil {
		return ErrContextRequired
	}
	return nil
}
