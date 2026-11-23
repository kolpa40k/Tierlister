package storage

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"
	"tierlist-app/models"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const dataFile = "categories.json"

func dataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tierlist-app", dataFile), nil
}

func LoadCategories() ([]models.Category, error) {
	path, err := dataPath()
	if err != nil {
		return nil, err
	}
	var categories []models.Category
	if err := readJSON(path, &categories); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.Category{}, nil
		}
		return nil, err
	}
	if categories == nil {
		return []models.Category{}, nil
	}
	return categories, nil
}

func SaveCategories(categories []models.Category) error {
	path, err := dataPath()
	if err != nil {
		return err
	}
	return writeJSONAtomic(path, categories)
}

// ImportImage imports an image to private storage, preserving the original on failure.
func ImportImage(srcPath, id string) (string, error) {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\\`) {
		return "", errors.New("invalid image ID")
	}
	dir, err := imagesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	srcFile, err := OpenFileAtPath(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source image: %w", err)
	}
	srcImg, _, decodeErr := image.Decode(srcFile)
	closeErr := srcFile.Close()
	if decodeErr != nil {
		return "", fmt.Errorf("decode image: %w", decodeErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("close source image: %w", closeErr)
	}
	bounds := srcImg.Bounds()
	if bounds.Empty() {
		return "", errors.New("empty image")
	}
	const size = 200
	scale := min(float64(size)/float64(bounds.Dx()), float64(size)/float64(bounds.Dy()))
	w := max(1, min(size, int(float64(bounds.Dx())*scale)))
	h := max(1, min(size, int(float64(bounds.Dy())*scale)))
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	rect := image.Rect((size-w)/2, (size-h)/2, (size-w)/2+w, (size-h)/2+h)
	draw.ApproxBiLinear.Scale(dst, rect, srcImg, bounds, draw.Src, nil)

	tmp, err := os.CreateTemp(dir, ".image-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return "", err
	}
	encodeErr := png.Encode(tmp, dst)
	if encodeErr == nil {
		encodeErr = tmp.Sync()
	}
	closeErr = tmp.Close()
	if err := errors.Join(encodeErr, closeErr); err != nil {
		return "", fmt.Errorf("encode thumbnail: %w", err)
	}
	dstPath := filepath.Join(dir, id+".png")
	if err := os.Rename(tmp.Name(), dstPath); err != nil {
		return "", fmt.Errorf("store thumbnail: %w", err)
	}
	return dstPath, nil
}

func DeleteImage(path string) {
	if path == "" {
		return
	}
	dir, err := imagesDir()
	if err != nil || !isInside(path, dir) {
		return
	}
	// Never follow symlinks when deleting an image owned by the application.
	_ = os.Remove(path)
}

// DeleteImageChecked removes only application-owned images and reports failures.
func DeleteImageChecked(path string) error {
	if path == "" {
		return nil
	}
	dir, err := imagesDir()
	if err != nil {
		return err
	}
	if !isInside(path, dir) {
		return errors.New("image is outside application storage")
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("not a regular image file")
	}
	return os.Remove(path)
}
