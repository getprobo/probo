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
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTintStatusIcon_SetsRGBKeepsAlpha(t *testing.T) {
	t.Parallel()

	src := testGrayAlphaPNG(t, 4, []struct {
		x, y int
		a    uint8
	}{
		{1, 1, 255},
		{2, 1, 128},
	})

	tint := color.RGBA{R: 52, G: 199, B: 89, A: 255}

	out, err := tintStatusIcon(src, tint, 4)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, 4, img.Bounds().Dx())
	assert.Equal(t, 4, img.Bounds().Dy())

	opaque := color.NRGBAModel.Convert(img.At(1, 1)).(color.NRGBA)
	assert.Equal(t, tint.R, opaque.R)
	assert.Equal(t, tint.G, opaque.G)
	assert.Equal(t, tint.B, opaque.B)
	assert.Equal(t, uint8(255), opaque.A)

	partial := color.NRGBAModel.Convert(img.At(2, 1)).(color.NRGBA)
	assert.Equal(t, tint.R, partial.R)
	assert.Equal(t, tint.G, partial.G)
	assert.Equal(t, tint.B, partial.B)
	assert.Equal(t, uint8(128), partial.A)

	clear := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA)
	assert.Equal(t, uint8(0), clear.A)
}

func TestTintStatusIcon_TintsCommittedTemplate(t *testing.T) {
	t.Parallel()

	template, err := os.ReadFile("icon_template.png")
	require.NoError(t, err)

	tint := color.RGBA{R: 52, G: 199, B: 89, A: 255}

	out, err := tintStatusIcon(template, tint, 36)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, 36, img.Bounds().Dx())

	found := false

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if px.A != 255 {
				continue
			}

			found = true

			assert.Equal(t, tint.R, px.R)
			assert.Equal(t, tint.G, px.G)
			assert.Equal(t, tint.B, px.B)
		}
	}

	assert.True(t, found)
}

func TestTintStatusIcon_ResizesToRequestedSize(t *testing.T) {
	t.Parallel()

	src := testGrayAlphaPNG(t, 8, []struct {
		x, y int
		a    uint8
	}{
		{3, 3, 255},
	})

	out, err := tintStatusIcon(src, color.RGBA{R: 255, G: 149, B: 0, A: 255}, 16)
	require.NoError(t, err)

	cfg, err := png.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, 16, cfg.Width)
	assert.Equal(t, 16, cfg.Height)
}

func testGrayAlphaPNG(
	t *testing.T,
	size int,
	opaque []struct {
		x, y int
		a    uint8
	},
) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for _, p := range opaque {
		img.SetNRGBA(p.x, p.y, color.NRGBA{A: p.a})
	}

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}
