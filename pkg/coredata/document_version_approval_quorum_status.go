// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
	"fmt"
)

type DocumentVersionApprovalQuorumStatus string

const (
	DocumentVersionApprovalQuorumStatusPending  DocumentVersionApprovalQuorumStatus = "PENDING"
	DocumentVersionApprovalQuorumStatusApproved DocumentVersionApprovalQuorumStatus = "APPROVED"
	DocumentVersionApprovalQuorumStatusRejected DocumentVersionApprovalQuorumStatus = "REJECTED"
)

func (s DocumentVersionApprovalQuorumStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *DocumentVersionApprovalQuorumStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentVersionApprovalQuorumStatusPending.String():
		*s = DocumentVersionApprovalQuorumStatusPending
	case DocumentVersionApprovalQuorumStatusApproved.String():
		*s = DocumentVersionApprovalQuorumStatusApproved
	case DocumentVersionApprovalQuorumStatusRejected.String():
		*s = DocumentVersionApprovalQuorumStatusRejected
	default:
		return fmt.Errorf("invalid DocumentVersionApprovalQuorumStatus value: %q", val)
	}

	return nil
}

func (s DocumentVersionApprovalQuorumStatus) String() string {
	var val string

	switch s {
	case DocumentVersionApprovalQuorumStatusPending:
		val = "PENDING"
	case DocumentVersionApprovalQuorumStatusApproved:
		val = "APPROVED"
	case DocumentVersionApprovalQuorumStatusRejected:
		val = "REJECTED"
	default:
		panic(fmt.Errorf("invalid DocumentVersionApprovalQuorumStatus value: %q", string(s)))
	}

	return val
}

func (s *DocumentVersionApprovalQuorumStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentVersionApprovalQuorumStatus, expected string got %T", value)
	}

	return s.UnmarshalText([]byte(val))
}

func (s DocumentVersionApprovalQuorumStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
