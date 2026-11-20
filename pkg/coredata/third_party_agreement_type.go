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

type ThirdPartyAgreementType string

const (
	ThirdPartyAgreementTypeBusinessAssociate ThirdPartyAgreementType = "BUSINESS_ASSOCIATE"
	ThirdPartyAgreementTypeDataPrivacy       ThirdPartyAgreementType = "DATA_PRIVACY"
	ThirdPartyAgreementTypeTerms             ThirdPartyAgreementType = "TERMS"
	ThirdPartyAgreementTypeMasterService     ThirdPartyAgreementType = "MASTER_SERVICE"
	ThirdPartyAgreementTypeAmendment         ThirdPartyAgreementType = "AMENDMENT"
	ThirdPartyAgreementTypeAnnex             ThirdPartyAgreementType = "ANNEX"
	ThirdPartyAgreementTypeOther             ThirdPartyAgreementType = "OTHER"
)

var (
	_ fmt.Stringer             = ThirdPartyAgreementType("")
	_ encoding.TextMarshaler   = ThirdPartyAgreementType("")
	_ encoding.TextUnmarshaler = (*ThirdPartyAgreementType)(nil)
)

func ThirdPartyAgreementTypes() []ThirdPartyAgreementType {
	return []ThirdPartyAgreementType{
		ThirdPartyAgreementTypeBusinessAssociate,
		ThirdPartyAgreementTypeDataPrivacy,
		ThirdPartyAgreementTypeTerms,
		ThirdPartyAgreementTypeMasterService,
		ThirdPartyAgreementTypeAmendment,
		ThirdPartyAgreementTypeAnnex,
		ThirdPartyAgreementTypeOther,
	}
}

func (t ThirdPartyAgreementType) IsValid() bool {
	switch t {
	case ThirdPartyAgreementTypeBusinessAssociate,
		ThirdPartyAgreementTypeDataPrivacy,
		ThirdPartyAgreementTypeTerms,
		ThirdPartyAgreementTypeMasterService,
		ThirdPartyAgreementTypeAmendment,
		ThirdPartyAgreementTypeAnnex,
		ThirdPartyAgreementTypeOther:
		return true
	}

	return false
}

func (t ThirdPartyAgreementType) String() string {
	return string(t)
}

func (t ThirdPartyAgreementType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *ThirdPartyAgreementType) UnmarshalText(text []byte) error {
	value := ThirdPartyAgreementType(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid ThirdPartyAgreementType value: %q", string(text))
	}

	*t = value

	return nil
}
