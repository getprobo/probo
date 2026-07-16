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

type (
	CompliancePortalCommitmentGroupOrderField string
)

const (
	CompliancePortalCommitmentGroupOrderFieldRank      CompliancePortalCommitmentGroupOrderField = "RANK"
	CompliancePortalCommitmentGroupOrderFieldCreatedAt CompliancePortalCommitmentGroupOrderField = "CREATED_AT"
	CompliancePortalCommitmentGroupOrderFieldUpdatedAt CompliancePortalCommitmentGroupOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = CompliancePortalCommitmentGroupOrderField("")
	_ fmt.Stringer             = CompliancePortalCommitmentGroupOrderField("")
	_ encoding.TextMarshaler   = CompliancePortalCommitmentGroupOrderField("")
	_ encoding.TextUnmarshaler = (*CompliancePortalCommitmentGroupOrderField)(nil)
)

func CompliancePortalCommitmentGroupOrderFields() []CompliancePortalCommitmentGroupOrderField {
	return []CompliancePortalCommitmentGroupOrderField{
		CompliancePortalCommitmentGroupOrderFieldRank,
		CompliancePortalCommitmentGroupOrderFieldCreatedAt,
		CompliancePortalCommitmentGroupOrderFieldUpdatedAt,
	}
}

func (v CompliancePortalCommitmentGroupOrderField) IsValid() bool {
	switch v {
	case
		CompliancePortalCommitmentGroupOrderFieldRank,
		CompliancePortalCommitmentGroupOrderFieldCreatedAt,
		CompliancePortalCommitmentGroupOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v CompliancePortalCommitmentGroupOrderField) String() string {
	return string(v)
}

func (v CompliancePortalCommitmentGroupOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CompliancePortalCommitmentGroupOrderField) UnmarshalText(text []byte) error {
	val := CompliancePortalCommitmentGroupOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CompliancePortalCommitmentGroupOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p CompliancePortalCommitmentGroupOrderField) Column() string {
	switch p {
	case CompliancePortalCommitmentGroupOrderFieldRank:
		return "rank"
	case CompliancePortalCommitmentGroupOrderFieldCreatedAt:
		return "created_at"
	case CompliancePortalCommitmentGroupOrderFieldUpdatedAt:
		return "updated_at"
	default:
		return string(p)
	}
}
