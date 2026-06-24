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

type TrackerType string

const (
	TrackerTypeCookie         TrackerType = "COOKIE"
	TrackerTypeLocalStorage   TrackerType = "LOCAL_STORAGE"
	TrackerTypeSessionStorage TrackerType = "SESSION_STORAGE"
	TrackerTypeIndexedDB      TrackerType = "INDEXED_DB"
	TrackerTypeCacheStorage   TrackerType = "CACHE_STORAGE"
)

var (
	_ fmt.Stringer             = TrackerType("")
	_ encoding.TextMarshaler   = TrackerType("")
	_ encoding.TextUnmarshaler = (*TrackerType)(nil)
)

func TrackerTypes() []TrackerType {
	return []TrackerType{
		TrackerTypeCookie,
		TrackerTypeLocalStorage,
		TrackerTypeSessionStorage,
		TrackerTypeIndexedDB,
		TrackerTypeCacheStorage,
	}
}

func (v TrackerType) IsValid() bool {
	switch v {
	case
		TrackerTypeCookie,
		TrackerTypeLocalStorage,
		TrackerTypeSessionStorage,
		TrackerTypeIndexedDB,
		TrackerTypeCacheStorage:
		return true
	}

	return false
}

func (v TrackerType) String() string {
	return string(v)
}

// Label returns a human-readable name for the tracker type, suitable for
// display in visitor-facing documents such as the cookie and tracking
// technologies policy.
func (v TrackerType) Label() string {
	switch v {
	case TrackerTypeCookie:
		return "Cookie"
	case TrackerTypeLocalStorage:
		return "Local storage"
	case TrackerTypeSessionStorage:
		return "Session storage"
	case TrackerTypeIndexedDB:
		return "IndexedDB"
	case TrackerTypeCacheStorage:
		return "Cache storage"
	default:
		return string(v)
	}
}

func (v TrackerType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrackerType) UnmarshalText(text []byte) error {
	val := TrackerType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrackerType value: %q", string(text))
	}

	*v = val

	return nil
}
