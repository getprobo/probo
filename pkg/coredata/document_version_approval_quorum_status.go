// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
)

type DocumentVersionApprovalQuorumStatus string

const (
	DocumentVersionApprovalQuorumStatusPending  DocumentVersionApprovalQuorumStatus = "PENDING"
	DocumentVersionApprovalQuorumStatusApproved DocumentVersionApprovalQuorumStatus = "APPROVED"
	DocumentVersionApprovalQuorumStatusRejected DocumentVersionApprovalQuorumStatus = "REJECTED"
	DocumentVersionApprovalQuorumStatusVoided   DocumentVersionApprovalQuorumStatus = "VOIDED"
)

var (
	_ fmt.Stringer             = DocumentVersionApprovalQuorumStatus("")
	_ encoding.TextMarshaler   = DocumentVersionApprovalQuorumStatus("")
	_ encoding.TextUnmarshaler = (*DocumentVersionApprovalQuorumStatus)(nil)
)

func DocumentVersionApprovalQuorumStatuses() []DocumentVersionApprovalQuorumStatus {
	return []DocumentVersionApprovalQuorumStatus{
		DocumentVersionApprovalQuorumStatusPending,
		DocumentVersionApprovalQuorumStatusApproved,
		DocumentVersionApprovalQuorumStatusRejected,
		DocumentVersionApprovalQuorumStatusVoided,
	}
}

func (v DocumentVersionApprovalQuorumStatus) IsValid() bool {
	switch v {
	case
		DocumentVersionApprovalQuorumStatusPending,
		DocumentVersionApprovalQuorumStatusApproved,
		DocumentVersionApprovalQuorumStatusRejected,
		DocumentVersionApprovalQuorumStatusVoided:
		return true
	}

	return false
}

func (v DocumentVersionApprovalQuorumStatus) String() string {
	return string(v)
}

func (v DocumentVersionApprovalQuorumStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionApprovalQuorumStatus) UnmarshalText(text []byte) error {
	val := DocumentVersionApprovalQuorumStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionApprovalQuorumStatus value: %q", string(text))
	}

	*v = val

	return nil
}
