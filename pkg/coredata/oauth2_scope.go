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
	"fmt"
	"strings"
)

type (
	OAuth2Scope  string
	OAuth2Scopes []OAuth2Scope
)

const (
	OAuth2ScopeOpenID  OAuth2Scope = "openid"
	OAuth2ScopeProfile OAuth2Scope = "profile"
	OAuth2ScopeEmail   OAuth2Scope = "email"
)

func (s OAuth2Scope) IsValid() bool {
	switch s {
	case OAuth2ScopeOpenID, OAuth2ScopeProfile, OAuth2ScopeEmail:
		return true
	}

	return false
}

func (s OAuth2Scope) String() string { return string(s) }

func (s *OAuth2Scope) UnmarshalText(text []byte) error {
	*s = OAuth2Scope(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2Scope", string(text))
	}

	return nil
}

func (s OAuth2Scope) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s OAuth2Scopes) String() string {
	ss := make([]string, len(s))
	for i, scope := range s {
		ss[i] = scope.String()
	}

	return strings.Join(ss, " ")
}

func (s OAuth2Scopes) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s OAuth2Scopes) OrDefault(defaultScopes OAuth2Scopes) OAuth2Scopes {
	if len(s) == 0 {
		return defaultScopes
	}
	return s
}

func (s *OAuth2Scopes) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" {
		*s = nil
		return nil
	}

	fields := strings.Fields(str)
	scopes := make(OAuth2Scopes, len(fields))
	for i, f := range fields {
		if err := scopes[i].UnmarshalText([]byte(f)); err != nil {
			return err
		}
	}

	*s = scopes
	return nil
}
