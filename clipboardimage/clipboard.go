// Package clipboardimage provides platform-specific image clipboard access.
package clipboardimage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const maxBytes = 25 << 20

var ErrNoImage = errors.New("clipboard does not contain a supported image")

// ReadImage returns a validated encoded image from the native clipboard.
func ReadImage(ctx context.Context) ([]byte, error) {
	data, err := readPlatformImage(ctx)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrNoImage
	}
	if len(data) > maxBytes {
		return nil, errors.New("clipboard image exceeds 25 MiB")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("clipboard image is invalid: %w", err)
	}
	if cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 40_000_000 {
		return nil, errors.New("clipboard image dimensions exceed safety limit")
	}
	return data, nil
}
