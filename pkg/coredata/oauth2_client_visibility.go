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

type OAuth2ClientVisibility string

const (
	OAuth2ClientVisibilityPrivate OAuth2ClientVisibility = "private"
	OAuth2ClientVisibilityPublic  OAuth2ClientVisibility = "public"
)

var (
	_ fmt.Stringer             = OAuth2ClientVisibility("")
	_ encoding.TextMarshaler   = OAuth2ClientVisibility("")
	_ encoding.TextUnmarshaler = (*OAuth2ClientVisibility)(nil)
)

func OAuth2ClientVisibilities() []OAuth2ClientVisibility {
	return []OAuth2ClientVisibility{
		OAuth2ClientVisibilityPrivate,
		OAuth2ClientVisibilityPublic,
	}
}

func (v OAuth2ClientVisibility) IsValid() bool {
	switch v {
	case
		OAuth2ClientVisibilityPrivate,
		OAuth2ClientVisibilityPublic:
		return true
	}

	return false
}

func (v OAuth2ClientVisibility) String() string {
	return string(v)
}

func (v OAuth2ClientVisibility) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ClientVisibility) UnmarshalText(text []byte) error {
	val := OAuth2ClientVisibility(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ClientVisibility value: %q", string(text))
	}

	*v = val

	return nil
}
