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
