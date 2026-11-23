package storage

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ImportImageBytes validates and stores an image from a clipboard or other byte source.
// It uses the same private image directory and thumbnail format as file imports.
func ImportImageBytes(data []byte, id string) (string, error) {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\\`) {
		return "", errors.New("invalid image ID")
	}
	if len(data) == 0 || len(data) > 25<<20 {
		return "", errors.New("image must be between 1 byte and 25 MiB")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image header: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 40_000_000 {
		return "", errors.New("image dimensions exceed safety limit")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}
	bounds := src.Bounds()
	const size = 200
	scale := min(float64(size)/float64(bounds.Dx()), float64(size)/float64(bounds.Dy()))
	w := max(1, min(size, int(float64(bounds.Dx())*scale)))
	h := max(1, min(size, int(float64(bounds.Dy())*scale)))
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	rect := image.Rect((size-w)/2, (size-h)/2, (size-w)/2+w, (size-h)/2+h)
	draw.ApproxBiLinear.Scale(dst, rect, src, bounds, draw.Src, nil)
	dir, err := imagesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".image-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	encodeErr := png.Encode(tmp, dst)
	if encodeErr == nil {
		encodeErr = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err := errors.Join(encodeErr, closeErr); err != nil {
		return "", fmt.Errorf("save image: %w", err)
	}
	path := filepath.Join(dir, id+".png")
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}
