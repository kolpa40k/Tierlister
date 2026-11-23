//go:build windows

package clipboardimage

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/png"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/image/bmp"
)

const (
	cfDIB   = 8
	cfDIBV5 = 17
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	openClipboard              = user32.NewProc("OpenClipboard")
	closeClipboard             = user32.NewProc("CloseClipboard")
	getClipboardData           = user32.NewProc("GetClipboardData")
	isClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	registerClipboardFormat    = user32.NewProc("RegisterClipboardFormatW")
	globalLock                 = kernel32.NewProc("GlobalLock")
	globalUnlock               = kernel32.NewProc("GlobalUnlock")
	globalSize                 = kernel32.NewProc("GlobalSize")
)

// Win32 clipboard ownership is thread-affine: Open/Close must run on the same OS thread.
func readPlatformImage(ctx context.Context) ([]byte, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opened, _, err := openClipboard.Call(0)
	if opened == 0 {
		return nil, fmt.Errorf("open Windows clipboard: %w", err)
	}
	defer closeClipboard.Call()

	pngName, _ := syscall.UTF16PtrFromString("PNG")
	pngFormat, _, _ := registerClipboardFormat.Call(uintptr(unsafe.Pointer(pngName)))
	if pngFormat != 0 && available(pngFormat) {
		data, readErr := readGlobalData(pngFormat)
		if readErr == nil {
			return data, nil
		}
	}
	for _, format := range []uintptr{cfDIBV5, cfDIB} {
		if !available(format) {
			continue
		}
		data, readErr := readGlobalData(format)
		if readErr != nil {
			return nil, readErr
		}
		return dibToPNG(data)
	}
	return nil, ErrNoImage
}

func available(format uintptr) bool {
	result, _, _ := isClipboardFormatAvailable.Call(format)
	return result != 0
}

func readGlobalData(format uintptr) ([]byte, error) {
	handle, _, err := getClipboardData.Call(format)
	if handle == 0 {
		return nil, fmt.Errorf("get clipboard data: %w", err)
	}
	size, _, _ := globalSize.Call(handle)
	if size == 0 {
		return nil, errors.New("empty clipboard data")
	}
	// DIB may be larger than the encoded PNG; limit memory allocation.
	if size > 160<<20 {
		return nil, errors.New("clipboard bitmap is too large")
	}
	ptr, _, err := globalLock.Call(handle)
	if ptr == 0 {
		return nil, fmt.Errorf("lock clipboard data: %w", err)
	}
	defer globalUnlock.Call(handle)
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size))), nil
}

// A CF_DIB contains a DIB header and pixels, but no BMP file header.
// Convert it to a BMP stream, decode, then encode as PNG for the shared importer.
func dibToPNG(dib []byte) ([]byte, error) {
	if len(dib) < 40 {
		return nil, errors.New("invalid DIB header")
	}
	headerSize := binary.LittleEndian.Uint32(dib[:4])
	if headerSize < 40 || headerSize > uint32(len(dib)) {
		return nil, errors.New("invalid DIB header size")
	}
	bits := binary.LittleEndian.Uint16(dib[14:16])
	compression := binary.LittleEndian.Uint32(dib[16:20])
	if bits != 24 && bits != 32 {
		return nil, fmt.Errorf("unsupported clipboard bitmap depth: %d", bits)
	}
	if compression != 0 && compression != 3 {
		return nil, fmt.Errorf("unsupported clipboard bitmap compression: %d", compression)
	}
	offset := uint64(headerSize)
	// BITMAPINFOHEADER bitfield masks follow the header; V4/V5 include masks inside it.
	if compression == 3 && headerSize == 40 {
		offset += 12
	}
	if offset > uint64(len(dib)) {
		return nil, errors.New("truncated clipboard bitmap")
	}
	if uint64(len(dib))+14 > 0xffffffff {
		return nil, errors.New("clipboard bitmap too large")
	}
	bmpData := make([]byte, 14+len(dib))
	copy(bmpData[14:], dib)
	copy(bmpData[:2], "BM")
	binary.LittleEndian.PutUint32(bmpData[2:6], uint32(len(bmpData)))
	binary.LittleEndian.PutUint32(bmpData[10:14], uint32(14+offset))
	img, err := bmp.Decode(bytes.NewReader(bmpData))
	if err != nil {
		return nil, fmt.Errorf("decode Windows clipboard bitmap: %w", err)
	}
	if err := checkDimensions(img); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func checkDimensions(img image.Image) error {
	b := img.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 || int64(b.Dx())*int64(b.Dy()) > 40_000_000 {
		return errors.New("clipboard image dimensions exceed safety limit")
	}
	return nil
}
