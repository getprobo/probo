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
)

const (
	jpegSOI            = 0xD8
	jpegEOI            = 0xD9
	jpegSOS            = 0xDA
	jpegAPP1           = 0xE1
	exifOrientationTag = 0x0112
	tiffTypeShort      = 3
)

var exifHeader = []byte("Exif\x00\x00")

// jpegOrientation returns the EXIF Orientation tag (1–8) from a JPEG.
// Missing or unreadable metadata yields 1 (identity). image.Decode does
// not apply this tag, and jpeg.Encode drops APP1, so callers must rotate
// pixels before re-encoding.
func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != jpegSOI {
		return 1
	}

	i := 2
	for i < len(data) {
		if data[i] != 0xFF {
			return 1
		}

		i++
		for i < len(data) && data[i] == 0xFF {
			i++
		}

		if i >= len(data) {
			return 1
		}

		marker := data[i]
		i++

		if marker == jpegSOI || marker == jpegEOI || (marker >= 0xD0 && marker <= 0xD7) {
			continue
		}

		if marker == jpegSOS {
			return 1
		}

		if i+2 > len(data) {
			return 1
		}

		segLen := int(data[i])<<8 | int(data[i+1])
		if segLen < 2 || i+segLen > len(data) {
			return 1
		}

		if marker == jpegAPP1 {
			if orientation := exifOrientation(data[i+2 : i+segLen]); orientation != 0 {
				return orientation
			}
		}

		i += segLen
	}

	return 1
}

func exifOrientation(payload []byte) int {
	if !bytes.HasPrefix(payload, exifHeader) {
		return 0
	}

	return tiffOrientation(payload[len(exifHeader):])
}

func tiffOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 0
	}

	var order binary.ByteOrder

	switch string(tiff[0:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 0
	}

	if order.Uint16(tiff[2:4]) != 0x002A {
		return 0
	}

	ifdOff := int(order.Uint32(tiff[4:8]))
	if ifdOff < 8 || ifdOff+2 > len(tiff) {
		return 0
	}

	n := int(order.Uint16(tiff[ifdOff : ifdOff+2]))
	entry := ifdOff + 2

	for range n {
		if entry+12 > len(tiff) {
			return 0
		}

		tag := order.Uint16(tiff[entry : entry+2])
		typ := order.Uint16(tiff[entry+2 : entry+4])
		count := order.Uint32(tiff[entry+4 : entry+8])

		if tag == exifOrientationTag && typ == tiffTypeShort && count == 1 {
			orientation := int(order.Uint16(tiff[entry+8 : entry+10]))
			if orientation >= 1 && orientation <= 8 {
				return orientation
			}

			return 0
		}

		entry += 12
	}

	return 0
}

func applyOrientation(img image.Image, orientation int) image.Image {
	if orientation <= 1 || orientation > 8 {
		return img
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	dstW, dstH := width, height
	if orientation >= 5 {
		dstW, dstH = height, width
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for y := range height {
		for x := range width {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			dx, dy := orientedXY(x, y, width, height, orientation)
			dst.Set(dx, dy, c)
		}
	}

	return dst
}

func orientedXY(x, y, width, height, orientation int) (int, int) {
	switch orientation {
	case 2:
		return width - 1 - x, y
	case 3:
		return width - 1 - x, height - 1 - y
	case 4:
		return x, height - 1 - y
	case 5:
		return y, x
	case 6:
		return height - 1 - y, x
	case 7:
		return height - 1 - y, width - 1 - x
	case 8:
		return y, width - 1 - x
	default:
		return x, y
	}
}
