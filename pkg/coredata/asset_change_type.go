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
	"encoding"
	"fmt"
)

type AssetChangeType string

const (
	AssetChangeTypeCreated     AssetChangeType = "CREATED"
	AssetChangeTypeUpdated     AssetChangeType = "UPDATED"
	AssetChangeTypeDeactivated AssetChangeType = "DEACTIVATED"
)

var (
	_ fmt.Stringer             = AssetChangeType("")
	_ encoding.TextMarshaler   = AssetChangeType("")
	_ encoding.TextUnmarshaler = (*AssetChangeType)(nil)
)

func AssetChangeTypes() []AssetChangeType {
	return []AssetChangeType{
		AssetChangeTypeCreated,
		AssetChangeTypeUpdated,
		AssetChangeTypeDeactivated,
	}
}

func (v AssetChangeType) IsValid() bool {
	switch v {
	case AssetChangeTypeCreated, AssetChangeTypeUpdated, AssetChangeTypeDeactivated:
		return true
	}

	return false
}

func (v AssetChangeType) String() string {
	return string(v)
}

func (v AssetChangeType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AssetChangeType) UnmarshalText(text []byte) error {
	val := AssetChangeType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AssetChangeType value: %q", string(text))
	}

	*v = val

	return nil
}
