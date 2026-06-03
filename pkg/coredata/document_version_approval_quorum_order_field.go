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
	DocumentVersionApprovalQuorumOrderField string
)

const (
	DocumentVersionApprovalQuorumOrderFieldCreatedAt DocumentVersionApprovalQuorumOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DocumentVersionApprovalQuorumOrderField("")
	_ fmt.Stringer             = DocumentVersionApprovalQuorumOrderField("")
	_ encoding.TextMarshaler   = DocumentVersionApprovalQuorumOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentVersionApprovalQuorumOrderField)(nil)
)

func DocumentVersionApprovalQuorumOrderFields() []DocumentVersionApprovalQuorumOrderField {
	return []DocumentVersionApprovalQuorumOrderField{
		DocumentVersionApprovalQuorumOrderFieldCreatedAt,
	}
}

func (v DocumentVersionApprovalQuorumOrderField) IsValid() bool {
	switch v {
	case
		DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v DocumentVersionApprovalQuorumOrderField) String() string {
	return string(v)
}

func (v DocumentVersionApprovalQuorumOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionApprovalQuorumOrderField) UnmarshalText(text []byte) error {
	val := DocumentVersionApprovalQuorumOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionApprovalQuorumOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (e DocumentVersionApprovalQuorumOrderField) Column() string {
	switch e {
	case DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", e))
}
