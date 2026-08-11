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

	"go.probo.inc/probo/pkg/page"
)

type BusinessFunctionOrderField string

const (
	BusinessFunctionOrderFieldCreatedAt      BusinessFunctionOrderField = "CREATED_AT"
	BusinessFunctionOrderFieldReferenceID    BusinessFunctionOrderField = "REFERENCE_ID"
	BusinessFunctionOrderFieldName           BusinessFunctionOrderField = "NAME"
	BusinessFunctionOrderFieldClassification BusinessFunctionOrderField = "CLASSIFICATION"
	BusinessFunctionOrderFieldMTDMinutes     BusinessFunctionOrderField = "MTD_MINUTES"
	BusinessFunctionOrderFieldRTOMinutes     BusinessFunctionOrderField = "RTO_MINUTES"
	BusinessFunctionOrderFieldRPOMinutes     BusinessFunctionOrderField = "RPO_MINUTES"
)

var (
	_ page.OrderField          = BusinessFunctionOrderField("")
	_ fmt.Stringer             = BusinessFunctionOrderField("")
	_ encoding.TextMarshaler   = BusinessFunctionOrderField("")
	_ encoding.TextUnmarshaler = (*BusinessFunctionOrderField)(nil)
)

func BusinessFunctionOrderFields() []BusinessFunctionOrderField {
	return []BusinessFunctionOrderField{
		BusinessFunctionOrderFieldCreatedAt,
		BusinessFunctionOrderFieldReferenceID,
		BusinessFunctionOrderFieldName,
		BusinessFunctionOrderFieldClassification,
		BusinessFunctionOrderFieldMTDMinutes,
		BusinessFunctionOrderFieldRTOMinutes,
		BusinessFunctionOrderFieldRPOMinutes,
	}
}

func (v BusinessFunctionOrderField) IsValid() bool {
	switch v {
	case
		BusinessFunctionOrderFieldCreatedAt,
		BusinessFunctionOrderFieldReferenceID,
		BusinessFunctionOrderFieldName,
		BusinessFunctionOrderFieldClassification,
		BusinessFunctionOrderFieldMTDMinutes,
		BusinessFunctionOrderFieldRTOMinutes,
		BusinessFunctionOrderFieldRPOMinutes:
		return true
	}

	return false
}

func (v BusinessFunctionOrderField) String() string {
	return string(v)
}

func (v BusinessFunctionOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *BusinessFunctionOrderField) UnmarshalText(text []byte) error {
	val := BusinessFunctionOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid BusinessFunctionOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p BusinessFunctionOrderField) Column() string {
	return string(p)
}
