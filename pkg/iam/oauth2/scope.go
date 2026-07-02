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
