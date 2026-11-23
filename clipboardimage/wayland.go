//go:build linux

package clipboardimage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// readPlatformImage reads an image from the native Wayland clipboard via wl-paste.
// wl-clipboard must be installed and WAYLAND_DISPLAY must be set.
func readPlatformImage(ctx context.Context) ([]byte, error) {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return nil, errors.New("native Wayland clipboard is unavailable (WAYLAND_DISPLAY is empty)")
	}
	if _, err := exec.LookPath("wl-paste"); err != nil {
		return nil, errors.New("install wl-clipboard to paste images on Wayland")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	types, err := exec.CommandContext(ctx, "wl-paste", "--list-types").Output()
	if err != nil {
		return nil, fmt.Errorf("read clipboard formats: %w", err)
	}
	var format string
	for _, candidate := range []string{"image/png", "image/jpeg", "image/webp"} {
		for line := range strings.SplitSeq(string(types), "\n") {
			if strings.TrimSpace(line) == candidate {
				format = candidate
				break
			}
		}
		if format != "" {
			break
		}
	}
	if format == "" {
		return nil, errors.New("clipboard does not contain a PNG, JPEG or WebP image")
	}
	// Keep the executable and its arguments fixed: clipboard-provided MIME
	// types are only used to select one of these known commands.
	var cmd *exec.Cmd
	switch format {
	case "image/png":
		cmd = exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", "image/png")
	case "image/jpeg":
		cmd = exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", "image/jpeg")
	case "image/webp":
		cmd = exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", "image/webp")
	default:
		return nil, errors.New("unsupported clipboard image format")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(stdout, maxBytes+1))
	if len(data) > maxBytes {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if len(data) > maxBytes {
		return nil, errors.New("clipboard image exceeds 25 MiB")
	}
	if readErr != nil {
		return nil, fmt.Errorf("read clipboard image: %w", readErr)
	}
	if waitErr != nil {
		return nil, fmt.Errorf("paste clipboard image: %w", waitErr)
	}
	return data, nil
}
