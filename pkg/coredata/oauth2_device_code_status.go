// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
