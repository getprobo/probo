// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type DocumentClassification string

const (
	DocumentClassificationPublic       DocumentClassification = "PUBLIC"
	DocumentClassificationInternal     DocumentClassification = "INTERNAL"
	DocumentClassificationConfidential DocumentClassification = "CONFIDENTIAL"
	DocumentClassificationSecret       DocumentClassification = "SECRET"
)

var (
	_ fmt.Stringer             = DocumentClassification("")
	_ encoding.TextMarshaler   = DocumentClassification("")
	_ encoding.TextUnmarshaler = (*DocumentClassification)(nil)
)

func DocumentClassifications() []DocumentClassification {
	return []DocumentClassification{
		DocumentClassificationPublic,
		DocumentClassificationInternal,
		DocumentClassificationConfidential,
		DocumentClassificationSecret,
	}
}

func (v DocumentClassification) IsValid() bool {
	switch v {
	case
		DocumentClassificationPublic,
		DocumentClassificationInternal,
		DocumentClassificationConfidential,
		DocumentClassificationSecret:
		return true
	}

	return false
}

func (v DocumentClassification) String() string {
	return string(v)
}

func (v DocumentClassification) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentClassification) UnmarshalText(text []byte) error {
	val := DocumentClassification(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentClassification value: %q", string(text))
	}

	*v = val

	return nil
}

// Scan implements the sql.Scanner interface for database deserialization.
// Value implements the driver.Valuer interface for database serialization.
