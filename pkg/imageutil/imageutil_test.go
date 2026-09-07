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

package imageutil_test

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/imageutil"
)

func pngBytes(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}

func jpegBytes(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))

	return buf.Bytes()
}

func pngHeader(width, height uint32) []byte {
	var ihdr bytes.Buffer

	_ = binary.Write(&ihdr, binary.BigEndian, width)
	_ = binary.Write(&ihdr, binary.BigEndian, height)
	ihdr.Write([]byte{8, 2, 0, 0, 0})

	payload := ihdr.Bytes()
	crc := crc32.ChecksumIEEE(append([]byte("IHDR"), payload...))

	var buf bytes.Buffer
	buf.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(payload)))
	buf.WriteString("IHDR")
	buf.Write(payload)
	_ = binary.Write(&buf, binary.BigEndian, crc)

	return buf.Bytes()
}

func TestFit_ScalesLongestEdge(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 800, 600))
	got := imageutil.Fit(src, 512)

	assert.Equal(t, 512, got.Bounds().Dx())
	assert.Equal(t, 384, got.Bounds().Dy())
}

func TestFit_LeavesSmallImageUnchanged(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 100, 80))
	got := imageutil.Fit(src, 512)

	assert.Same(t, src, got)
}

func TestFit_PanicsOnNonPositiveMaxEdge(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 10, 10))

	assert.Panics(
		t,
		func() {
			imageutil.Fit(src, 0)
		},
	)
}

func TestDownscale_KeepsPNGWithinMaxEdge(t *testing.T) {
	t.Parallel()

	result, err := imageutil.Downscale(bytes.NewReader(pngBytes(t, 800, 600)), 512)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, imageutil.FormatPNG, result.Format)
	assert.Equal(t, imageutil.ContentTypePNG, result.ContentType)

	img, format, err := image.Decode(bytes.NewReader(result.Bytes))
	require.NoError(t, err)
	assert.Equal(t, imageutil.FormatPNG, format)
	assert.Equal(t, 512, img.Bounds().Dx())
	assert.Equal(t, 384, img.Bounds().Dy())
}

func TestDownscale_KeepsSmallJPEGDimensions(t *testing.T) {
	t.Parallel()

	result, err := imageutil.Downscale(bytes.NewReader(jpegBytes(t, 64, 48)), 512)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, imageutil.FormatJPEG, result.Format)
	assert.Equal(t, imageutil.ContentTypeJPEG, result.ContentType)

	img, format, err := image.Decode(bytes.NewReader(result.Bytes))
	require.NoError(t, err)
	assert.Equal(t, imageutil.FormatJPEG, format)
	assert.Equal(t, 64, img.Bounds().Dx())
	assert.Equal(t, 48, img.Bounds().Dy())
}

func TestDownscale_RejectsGIF(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))

	var buf bytes.Buffer
	require.NoError(t, gif.Encode(&buf, src, nil))

	_, err := imageutil.Downscale(&buf, 512)
	require.Error(t, err)
}

func TestDecode_RejectsTooManyPixels(t *testing.T) {
	t.Parallel()

	_, _, err := imageutil.Decode(bytes.NewReader(pngHeader(5000, 5000)))
	require.Error(t, err)
	assert.ErrorIs(t, err, imageutil.ErrTooManyPixels)
}

func TestEncode_RejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	err := imageutil.Encode(&bytes.Buffer{}, src, "gif")
	require.Error(t, err)
	assert.ErrorIs(t, err, imageutil.ErrUnsupportedFormat)
}
