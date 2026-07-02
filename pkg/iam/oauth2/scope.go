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

package oauth2

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	ScopeOpenID        coredata.OAuth2Scope = "openid"
	ScopeProfile       coredata.OAuth2Scope = "profile"
	ScopeEmail         coredata.OAuth2Scope = "email"
	ScopeOfflineAccess coredata.OAuth2Scope = "offline_access"
)

func IsStandardScope(scope coredata.OAuth2Scope) bool {
	switch scope {
	case ScopeOpenID, ScopeProfile, ScopeEmail, ScopeOfflineAccess:
		return true
	}

	return false
}

func IsValid(scope coredata.OAuth2Scope) bool {
	return IsStandardScope(scope)
}

func UnmarshalScope(text []byte) (coredata.OAuth2Scope, error) {
	scope := coredata.OAuth2Scope(text)
	if !IsValid(scope) {
		return "", fmt.Errorf("invalid oauth2 scope value: %q", string(text))
	}

	return scope, nil
}

func UnmarshalScopes(text []byte) (coredata.OAuth2Scopes, error) {
	str := string(text)
	if str == "" {
		return nil, nil
	}

	fields := strings.Fields(str)

	scopes := make(coredata.OAuth2Scopes, len(fields))
	for i, f := range fields {
		scope, err := UnmarshalScope([]byte(f))
		if err != nil {
			return nil, err
		}

		scopes[i] = scope
	}

	return scopes, nil
}
