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
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func redBlueRow() *image.RGBA {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.Set(0, 0, color.RGBA{R: 255, A: 255})
	src.Set(1, 0, color.RGBA{B: 255, A: 255})

	return src
}

func jpegWithOrientation(t *testing.T, img image.Image, orientation uint16) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}))

	encoded := buf.Bytes()
	require.GreaterOrEqual(t, len(encoded), 2)

	app1 := exifAPP1(orientation)
	out := make([]byte, 0, 2+len(app1)+len(encoded)-2)
	out = append(out, 0xFF, 0xD8)
	out = append(out, app1...)
	out = append(out, encoded[2:]...)

	return out
}

func exifAPP1(orientation uint16) []byte {
	var tiff bytes.Buffer
	tiff.WriteString("MM")
	_ = binary.Write(&tiff, binary.BigEndian, uint16(0x002A))
	_ = binary.Write(&tiff, binary.BigEndian, uint32(8))
	_ = binary.Write(&tiff, binary.BigEndian, uint16(1))
	_ = binary.Write(&tiff, binary.BigEndian, uint16(exifOrientationTag))
	_ = binary.Write(&tiff, binary.BigEndian, uint16(tiffTypeShort))
	_ = binary.Write(&tiff, binary.BigEndian, uint32(1))
	_ = binary.Write(&tiff, binary.BigEndian, orientation)
	_ = binary.Write(&tiff, binary.BigEndian, uint16(0))
	_ = binary.Write(&tiff, binary.BigEndian, uint32(0))

	payload := append(append([]byte{}, exifHeader...), tiff.Bytes()...)
	length := uint16(len(payload) + 2)
	out := []byte{0xFF, jpegAPP1, byte(length >> 8), byte(length)}

	return append(out, payload...)
}

func TestApplyOrientation_Rotate90CW(t *testing.T) {
	t.Parallel()

	got := applyOrientation(redBlueRow(), 6)

	assert.Equal(t, 1, got.Bounds().Dx())
	assert.Equal(t, 2, got.Bounds().Dy())

	topR, _, topB, _ := got.At(0, 0).RGBA()
	bottomR, _, bottomB, _ := got.At(0, 1).RGBA()

	assert.Greater(t, topR, topB)
	assert.Greater(t, bottomB, bottomR)
}

func TestApplyOrientation_LeavesIdentityUnchanged(t *testing.T) {
	t.Parallel()

	src := redBlueRow()
	got := applyOrientation(src, 1)

	assert.Same(t, src, got)
}

func TestJpegOrientation_ReadsAPP1(t *testing.T) {
	t.Parallel()

	data := jpegWithOrientation(t, redBlueRow(), 6)

	assert.Equal(t, 6, jpegOrientation(data))
}

func TestJpegOrientation_SkipsFillBytes(t *testing.T) {
	t.Parallel()

	data := jpegWithOrientation(t, redBlueRow(), 6)
	filled := append([]byte{0xFF, jpegSOI, 0xFF, 0xFF}, data[2:]...)

	assert.Equal(t, 6, jpegOrientation(filled))
}

func TestJpegOrientation_DefaultsWhenMissing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, redBlueRow(), &jpeg.Options{Quality: 90}))

	assert.Equal(t, 1, jpegOrientation(buf.Bytes()))
}

func TestDecode_AppliesJPEGOrientation(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 16, 8))

	for y := range 8 {
		for x := range 16 {
			if x < 8 {
				src.Set(x, y, color.RGBA{R: 255, A: 255})
				continue
			}

			src.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}

	img, format, err := Decode(bytes.NewReader(jpegWithOrientation(t, src, 6)))
	require.NoError(t, err)
	assert.Equal(t, FormatJPEG, format)
	assert.Equal(t, 8, img.Bounds().Dx())
	assert.Equal(t, 16, img.Bounds().Dy())

	top := img.At(img.Bounds().Min.X+4, img.Bounds().Min.Y+2)
	bottom := img.At(img.Bounds().Min.X+4, img.Bounds().Min.Y+14)
	topR, _, topB, _ := top.RGBA()
	bottomR, _, bottomB, _ := bottom.RGBA()

	assert.Greater(t, topR, topB)
	assert.Greater(t, bottomB, bottomR)
}
