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

type FindingPriority string

const (
	FindingPriorityLow    FindingPriority = "LOW"
	FindingPriorityMedium FindingPriority = "MEDIUM"
	FindingPriorityHigh   FindingPriority = "HIGH"
)

var (
	_ fmt.Stringer             = FindingPriority("")
	_ encoding.TextMarshaler   = FindingPriority("")
	_ encoding.TextUnmarshaler = (*FindingPriority)(nil)
)

func FindingPriorities() []FindingPriority {
	return []FindingPriority{
		FindingPriorityLow,
		FindingPriorityMedium,
		FindingPriorityHigh,
	}
}

func (v FindingPriority) IsValid() bool {
	switch v {
	case
		FindingPriorityLow,
		FindingPriorityMedium,
		FindingPriorityHigh:
		return true
	}

	return false
}

func (v FindingPriority) String() string {
	return string(v)
}

func (v FindingPriority) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *FindingPriority) UnmarshalText(text []byte) error {
	val := FindingPriority(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid FindingPriority value: %q", string(text))
	}

	*v = val

	return nil
}
