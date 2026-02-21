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
	"strings"
)

type WebhookEventType string

const (
	WebhookEventTypeMeetingCreated WebhookEventType = "meeting:created"
	WebhookEventTypeMeetingUpdated WebhookEventType = "meeting:updated"
	WebhookEventTypeMeetingDeleted WebhookEventType = "meeting:deleted"
	WebhookEventTypeVendorCreated  WebhookEventType = "vendor:created"
	WebhookEventTypeVendorUpdated  WebhookEventType = "vendor:updated"
	WebhookEventTypeVendorDeleted  WebhookEventType = "vendor:deleted"
)

func (w WebhookEventType) String() string {
	return string(w)
}

func (w WebhookEventType) IsValid() bool {
	switch w {
	case WebhookEventTypeMeetingCreated, WebhookEventTypeMeetingUpdated, WebhookEventTypeMeetingDeleted,
		WebhookEventTypeVendorCreated, WebhookEventTypeVendorUpdated, WebhookEventTypeVendorDeleted:
		return true
	}
	return false
}

func (w WebhookEventType) MarshalText() ([]byte, error) {
	return []byte(w.String()), nil
}

func (w *WebhookEventType) UnmarshalText(text []byte) error {
	*w = WebhookEventType(text)
	if !w.IsValid() {
		return fmt.Errorf("%s is not a valid WebhookEventType", string(text))
	}
	return nil
}

func (w *WebhookEventType) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for WebhookEventType: %T", value)
	}

	return w.UnmarshalText([]byte(s))
}

func (w WebhookEventType) Value() (driver.Value, error) {
	return w.String(), nil
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
		if err := et.Scan(part); err != nil {
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
