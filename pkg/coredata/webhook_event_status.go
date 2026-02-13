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

type WebhookEventStatus string

const (
	WebhookEventStatusPending WebhookEventStatus = "PENDING"
	WebhookEventStatusSucceeded  WebhookEventStatus = "SUCCEEDED"
	WebhookEventStatusFailed     WebhookEventStatus = "FAILED"
)

func (s WebhookEventStatus) String() string {
	return string(s)
}

func (s WebhookEventStatus) IsValid() bool {
	switch s {
	case WebhookEventStatusPending, WebhookEventStatusSucceeded, WebhookEventStatusFailed:
		return true
	}
	return false
}

func (s WebhookEventStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *WebhookEventStatus) UnmarshalText(text []byte) error {
	*s = WebhookEventStatus(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid WebhookEventStatus", string(text))
	}
	return nil
}

func (s *WebhookEventStatus) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.UnmarshalText([]byte(v))
	case []byte:
		return s.UnmarshalText(v)
	default:
		return fmt.Errorf("unsupported type for WebhookEventStatus: %T", value)
	}
}

func (s WebhookEventStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
