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

type MalaysiaPDPATransferBasis string

const (
	MalaysiaPDPATransferBasisSubstantiallySimilarLaw      MalaysiaPDPATransferBasis = "SUBSTANTIALLY_SIMILAR_LAW"
	MalaysiaPDPATransferBasisAdequateEquivalentProtection MalaysiaPDPATransferBasis = "ADEQUATE_EQUIVALENT_PROTECTION"
	MalaysiaPDPATransferBasisDataSubjectConsent           MalaysiaPDPATransferBasis = "DATA_SUBJECT_CONSENT"
	MalaysiaPDPATransferBasisDataSubjectContract          MalaysiaPDPATransferBasis = "DATA_SUBJECT_CONTRACT"
	MalaysiaPDPATransferBasisThirdPartyContract           MalaysiaPDPATransferBasis = "THIRD_PARTY_CONTRACT"
	MalaysiaPDPATransferBasisLegalProceedings             MalaysiaPDPATransferBasis = "LEGAL_PROCEEDINGS"
	MalaysiaPDPATransferBasisAdverseAction                MalaysiaPDPATransferBasis = "ADVERSE_ACTION"
	MalaysiaPDPATransferBasisReasonablePrecautions        MalaysiaPDPATransferBasis = "REASONABLE_PRECAUTIONS"
	MalaysiaPDPATransferBasisVitalInterests               MalaysiaPDPATransferBasis = "VITAL_INTERESTS"
)

var (
	_ fmt.Stringer             = MalaysiaPDPATransferBasis("")
	_ encoding.TextMarshaler   = MalaysiaPDPATransferBasis("")
	_ encoding.TextUnmarshaler = (*MalaysiaPDPATransferBasis)(nil)
)

func MalaysiaPDPATransferBases() []MalaysiaPDPATransferBasis {
	return []MalaysiaPDPATransferBasis{
		MalaysiaPDPATransferBasisSubstantiallySimilarLaw,
		MalaysiaPDPATransferBasisAdequateEquivalentProtection,
		MalaysiaPDPATransferBasisDataSubjectConsent,
		MalaysiaPDPATransferBasisDataSubjectContract,
		MalaysiaPDPATransferBasisThirdPartyContract,
		MalaysiaPDPATransferBasisLegalProceedings,
		MalaysiaPDPATransferBasisAdverseAction,
		MalaysiaPDPATransferBasisReasonablePrecautions,
		MalaysiaPDPATransferBasisVitalInterests,
	}
}

func (v MalaysiaPDPATransferBasis) RequiresTIA() bool {
	return v == MalaysiaPDPATransferBasisSubstantiallySimilarLaw ||
		v == MalaysiaPDPATransferBasisAdequateEquivalentProtection
}

func (v MalaysiaPDPATransferBasis) IsValid() bool {
	for _, value := range MalaysiaPDPATransferBases() {
		if v == value {
			return true
		}
	}

	return false
}

func (v MalaysiaPDPATransferBasis) String() string { return string(v) }

func (v MalaysiaPDPATransferBasis) MarshalText() ([]byte, error) { return []byte(v.String()), nil }

func (v *MalaysiaPDPATransferBasis) UnmarshalText(text []byte) error {
	value := MalaysiaPDPATransferBasis(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid MalaysiaPDPATransferBasis value: %q", string(text))
	}

	*v = value

	return nil
}
