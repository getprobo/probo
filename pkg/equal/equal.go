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

// Package equal contains small helpers for value equality across the
// codebase, with no business semantics.
package equal

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Ptr reports whether two nullable values are equal: both nil are equal,
// one nil is not equal to a non-nil, otherwise the pointed-to values are
// compared with ==.
func Ptr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// JSON reports whether two JSON blobs are semantically identical after
// normalising whitespace and (top-level and nested) object key ordering.
// Array element order is preserved as significant. Numbers are compared
// after JSON unmarshalling, so 1 and 1.0 compare equal.
func JSON(a, b json.RawMessage) (bool, error) {
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		return false, fmt.Errorf("cannot unmarshal first json blob: %w", err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false, fmt.Errorf("cannot unmarshal second json blob: %w", err)
	}
	return reflect.DeepEqual(av, bv), nil
}
