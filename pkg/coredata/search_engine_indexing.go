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

type SearchEngineIndexing string

const (
	SearchEngineIndexingIndexable    SearchEngineIndexing = "INDEXABLE"
	SearchEngineIndexingNotIndexable SearchEngineIndexing = "NOT_INDEXABLE"
)

var (
	_ fmt.Stringer             = SearchEngineIndexing("")
	_ encoding.TextMarshaler   = SearchEngineIndexing("")
	_ encoding.TextUnmarshaler = (*SearchEngineIndexing)(nil)
)

func SearchEngineIndexings() []SearchEngineIndexing {
	return []SearchEngineIndexing{
		SearchEngineIndexingIndexable,
		SearchEngineIndexingNotIndexable,
	}
}

func (v SearchEngineIndexing) IsValid() bool {
	switch v {
	case
		SearchEngineIndexingIndexable,
		SearchEngineIndexingNotIndexable:
		return true
	}

	return false
}

func (v SearchEngineIndexing) String() string {
	return string(v)
}

func (v SearchEngineIndexing) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SearchEngineIndexing) UnmarshalText(text []byte) error {
	val := SearchEngineIndexing(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SearchEngineIndexing value: %q", string(text))
	}

	*v = val

	return nil
}
