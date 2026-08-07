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

package native

import "fmt"

type booleanDecoder struct {
	reader    *reader
	value     bool
	remaining uint64
}

func newBooleanDecoder(data []byte) *booleanDecoder {
	return &booleanDecoder{
		reader: newReader(data),
		value:  true,
	}
}

func (d *booleanDecoder) next() (bool, error) {
	for d.remaining == 0 {
		if d.reader.done() {
			return false, nil
		}

		count, err := d.reader.readULEB128()
		if err != nil {
			return false, fmt.Errorf("cannot read boolean run: %w", err)
		}
		if count > maxRLERunLength {
			return false, fmt.Errorf("boolean run %d exceeds limit", count)
		}
		d.value = !d.value
		d.remaining = count
	}

	d.remaining--
	return d.value, nil
}

type deltaDecoder struct {
	rle      *rleDecoder[int64]
	absolute int64
}

func newDeltaDecoder(data []byte) *deltaDecoder {
	return &deltaDecoder{rle: decodeRLEInt(data)}
}

func (d *deltaDecoder) next() (int64, bool, error) {
	delta, null, err := d.rle.next()
	if err != nil || null {
		return 0, null, err
	}
	d.absolute += delta
	return d.absolute, false, nil
}
