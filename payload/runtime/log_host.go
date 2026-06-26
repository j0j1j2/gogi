//go:build !android

package runtime

import "log"

func Logf(format string, args ...any) {
	log.Printf(format, args...)
}
