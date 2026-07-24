// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type DeviceOrderField string

const (
	DeviceOrderFieldCreatedAt  DeviceOrderField = "CREATED_AT"
	DeviceOrderFieldUpdatedAt  DeviceOrderField = "UPDATED_AT"
	DeviceOrderFieldHostname   DeviceOrderField = "HOSTNAME"
	DeviceOrderFieldLastSeenAt DeviceOrderField = "LAST_SEEN_AT"
)

var (
	_ page.OrderField          = DeviceOrderField("")
	_ fmt.Stringer             = DeviceOrderField("")
	_ encoding.TextMarshaler   = DeviceOrderField("")
	_ encoding.TextUnmarshaler = (*DeviceOrderField)(nil)
)

func DeviceOrderFields() []DeviceOrderField {
	return []DeviceOrderField{
		DeviceOrderFieldCreatedAt,
		DeviceOrderFieldUpdatedAt,
		DeviceOrderFieldHostname,
		DeviceOrderFieldLastSeenAt,
	}
}

func (v DeviceOrderField) IsValid() bool {
	switch v {
	case
		DeviceOrderFieldCreatedAt,
		DeviceOrderFieldUpdatedAt,
		DeviceOrderFieldHostname,
		DeviceOrderFieldLastSeenAt:
		return true
	}

	return false
}

func (v DeviceOrderField) String() string {
	return string(v)
}

func (v DeviceOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DeviceOrderField) UnmarshalText(text []byte) error {
	val := DeviceOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DeviceOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f DeviceOrderField) Column() string {
	switch f {
	case DeviceOrderFieldCreatedAt:
		return "created_at"
	case DeviceOrderFieldUpdatedAt:
		return "updated_at"
	case DeviceOrderFieldHostname:
		return "COALESCE(hostname, '')"
	case DeviceOrderFieldLastSeenAt:
		return "COALESCE(last_seen_at, '0001-01-01T00:00:00Z'::timestamptz)"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
