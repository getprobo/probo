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
	ThirdPartyDataPrivacyAgreementOrderField string
)

const (
	ThirdPartyDataPrivacyAgreementOrderFieldValidFrom ThirdPartyDataPrivacyAgreementOrderField = "VALID_FROM"
	ThirdPartyDataPrivacyAgreementOrderFieldCreatedAt ThirdPartyDataPrivacyAgreementOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = ThirdPartyDataPrivacyAgreementOrderField("")
	_ fmt.Stringer             = ThirdPartyDataPrivacyAgreementOrderField("")
	_ encoding.TextMarshaler   = ThirdPartyDataPrivacyAgreementOrderField("")
	_ encoding.TextUnmarshaler = (*ThirdPartyDataPrivacyAgreementOrderField)(nil)
)

func ThirdPartyDataPrivacyAgreementOrderFields() []ThirdPartyDataPrivacyAgreementOrderField {
	return []ThirdPartyDataPrivacyAgreementOrderField{
		ThirdPartyDataPrivacyAgreementOrderFieldValidFrom,
		ThirdPartyDataPrivacyAgreementOrderFieldCreatedAt,
	}
}

func (v ThirdPartyDataPrivacyAgreementOrderField) IsValid() bool {
	switch v {
	case
		ThirdPartyDataPrivacyAgreementOrderFieldValidFrom,
		ThirdPartyDataPrivacyAgreementOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v ThirdPartyDataPrivacyAgreementOrderField) String() string {
	return string(v)
}

func (v ThirdPartyDataPrivacyAgreementOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ThirdPartyDataPrivacyAgreementOrderField) UnmarshalText(text []byte) error {
	val := ThirdPartyDataPrivacyAgreementOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ThirdPartyDataPrivacyAgreementOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ThirdPartyDataPrivacyAgreementOrderField) Column() string {
	return string(p)
}
