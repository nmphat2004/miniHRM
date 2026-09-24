package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestValidateAvatar(t *testing.T) {
	var buf bytes.Buffer
	imageData := image.NewRGBA(image.Rect(0, 0, 2, 2))
	imageData.Set(0, 0, color.RGBA{R: 220, A: 255})
	if err := png.Encode(&buf, imageData); err != nil {
		t.Fatal(err)
	}
	valid := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	if err := validateAvatar(valid); err != nil {
		t.Fatalf("valid PNG rejected: %v", err)
	}
	if err := validateAvatar(""); err != nil {
		t.Fatalf("delete rejected: %v", err)
	}
	for _, tc := range []struct{ name, value string }{
		{"invalid base64", "data:image/png;base64,invalid!"},
		{"mismatched mime", strings.Replace(valid, "image/png", "image/jpeg", 1)},
		{"oversized", "data:image/png;base64," + strings.Repeat("A", base64.StdEncoding.EncodedLen(64*1024)+4)},
		{"remote URL", "https://example.com/avatar.png"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateAvatar(tc.value); !errors.Is(err, ErrInvalidAvatar) {
				t.Errorf("got %v, want ErrInvalidAvatar", err)
			}
		})
	}
}
