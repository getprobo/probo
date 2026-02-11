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

type WebhookCallStatus string

const (
	WebhookCallStatusSucceeded WebhookCallStatus = "SUCCEEDED"
	WebhookCallStatusFailed    WebhookCallStatus = "FAILED"
)

func (s WebhookCallStatus) String() string {
	return string(s)
}

func (s WebhookCallStatus) IsValid() bool {
	switch s {
	case WebhookCallStatusSucceeded, WebhookCallStatusFailed:
		return true
	}
	return false
}

func (s WebhookCallStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *WebhookCallStatus) UnmarshalText(text []byte) error {
	*s = WebhookCallStatus(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid WebhookCallStatus", string(text))
	}
	return nil
}

func (s *WebhookCallStatus) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported type for WebhookCallStatus: %T", value)
	}

	return s.UnmarshalText([]byte(str))
}

func (s WebhookCallStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
