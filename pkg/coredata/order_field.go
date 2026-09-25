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

package coredata

import (
	"fmt"
	"slices"
	"strings"
)

func isValidOrderField[T ~string](v T, values []T) bool {
	return slices.Contains(values, v)
}

func unmarshalOrderField[T ~string](dst *T, text []byte, values []T) error {
	val := T(text)
	if !isValidOrderField(val, values) {
		return fmt.Errorf("invalid %s value: %q", orderFieldTypeName[T](), string(text))
	}

	*dst = val

	return nil
}

func orderFieldTypeName[T any]() string {
	var zero T

	name := fmt.Sprintf("%T", zero)
	_, after, ok := strings.Cut(name, ".")

	if ok {
		return after
	}

	return name
}
