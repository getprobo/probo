// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type WebhookDataStatus string

const (
	WebhookDataStatusPending    WebhookDataStatus = "PENDING"
	WebhookDataStatusProcessing WebhookDataStatus = "PROCESSING"
	WebhookDataStatusDelivered  WebhookDataStatus = "DELIVERED"
)

func (s WebhookDataStatus) String() string {
	return string(s)
}

func (s WebhookDataStatus) IsValid() bool {
	switch s {
	case WebhookDataStatusPending, WebhookDataStatusProcessing, WebhookDataStatusDelivered:
		return true
	}
	return false
}

func (s WebhookDataStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *WebhookDataStatus) UnmarshalText(text []byte) error {
	*s = WebhookDataStatus(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid WebhookDataStatus", string(text))
	}
	return nil
}

func (s *WebhookDataStatus) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.UnmarshalText([]byte(v))
	case []byte:
		return s.UnmarshalText(v)
	default:
		return fmt.Errorf("unsupported type for WebhookDataStatus: %T", value)
	}
}

func (s WebhookDataStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
