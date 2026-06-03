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

type (
	EmailStatus string
)

const (
	EmailStatusPending    EmailStatus = "PENDING"
	EmailStatusProcessing EmailStatus = "PROCESSING"
	EmailStatusSent       EmailStatus = "SENT"
	EmailStatusFailed     EmailStatus = "FAILED"
)

var (
	_ fmt.Stringer             = EmailStatus("")
	_ encoding.TextMarshaler   = EmailStatus("")
	_ encoding.TextUnmarshaler = (*EmailStatus)(nil)
)

func EmailStatuses() []EmailStatus {
	return []EmailStatus{
		EmailStatusPending,
		EmailStatusProcessing,
		EmailStatusSent,
		EmailStatusFailed,
	}
}

func (v EmailStatus) IsValid() bool {
	switch v {
	case
		EmailStatusPending,
		EmailStatusProcessing,
		EmailStatusSent,
		EmailStatusFailed:
		return true
	}

	return false
}

func (v EmailStatus) String() string {
	return string(v)
}

func (v EmailStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *EmailStatus) UnmarshalText(text []byte) error {
	val := EmailStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid EmailStatus value: %q", string(text))
	}

	*v = val

	return nil
}
