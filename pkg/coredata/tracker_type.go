// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
