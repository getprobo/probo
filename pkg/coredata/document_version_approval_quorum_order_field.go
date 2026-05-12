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

import "fmt"

type (
	DocumentVersionApprovalQuorumOrderField string
)

const (
	DocumentVersionApprovalQuorumOrderFieldCreatedAt DocumentVersionApprovalQuorumOrderField = "CREATED_AT"
)

func (e DocumentVersionApprovalQuorumOrderField) Column() string {
	switch e {
	case DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return "created_at"
	}
	panic(fmt.Sprintf("unsupported order by: %s", e))
}

func (e DocumentVersionApprovalQuorumOrderField) IsValid() bool {
	switch e {
	case DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return true
	}
	return false
}

func (e DocumentVersionApprovalQuorumOrderField) String() string { return string(e) }

func (e *DocumentVersionApprovalQuorumOrderField) UnmarshalText(text []byte) error {
	*e = DocumentVersionApprovalQuorumOrderField(text)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid DocumentVersionApprovalQuorumOrderField", string(text))
	}
	return nil
}

func (e DocumentVersionApprovalQuorumOrderField) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}
