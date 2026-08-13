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
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPNGToICO_WrapsSinglePNGFrame(t *testing.T) {
	t.Parallel()

	src := testPNG(t, 16)

	ico, err := pngToICO(src)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(ico), 22)
	assert.Equal(t, uint16(0), binary.LittleEndian.Uint16(ico[0:2]))
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(ico[2:4]))
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(ico[4:6]))
	assert.Equal(t, byte(16), ico[6])
	assert.Equal(t, byte(16), ico[7])
	assert.Equal(t, uint32(len(src)), binary.LittleEndian.Uint32(ico[14:18]))

	offset := binary.LittleEndian.Uint32(ico[18:22])
	assert.Equal(t, uint32(22), offset)
	assert.Equal(t, src, ico[offset:])
}

func TestPNGToICO_Writes256Sentinel(t *testing.T) {
	t.Parallel()

	src := testPNG(t, 256)

	ico, err := pngToICO(src)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(ico), 22)
	assert.Equal(t, byte(0), ico[6])
	assert.Equal(t, byte(0), ico[7])

	offset := binary.LittleEndian.Uint32(ico[18:22])
	cfg, err := png.DecodeConfig(bytes.NewReader(ico[offset:]))
	require.NoError(t, err)
	assert.Equal(t, 256, cfg.Width)
	assert.Equal(t, 256, cfg.Height)
}

func TestPNGToICO_RejectsOversizedPNG(t *testing.T) {
	t.Parallel()

	src := testPNG(t, 1200)

	ico, err := pngToICO(src)
	require.Error(t, err)
	assert.Nil(t, ico)
	assert.ErrorContains(t, err, "must be 1-256")
}

func TestPNGToMultiSizeICO_WritesFourFrames(t *testing.T) {
	t.Parallel()

	src := testPNG(t, 256)

	ico, err := PNGToMultiSizeICO(src, 16, 32, 48, 256)
	require.NoError(t, err)

	assert.Equal(t, uint16(4), binary.LittleEndian.Uint16(ico[4:6]))
	assert.Equal(t, byte(16), ico[6])
	assert.Equal(t, byte(32), ico[22])
	assert.Equal(t, byte(48), ico[38])
	assert.Equal(t, byte(0), ico[54])

	offset := binary.LittleEndian.Uint32(ico[66:70])
	cfg, err := png.DecodeConfig(bytes.NewReader(ico[offset:]))
	require.NoError(t, err)
	assert.Equal(t, 256, cfg.Width)
	assert.Equal(t, 256, cfg.Height)
}

func TestPNGToMultiSizeICO_RejectsInvalidSize(t *testing.T) {
	t.Parallel()

	src := testPNG(t, 16)

	t.Run(
		"size 257",
		func(t *testing.T) {
			t.Parallel()

			ico, err := PNGToMultiSizeICO(src, 257)
			require.Error(t, err)
			assert.Nil(t, ico)
			assert.ErrorContains(t, err, "must be 1-256")
		},
	)

	t.Run(
		"size 0",
		func(t *testing.T) {
			t.Parallel()

			ico, err := PNGToMultiSizeICO(src, 0)
			require.Error(t, err)
			assert.Nil(t, ico)
			assert.ErrorContains(t, err, "must be 1-256")
		},
	)
}

func testPNG(t *testing.T, size int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			img.SetRGBA(x, y, color.RGBA{R: 0x20, G: 0xc0, B: 0x40, A: 0xff})
		}
	}

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}
