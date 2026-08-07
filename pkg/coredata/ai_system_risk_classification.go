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

type AiSystemRiskClassification string

const (
	AiSystemRiskClassificationHighRisk AiSystemRiskClassification = "HIGH_RISK"
	AiSystemRiskClassificationLimited  AiSystemRiskClassification = "LIMITED"
	AiSystemRiskClassificationMinimal  AiSystemRiskClassification = "MINIMAL"
	AiSystemRiskClassificationGPAI     AiSystemRiskClassification = "GPAI"
)

var (
	_ fmt.Stringer             = AiSystemRiskClassification("")
	_ encoding.TextMarshaler   = AiSystemRiskClassification("")
	_ encoding.TextUnmarshaler = (*AiSystemRiskClassification)(nil)
)

func AiSystemRiskClassifications() []AiSystemRiskClassification {
	return []AiSystemRiskClassification{
		AiSystemRiskClassificationHighRisk,
		AiSystemRiskClassificationLimited,
		AiSystemRiskClassificationMinimal,
		AiSystemRiskClassificationGPAI,
	}
}

func (v AiSystemRiskClassification) IsValid() bool {
	switch v {
	case
		AiSystemRiskClassificationHighRisk,
		AiSystemRiskClassificationLimited,
		AiSystemRiskClassificationMinimal,
		AiSystemRiskClassificationGPAI:
		return true
	}

	return false
}

func (v AiSystemRiskClassification) String() string {
	return string(v)
}

func (v AiSystemRiskClassification) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AiSystemRiskClassification) UnmarshalText(text []byte) error {
	val := AiSystemRiskClassification(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AiSystemRiskClassification value: %q", string(text))
	}

	*v = val

	return nil
}
