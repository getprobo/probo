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

import "fmt"

type OAuth2SubjectType string

const (
	OAuth2SubjectTypePublic OAuth2SubjectType = "public"
)

func (s OAuth2SubjectType) IsValid() bool {
	switch s {
	case OAuth2SubjectTypePublic:
		return true
	}

	return false
}

func (s OAuth2SubjectType) String() string { return string(s) }

func (s *OAuth2SubjectType) UnmarshalText(text []byte) error {
	*s = OAuth2SubjectType(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2SubjectType", string(text))
	}

	return nil
}

func (s OAuth2SubjectType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
