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
