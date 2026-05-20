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
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type WebhookEventType string

const (
	WebhookEventTypeThirdPartyCreated WebhookEventType = "third-party:created"
	WebhookEventTypeThirdPartyUpdated WebhookEventType = "third-party:updated"
	WebhookEventTypeThirdPartyDeleted WebhookEventType = "third-party:deleted"
	WebhookEventTypeUserCreated       WebhookEventType = "user:created"
	WebhookEventTypeUserUpdated       WebhookEventType = "user:updated"
	WebhookEventTypeUserDeleted       WebhookEventType = "user:deleted"
	WebhookEventTypeObligationCreated WebhookEventType = "obligation:created"
	WebhookEventTypeObligationUpdated WebhookEventType = "obligation:updated"
	WebhookEventTypeObligationDeleted WebhookEventType = "obligation:deleted"
)

var (
	_ fmt.Stringer             = WebhookEventType("")
	_ encoding.TextMarshaler   = WebhookEventType("")
	_ encoding.TextUnmarshaler = (*WebhookEventType)(nil)
)

func (v WebhookEventType) IsValid() bool {
	switch v {
	case
		WebhookEventTypeThirdPartyCreated,
		WebhookEventTypeThirdPartyUpdated,
		WebhookEventTypeThirdPartyDeleted,
		WebhookEventTypeUserCreated,
		WebhookEventTypeUserUpdated,
		WebhookEventTypeUserDeleted,
		WebhookEventTypeObligationCreated,
		WebhookEventTypeObligationUpdated,
		WebhookEventTypeObligationDeleted:
		return true
	}

	return false
}

func (v WebhookEventType) String() string {
	return string(v)
}

func (v WebhookEventType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *WebhookEventType) UnmarshalText(text []byte) error {
	val := WebhookEventType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid WebhookEventType value: %q", string(text))
	}

	*v = val

	return nil
}

type WebhookEventTypes []WebhookEventType

func (s *WebhookEventTypes) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.scanFromString(v)
	case []byte:
		return s.scanFromString(string(v))
	default:
		return fmt.Errorf("unsupported type for WebhookEventTypes: %T", value)
	}
}

func (s *WebhookEventTypes) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "{}" || str == "" {
		*s = []WebhookEventType{}
		return nil
	}

	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	parts := strings.Split(str, ",")
	result := make([]WebhookEventType, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}

		var et WebhookEventType
		if err := et.UnmarshalText([]byte(part)); err != nil {
			return fmt.Errorf("invalid webhook event type in array: %s", part)
		}

		result[i] = et
	}

	*s = result

	return nil
}

func (s WebhookEventTypes) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{}", nil
	}

	values := make([]string, len(s))
	for i, et := range s {
		values[i] = et.String()
	}

	return "{" + strings.Join(values, ",") + "}", nil
}
