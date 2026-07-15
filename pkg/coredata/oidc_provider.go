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

type OIDCProvider string

const (
	OIDCProviderGoogle           OIDCProvider = "GOOGLE"
	OIDCProviderMicrosoft        OIDCProvider = "MICROSOFT"
	OIDCProviderCompliancePortal OIDCProvider = "COMPLIANCE_PORTAL"
)

var (
	_ fmt.Stringer             = OIDCProvider("")
	_ encoding.TextMarshaler   = OIDCProvider("")
	_ encoding.TextUnmarshaler = (*OIDCProvider)(nil)
)

func OIDCProviders() []OIDCProvider {
	return []OIDCProvider{
		OIDCProviderGoogle,
		OIDCProviderMicrosoft,
		OIDCProviderCompliancePortal,
	}
}

func (v OIDCProvider) IsValid() bool {
	switch v {
	case
		OIDCProviderGoogle,
		OIDCProviderMicrosoft,
		OIDCProviderCompliancePortal:
		return true
	}

	return false
}

func (v OIDCProvider) String() string {
	return string(v)
}

func (v OIDCProvider) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OIDCProvider) UnmarshalText(text []byte) error {
	val := OIDCProvider(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OIDCProvider value: %q", string(text))
	}

	*v = val

	return nil
}
