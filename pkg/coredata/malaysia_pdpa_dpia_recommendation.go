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

type MalaysiaPDPADPIARecommendation string

const (
	MalaysiaPDPADPIARecommendationNotIndicated      MalaysiaPDPADPIARecommendation = "NOT_INDICATED"
	MalaysiaPDPADPIARecommendationDPOReviewRequired MalaysiaPDPADPIARecommendation = "DPO_REVIEW_REQUIRED"
	MalaysiaPDPADPIARecommendationRequired          MalaysiaPDPADPIARecommendation = "REQUIRED"
)

var (
	_ fmt.Stringer             = MalaysiaPDPADPIARecommendation("")
	_ encoding.TextMarshaler   = MalaysiaPDPADPIARecommendation("")
	_ encoding.TextUnmarshaler = (*MalaysiaPDPADPIARecommendation)(nil)
)

func MalaysiaPDPADPIARecommendations() []MalaysiaPDPADPIARecommendation {
	return []MalaysiaPDPADPIARecommendation{
		MalaysiaPDPADPIARecommendationNotIndicated,
		MalaysiaPDPADPIARecommendationDPOReviewRequired,
		MalaysiaPDPADPIARecommendationRequired,
	}
}

func (v MalaysiaPDPADPIARecommendation) IsValid() bool {
	switch v {
	case MalaysiaPDPADPIARecommendationNotIndicated,
		MalaysiaPDPADPIARecommendationDPOReviewRequired,
		MalaysiaPDPADPIARecommendationRequired:
		return true
	}

	return false
}

func (v MalaysiaPDPADPIARecommendation) String() string { return string(v) }

func (v MalaysiaPDPADPIARecommendation) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MalaysiaPDPADPIARecommendation) UnmarshalText(text []byte) error {
	value := MalaysiaPDPADPIARecommendation(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid MalaysiaPDPADPIARecommendation value: %q", string(text))
	}

	*v = value

	return nil
}
