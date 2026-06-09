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

type OAuth2DeviceCodeStatus string

const (
	OAuth2DeviceCodeStatusPending    OAuth2DeviceCodeStatus = "pending"
	OAuth2DeviceCodeStatusAuthorized OAuth2DeviceCodeStatus = "authorized"
	OAuth2DeviceCodeStatusDenied     OAuth2DeviceCodeStatus = "denied"
	OAuth2DeviceCodeStatusExpired    OAuth2DeviceCodeStatus = "expired"
)

var (
	_ fmt.Stringer             = OAuth2DeviceCodeStatus("")
	_ encoding.TextMarshaler   = OAuth2DeviceCodeStatus("")
	_ encoding.TextUnmarshaler = (*OAuth2DeviceCodeStatus)(nil)
)

func OAuth2DeviceCodeStatuses() []OAuth2DeviceCodeStatus {
	return []OAuth2DeviceCodeStatus{
		OAuth2DeviceCodeStatusPending,
		OAuth2DeviceCodeStatusAuthorized,
		OAuth2DeviceCodeStatusDenied,
		OAuth2DeviceCodeStatusExpired,
	}
}

func (v OAuth2DeviceCodeStatus) IsValid() bool {
	switch v {
	case
		OAuth2DeviceCodeStatusPending,
		OAuth2DeviceCodeStatusAuthorized,
		OAuth2DeviceCodeStatusDenied,
		OAuth2DeviceCodeStatusExpired:
		return true
	}

	return false
}

func (v OAuth2DeviceCodeStatus) String() string {
	return string(v)
}

func (v OAuth2DeviceCodeStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2DeviceCodeStatus) UnmarshalText(text []byte) error {
	val := OAuth2DeviceCodeStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2DeviceCodeStatus value: %q", string(text))
	}

	*v = val

	return nil
}
