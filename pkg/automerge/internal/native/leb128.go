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

func appendULEB128(destination []byte, value uint64) []byte {
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		destination = append(destination, current)
		if value == 0 {
			return destination
		}
	}
}

func appendSLEB128(destination []byte, value int64) []byte {
	for {
		current := byte(value & 0x7f)
		value >>= 7
		done := (value == 0 && current&0x40 == 0) ||
			(value == -1 && current&0x40 != 0)
		if !done {
			current |= 0x80
		}
		destination = append(destination, current)
		if done {
			return destination
		}
	}
}

func readULEB128(data []byte) (uint64, int, error) {
	var (
		value uint64
		shift uint
	)

	for i, current := range data {
		if i == 10 || (i == 9 && current > 1) {
			return 0, 0, fmt.Errorf("unsigned LEB128 overflow")
		}

		value |= uint64(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, i + 1, nil
		}
		shift += 7
	}

	return 0, 0, fmt.Errorf("truncated unsigned LEB128")
}
