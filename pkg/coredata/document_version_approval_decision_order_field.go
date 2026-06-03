// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersionApprovalDecisionOrderField string
)

const (
	DocumentVersionApprovalDecisionOrderFieldCreatedAt DocumentVersionApprovalDecisionOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DocumentVersionApprovalDecisionOrderField("")
	_ fmt.Stringer             = DocumentVersionApprovalDecisionOrderField("")
	_ encoding.TextMarshaler   = DocumentVersionApprovalDecisionOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentVersionApprovalDecisionOrderField)(nil)
)

func DocumentVersionApprovalDecisionOrderFields() []DocumentVersionApprovalDecisionOrderField {
	return []DocumentVersionApprovalDecisionOrderField{
		DocumentVersionApprovalDecisionOrderFieldCreatedAt,
	}
}

func (v DocumentVersionApprovalDecisionOrderField) IsValid() bool {
	switch v {
	case
		DocumentVersionApprovalDecisionOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v DocumentVersionApprovalDecisionOrderField) String() string {
	return string(v)
}

func (v DocumentVersionApprovalDecisionOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionApprovalDecisionOrderField) UnmarshalText(text []byte) error {
	val := DocumentVersionApprovalDecisionOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionApprovalDecisionOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (e DocumentVersionApprovalDecisionOrderField) Column() string {
	switch e {
	case DocumentVersionApprovalDecisionOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", e))
}
