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

type MailingListSubscriberStatus string

const (
	MailingListSubscriberStatusPending   MailingListSubscriberStatus = "PENDING"
	MailingListSubscriberStatusConfirmed MailingListSubscriberStatus = "CONFIRMED"
)

var (
	_ fmt.Stringer             = MailingListSubscriberStatus("")
	_ encoding.TextMarshaler   = MailingListSubscriberStatus("")
	_ encoding.TextUnmarshaler = (*MailingListSubscriberStatus)(nil)
)

func MailingListSubscriberStatuses() []MailingListSubscriberStatus {
	return []MailingListSubscriberStatus{
		MailingListSubscriberStatusPending,
		MailingListSubscriberStatusConfirmed,
	}
}

func (v MailingListSubscriberStatus) IsValid() bool {
	switch v {
	case
		MailingListSubscriberStatusPending,
		MailingListSubscriberStatusConfirmed:
		return true
	}

	return false
}

func (v MailingListSubscriberStatus) String() string {
	return string(v)
}

func (v MailingListSubscriberStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MailingListSubscriberStatus) UnmarshalText(text []byte) error {
	val := MailingListSubscriberStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MailingListSubscriberStatus value: %q", string(text))
	}

	*v = val

	return nil
}
