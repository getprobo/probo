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
	"fmt"
	"image"
	"image/png"

	"golang.org/x/image/draw"
)

type icoFrame struct {
	width  int
	height int
	png    []byte
}

func pngToICO(pngBytes []byte) ([]byte, error) {
	cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot decode png: %w", err)
	}

	return wrapICO(
		[]icoFrame{
			{
				width:  cfg.Width,
				height: cfg.Height,
				png:    pngBytes,
			},
		},
	)
}

func pngToMultiSizeICO(pngBytes []byte, sizes ...int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot decode png: %w", err)
	}

	frames := make([]icoFrame, 0, len(sizes))
	for _, size := range sizes {
		if err := validateICODimension(size); err != nil {
			return nil, err
		}

		framePNG, err := resizePNG(src, size)
		if err != nil {
			return nil, err
		}

		frames = append(
			frames,
			icoFrame{
				width:  size,
				height: size,
				png:    framePNG,
			},
		)
	}

	return wrapICO(frames)
}

func resizePNG(src image.Image, size int) ([]byte, error) {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("cannot encode %dx%d png: %w", size, size, err)
	}

	return buf.Bytes(), nil
}

func wrapICO(frames []icoFrame) ([]byte, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("cannot wrap ico: no frames")
	}

	for _, frame := range frames {
		if err := validateICODimension(frame.width); err != nil {
			return nil, err
		}

		if err := validateICODimension(frame.height); err != nil {
			return nil, err
		}
	}

	const (
		iconDirLen   = 6
		iconEntryLen = 16
	)

	headerLen := iconDirLen + iconEntryLen*len(frames)

	size := headerLen
	for _, frame := range frames {
		size += len(frame.png)
	}

	buf := make([]byte, 0, size)
	buf = binary.LittleEndian.AppendUint16(buf, 0)
	buf = binary.LittleEndian.AppendUint16(buf, 1)
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(frames)))

	offset := uint32(headerLen)

	for _, frame := range frames {
		buf = append(buf, icoDimension(frame.width), icoDimension(frame.height), 0, 0)
		buf = binary.LittleEndian.AppendUint16(buf, 1)
		buf = binary.LittleEndian.AppendUint16(buf, 32)
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(frame.png)))
		buf = binary.LittleEndian.AppendUint32(buf, offset)
		offset += uint32(len(frame.png))
	}

	for _, frame := range frames {
		buf = append(buf, frame.png...)
	}

	return buf, nil
}

func validateICODimension(n int) error {
	if n < 1 || n > 256 {
		return fmt.Errorf("cannot encode ico dimension %d: must be 1-256", n)
	}

	return nil
}

func icoDimension(n int) byte {
	if n == 256 {
		return 0
	}

	return byte(n)
}
