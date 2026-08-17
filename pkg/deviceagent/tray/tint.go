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

package tray

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

func mustTintStatusIcon(templatePNG []byte, tint color.RGBA) []byte {
	out, err := tintStatusIcon(templatePNG, tint, 16)
	if err != nil {
		return templatePNG
	}

	return out
}

func tintStatusIcon(templatePNG []byte, tint color.RGBA, size int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(templatePNG))
	if err != nil {
		return nil, fmt.Errorf("cannot decode template png: %w", err)
	}

	b := src.Bounds()
	tinted := image.NewNRGBA(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			tinted.SetNRGBA(
				x,
				y,
				color.NRGBA{
					R: tint.R,
					G: tint.G,
					B: tint.B,
					A: uint8(a >> 8),
				},
			)
		}
	}

	if size == b.Dx() && size == b.Dy() {
		var buf bytes.Buffer
		if err := png.Encode(&buf, tinted); err != nil {
			return nil, fmt.Errorf("cannot encode tinted png: %w", err)
		}

		return buf.Bytes(), nil
	}

	return resizePNG(tinted, size)
}
