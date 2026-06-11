// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type AccessReviewEntryIncrementalTag string

const (
	AccessReviewEntryIncrementalTagNew       AccessReviewEntryIncrementalTag = "NEW"
	AccessReviewEntryIncrementalTagRemoved   AccessReviewEntryIncrementalTag = "REMOVED"
	AccessReviewEntryIncrementalTagUnchanged AccessReviewEntryIncrementalTag = "UNCHANGED"
)

var (
	_ fmt.Stringer             = AccessReviewEntryIncrementalTag("")
	_ encoding.TextMarshaler   = AccessReviewEntryIncrementalTag("")
	_ encoding.TextUnmarshaler = (*AccessReviewEntryIncrementalTag)(nil)
)

func AccessReviewEntryIncrementalTags() []AccessReviewEntryIncrementalTag {
	return []AccessReviewEntryIncrementalTag{
		AccessReviewEntryIncrementalTagNew,
		AccessReviewEntryIncrementalTagRemoved,
		AccessReviewEntryIncrementalTagUnchanged,
	}
}

func (v AccessReviewEntryIncrementalTag) IsValid() bool {
	switch v {
	case
		AccessReviewEntryIncrementalTagNew,
		AccessReviewEntryIncrementalTagRemoved,
		AccessReviewEntryIncrementalTagUnchanged:
		return true
	}

	return false
}

func (v AccessReviewEntryIncrementalTag) String() string {
	return string(v)
}

func (v AccessReviewEntryIncrementalTag) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewEntryIncrementalTag) UnmarshalText(text []byte) error {
	val := AccessReviewEntryIncrementalTag(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewEntryIncrementalTag value: %q", string(text))
	}

	*v = val

	return nil
}
