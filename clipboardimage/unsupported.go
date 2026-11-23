//go:build !linux && !windows

package clipboardimage

import (
	"context"
	"errors"
)

func readPlatformImage(context.Context) ([]byte, error) {
	return nil, errors.New("image clipboard is not supported on this platform")
}
