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
	"iter"
	"slices"
	"strings"
)

type (
	OAuth2Scope  string
	OAuth2Scopes []OAuth2Scope
)

var (
	_ fmt.Stringer             = OAuth2Scope("")
	_ encoding.TextMarshaler   = OAuth2Scope("")
	_ encoding.TextUnmarshaler = (*OAuth2Scope)(nil)
)

func (v OAuth2Scope) String() string {
	return string(v)
}

func (v OAuth2Scope) IsRead() bool {
	return strings.HasSuffix(string(v), ":read")
}

func (v OAuth2Scope) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2Scope) UnmarshalText(text []byte) error {
	*v = OAuth2Scope(text)

	return nil
}

func (s OAuth2Scopes) All() iter.Seq2[int, OAuth2Scope] {
	return slices.All(s)
}

func (s OAuth2Scopes) Values() iter.Seq[OAuth2Scope] {
	return slices.Values(s)
}

func (s OAuth2Scopes) Contains(scope OAuth2Scope) bool {
	return slices.Contains(s, scope)
}

func (s OAuth2Scopes) ContainsAll(seq iter.Seq[OAuth2Scope]) bool {
	for scope := range seq {
		if !s.Contains(scope) {
			return false
		}
	}

	return true
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
