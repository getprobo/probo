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

type (
	DocumentVersionStatus string
)

const (
	DocumentVersionStatusDraft           DocumentVersionStatus = "DRAFT"
	DocumentVersionStatusPendingApproval DocumentVersionStatus = "PENDING_APPROVAL"
	DocumentVersionStatusPublished       DocumentVersionStatus = "PUBLISHED"
)

var (
	_ fmt.Stringer             = DocumentVersionStatus("")
	_ encoding.TextMarshaler   = DocumentVersionStatus("")
	_ encoding.TextUnmarshaler = (*DocumentVersionStatus)(nil)
)

func DocumentVersionStatuses() []DocumentVersionStatus {
	return []DocumentVersionStatus{
		DocumentVersionStatusDraft,
		DocumentVersionStatusPendingApproval,
		DocumentVersionStatusPublished,
	}
}

func (v DocumentVersionStatus) IsValid() bool {
	switch v {
	case
		DocumentVersionStatusDraft,
		DocumentVersionStatusPendingApproval,
		DocumentVersionStatusPublished:
		return true
	}

	return false
}

func (v DocumentVersionStatus) String() string {
	return string(v)
}

func (v DocumentVersionStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionStatus) UnmarshalText(text []byte) error {
	val := DocumentVersionStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionStatus value: %q", string(text))
	}

	*v = val

	return nil
}
