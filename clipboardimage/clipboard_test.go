package clipboardimage

import (
	"errors"
	"testing"
)

func TestErrNoImage(t *testing.T) {
	if !errors.Is(ErrNoImage, ErrNoImage) {
		t.Fatal("sentinel error must be comparable")
	}
}
