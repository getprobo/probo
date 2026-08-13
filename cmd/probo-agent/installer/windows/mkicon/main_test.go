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

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteProductICO_WritesFourFrames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pngPath := filepath.Join(dir, "icon.png")
	icoPath := filepath.Join(dir, "icon.ico")

	require.NoError(t, os.WriteFile(pngPath, testPNG(t, 64), 0o644))

	err := writeProductICO(pngPath, icoPath)
	require.NoError(t, err)

	ico, err := os.ReadFile(icoPath)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ico), 6)
	assert.Equal(t, uint16(0), binary.LittleEndian.Uint16(ico[0:2]))
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(ico[2:4]))
	assert.Equal(t, uint16(4), binary.LittleEndian.Uint16(ico[4:6]))
	assert.Equal(t, byte(16), ico[6])
	assert.Equal(t, byte(32), ico[22])
	assert.Equal(t, byte(48), ico[38])
	assert.Equal(t, byte(0), ico[54])
}

func TestWriteProductSyso_WritesNonEmptyCOFF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pngPath := filepath.Join(dir, "icon.png")
	sysoPath := filepath.Join(dir, "rsrc_windows_amd64.syso")

	require.NoError(t, os.WriteFile(pngPath, testPNG(t, 64), 0o644))

	err := writeProductSyso(pngPath, sysoPath, "amd64")
	require.NoError(t, err)

	syso, err := os.ReadFile(sysoPath)
	require.NoError(t, err)
	require.Greater(t, len(syso), 64)
	assert.Equal(t, uint16(0x8664), binary.LittleEndian.Uint16(syso[0:2]))
}

func testPNG(t *testing.T, size int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			img.SetRGBA(x, y, color.RGBA{R: 0xE6, G: 0xFF, B: 0x03, A: 0xFF})
		}
	}

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}
