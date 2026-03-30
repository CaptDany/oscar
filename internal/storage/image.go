package storage

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/disintegration/imaging"
)

const (
	AvatarSize  = 200
	JpegQuality = 80
)

func CropAndResizeToSquare(reader io.Reader, size int) (io.Reader, error) {
	src, err := imaging.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var cropped image.Image

	if width > height {
		cropped = imaging.Crop(src, image.Rect(0, 0, height, height))
	} else if height > width {
		cropped = imaging.Crop(src, image.Rect(0, 0, width, width))
	} else {
		cropped = src
	}

	resized := imaging.Resize(cropped, size, size, imaging.Lanczos)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(JpegQuality))
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return &buf, nil
}
