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
	"database/sql/driver"
	"fmt"
)

type SearchEngineIndexing string

const (
	SearchEngineIndexingIndexable    SearchEngineIndexing = "INDEXABLE"
	SearchEngineIndexingNotIndexable SearchEngineIndexing = "NOT_INDEXABLE"
)

func (s SearchEngineIndexing) String() string {
	return string(s)
}

func (s SearchEngineIndexing) IsValid() bool {
	switch s {
	case SearchEngineIndexingIndexable, SearchEngineIndexingNotIndexable:
		return true
	}
	return false
}

func (s *SearchEngineIndexing) UnmarshalText(text []byte) error {
	*s = SearchEngineIndexing(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid SearchEngineIndexing", string(text))
	}
	return nil
}

func (s SearchEngineIndexing) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *SearchEngineIndexing) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type for SearchEngineIndexing: %T", value)
	}

	switch str {
	case "INDEXABLE":
		*s = SearchEngineIndexingIndexable
	case "NOT_INDEXABLE":
		*s = SearchEngineIndexingNotIndexable
	default:
		return fmt.Errorf("invalid SearchEngineIndexing value: %q", str)
	}
	return nil
}

func (s SearchEngineIndexing) Value() (driver.Value, error) {
	return s.String(), nil
}
