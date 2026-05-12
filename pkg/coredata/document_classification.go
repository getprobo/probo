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
	"database/sql/driver"
	"fmt"
)

type DocumentClassification string

const (
	DocumentClassificationPublic       DocumentClassification = "PUBLIC"
	DocumentClassificationInternal     DocumentClassification = "INTERNAL"
	DocumentClassificationConfidential DocumentClassification = "CONFIDENTIAL"
	DocumentClassificationSecret       DocumentClassification = "SECRET"
)

func DocumentClassifications() []DocumentClassification {
	return []DocumentClassification{
		DocumentClassificationPublic,
		DocumentClassificationInternal,
		DocumentClassificationConfidential,
		DocumentClassificationSecret,
	}
}

func (dc DocumentClassification) String() string {
	switch dc {
	case DocumentClassificationPublic:
		return "PUBLIC"
	case DocumentClassificationInternal:
		return "INTERNAL"
	case DocumentClassificationConfidential:
		return "CONFIDENTIAL"
	case DocumentClassificationSecret:
		return "SECRET"
	}
	panic(fmt.Errorf("invalid DocumentClassification value: %s", string(dc)))
}

// Scan implements the sql.Scanner interface for database deserialization.
func (dc *DocumentClassification) Scan(value any) error {
	if value == nil {
		return nil
	}

	var sv string
	switch v := value.(type) {
	case string:
		sv = v
	case []byte:
		sv = string(v)
	default:
		return fmt.Errorf("cannot scan DocumentClassification: expected string or []byte, got %T", value)
	}

	*dc = DocumentClassification(sv)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (dc DocumentClassification) Value() (driver.Value, error) {
	return string(dc), nil
}
