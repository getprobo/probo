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

type (
	ThirdPartyContactOrderField string
)

const (
	ThirdPartyContactOrderFieldCreatedAt ThirdPartyContactOrderField = "CREATED_AT"
	ThirdPartyContactOrderFieldFullName  ThirdPartyContactOrderField = "FULL_NAME"
	ThirdPartyContactOrderFieldEmail     ThirdPartyContactOrderField = "EMAIL"
)

var (
	_ page.OrderField          = ThirdPartyContactOrderField("")
	_ fmt.Stringer             = ThirdPartyContactOrderField("")
	_ encoding.TextMarshaler   = ThirdPartyContactOrderField("")
	_ encoding.TextUnmarshaler = (*ThirdPartyContactOrderField)(nil)
)

func ThirdPartyContactOrderFields() []ThirdPartyContactOrderField {
	return []ThirdPartyContactOrderField{
		ThirdPartyContactOrderFieldCreatedAt,
		ThirdPartyContactOrderFieldFullName,
		ThirdPartyContactOrderFieldEmail,
	}
}

func (v ThirdPartyContactOrderField) IsValid() bool {
	switch v {
	case
		ThirdPartyContactOrderFieldCreatedAt,
		ThirdPartyContactOrderFieldFullName,
		ThirdPartyContactOrderFieldEmail:
		return true
	}

	return false
}

func (v ThirdPartyContactOrderField) String() string {
	return string(v)
}

func (v ThirdPartyContactOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ThirdPartyContactOrderField) UnmarshalText(text []byte) error {
	val := ThirdPartyContactOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ThirdPartyContactOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ThirdPartyContactOrderField) Column() string {
	return string(p)
}
