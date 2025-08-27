// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"
)

type ContinualImprovementRegistriesOrderField string

const (
	ContinualImprovementRegistriesOrderFieldCreatedAt   ContinualImprovementRegistriesOrderField = "CREATED_AT"
	ContinualImprovementRegistriesOrderFieldTargetDate  ContinualImprovementRegistriesOrderField = "TARGET_DATE"
	ContinualImprovementRegistriesOrderFieldStatus      ContinualImprovementRegistriesOrderField = "STATUS"
	ContinualImprovementRegistriesOrderFieldPriority    ContinualImprovementRegistriesOrderField = "PRIORITY"
	ContinualImprovementRegistriesOrderFieldReferenceId ContinualImprovementRegistriesOrderField = "REFERENCE_ID"
)

func (p ContinualImprovementRegistriesOrderField) Column() string {
	return string(p)
}

func (p ContinualImprovementRegistriesOrderField) String() string {
	return string(p)
}

func (p ContinualImprovementRegistriesOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *ContinualImprovementRegistriesOrderField) UnmarshalText(text []byte) error {
	val := string(text)
	switch val {
	case string(ContinualImprovementRegistriesOrderFieldCreatedAt),
		string(ContinualImprovementRegistriesOrderFieldTargetDate),
		string(ContinualImprovementRegistriesOrderFieldStatus),
		string(ContinualImprovementRegistriesOrderFieldPriority),
		string(ContinualImprovementRegistriesOrderFieldReferenceId):
		*p = ContinualImprovementRegistriesOrderField(val)
		return nil
	}
	return fmt.Errorf("invalid ContinualImprovementRegistriesOrderField value: %q", val)
}
