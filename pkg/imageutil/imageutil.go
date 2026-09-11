// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package imageutil

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

const (
	FormatJPEG = "jpeg"
	FormatPNG  = "png"

	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"

	jpegQuality = 85

	// 4096². DecodeConfig is checked before allocating pixels so a
	// compressed bomb cannot expand past this.
	maxPixels = 16_777_216
)

var (
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrInvalidDimensions = errors.New("invalid image dimensions")
	ErrTooManyPixels     = errors.New("image exceeds pixel limit")
)

type (
	Result struct {
		Bytes       []byte
		ContentType string
		Format      string
	}
)

// Decode reads a JPEG or PNG image from r. Other formats are rejected
// even if a decoder is registered in the process. Metadata (EXIF, GPS,
// orientation, comments) is ignored; Encode writes pixels only.
func Decode(r io.Reader) (image.Image, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("cannot read image: %w", err)
	}

	br := bytes.NewReader(data)

	cfg, format, err := image.DecodeConfig(br)
	if err != nil {
		return nil, "", fmt.Errorf("cannot decode image config: %w", err)
	}

	if format != FormatJPEG && format != FormatPNG {
		return nil, "", fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}

	if cfg.Width < 1 || cfg.Height < 1 {
		return nil, "", fmt.Errorf("%w: %dx%d", ErrInvalidDimensions, cfg.Width, cfg.Height)
	}

	pixels := int64(cfg.Width) * int64(cfg.Height)
	if pixels > maxPixels {
		return nil, "", fmt.Errorf("%w: %dx%d", ErrTooManyPixels, cfg.Width, cfg.Height)
	}

	if _, err := br.Seek(0, io.SeekStart); err != nil {
		return nil, "", fmt.Errorf("cannot rewind image: %w", err)
	}

	img, _, err := image.Decode(br)
	if err != nil {
		return nil, "", fmt.Errorf("cannot decode image: %w", err)
	}

	return img, format, nil
}

// Fit scales img so its longest edge is at most maxEdge, preserving
// aspect ratio. Images already within the limit are returned unchanged.
func Fit(img image.Image, maxEdge int) image.Image {
	if maxEdge < 1 {
		panic("imageutil: maxEdge must be positive")
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	fittedWidth, fittedHeight := fitSize(width, height, maxEdge)

	if fittedWidth == width && fittedHeight == height {
		return img
	}

	dst := image.NewRGBA(image.Rect(0, 0, fittedWidth, fittedHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)

	return dst
}

// Encode writes img to w in format (jpeg or png) with no metadata
// segments or ancillary chunks.
func Encode(w io.Writer, img image.Image, format string) error {
	switch format {
	case FormatJPEG:
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return fmt.Errorf("cannot encode jpeg: %w", err)
		}

		return nil
	case FormatPNG:
		if err := png.Encode(w, img); err != nil {
			return fmt.Errorf("cannot encode png: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
}

// Downscale decodes a JPEG or PNG from r, scales it so the longest edge
// is at most maxEdge, and re-encodes it in the original format.
func Downscale(r io.Reader, maxEdge int) (*Result, error) {
	img, format, err := Decode(r)
	if err != nil {
		return nil, err
	}

	img = Fit(img, maxEdge)

	var buf bytes.Buffer
	if err := Encode(&buf, img, format); err != nil {
		return nil, err
	}

	return &Result{
		Bytes:       buf.Bytes(),
		ContentType: contentType(format),
		Format:      format,
	}, nil
}

func fitSize(width, height, maxEdge int) (int, int) {
	if width <= maxEdge && height <= maxEdge {
		return width, height
	}

	if width >= height {
		return maxEdge, max(1, height*maxEdge/width)
	}

	return max(1, width*maxEdge/height), maxEdge
}

func contentType(format string) string {
	switch format {
	case FormatJPEG:
		return ContentTypeJPEG
	case FormatPNG:
		return ContentTypePNG
	default:
		return ""
	}
}
