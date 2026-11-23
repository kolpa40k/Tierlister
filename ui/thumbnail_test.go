package ui

import (
	"image"
	"image/color"
	"testing"
)

func TestThumbnailPreservesAlpha(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	src.SetNRGBA(32, 32, color.NRGBA{R: 255, A: 128})
	thumbnail := thumbnailFromImage(src, 64, 64, "cover")
	if thumbnail == nil {
		t.Fatal("nil thumbnail")
	}
	got, ok := thumbnail.Image.(*image.NRGBA)
	if !ok {
		t.Fatalf("image type %T", thumbnail.Image)
	}
	if a := got.NRGBAAt(0, 0).A; a != 0 {
		t.Errorf("transparent pixel alpha = %d", a)
	}
	if a := got.NRGBAAt(32, 32).A; a != 128 {
		t.Errorf("semi-transparent pixel alpha = %d", a)
	}
}

func TestThumbnailContainHasTransparentPadding(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	thumbnail := thumbnailFromImage(src, 100, 100, "contain")
	if thumbnail == nil {
		t.Fatal("nil thumbnail")
	}
	got := thumbnail.Image.(*image.NRGBA)
	if got.NRGBAAt(50, 0).A != 0 {
		t.Fatal("top padding is not transparent")
	}
	if got.NRGBAAt(50, 50).A != 255 {
		t.Fatal("image center not opaque")
	}
	if got.NRGBAAt(50, 99).A != 0 {
		t.Fatal("bottom padding is not transparent")
	}
}

func TestThumbnailCoverFillsCard(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 200, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 200; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	thumbnail := thumbnailFromImage(src, 100, 100, "cover")
	if thumbnail == nil {
		t.Fatal("nil thumbnail")
	}
	got := thumbnail.Image.(*image.NRGBA)
	if got.Bounds().Dx() != 100 || got.Bounds().Dy() != 100 {
		t.Fatal("wrong dimensions")
	}
	if got.NRGBAAt(0, 0).A != 255 {
		t.Fatal("cover left a transparent border")
	}
}
