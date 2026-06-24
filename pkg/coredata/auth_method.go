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

type AccessReviewEntryAuthMethod string

const (
	AccessReviewEntryAuthMethodSSO            AccessReviewEntryAuthMethod = "SSO"
	AccessReviewEntryAuthMethodPassword       AccessReviewEntryAuthMethod = "PASSWORD"
	AccessReviewEntryAuthMethodAPIKey         AccessReviewEntryAuthMethod = "API_KEY"
	AccessReviewEntryAuthMethodServiceAccount AccessReviewEntryAuthMethod = "SERVICE_ACCOUNT"
	AccessReviewEntryAuthMethodUnknown        AccessReviewEntryAuthMethod = "UNKNOWN"
)

var (
	_ fmt.Stringer             = AccessReviewEntryAuthMethod("")
	_ encoding.TextMarshaler   = AccessReviewEntryAuthMethod("")
	_ encoding.TextUnmarshaler = (*AccessReviewEntryAuthMethod)(nil)
)

func AccessReviewEntryAuthMethods() []AccessReviewEntryAuthMethod {
	return []AccessReviewEntryAuthMethod{
		AccessReviewEntryAuthMethodSSO,
		AccessReviewEntryAuthMethodPassword,
		AccessReviewEntryAuthMethodAPIKey,
		AccessReviewEntryAuthMethodServiceAccount,
		AccessReviewEntryAuthMethodUnknown,
	}
}

func (v AccessReviewEntryAuthMethod) IsValid() bool {
	switch v {
	case
		AccessReviewEntryAuthMethodSSO,
		AccessReviewEntryAuthMethodPassword,
		AccessReviewEntryAuthMethodAPIKey,
		AccessReviewEntryAuthMethodServiceAccount,
		AccessReviewEntryAuthMethodUnknown:
		return true
	}

	return false
}

func (v AccessReviewEntryAuthMethod) String() string {
	return string(v)
}

func (v AccessReviewEntryAuthMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewEntryAuthMethod) UnmarshalText(text []byte) error {
	val := AccessReviewEntryAuthMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewEntryAuthMethod value: %q", string(text))
	}

	*v = val

	return nil
}
