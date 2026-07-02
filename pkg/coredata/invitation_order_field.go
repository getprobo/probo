// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// InvitationOrderField defines the fields that can be used to order invitations
type InvitationOrderField string

// InvitationOrderField constants
const (
	InvitationOrderFieldCreatedAt InvitationOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = InvitationOrderField("")
	_ fmt.Stringer             = InvitationOrderField("")
	_ encoding.TextMarshaler   = InvitationOrderField("")
	_ encoding.TextUnmarshaler = (*InvitationOrderField)(nil)
)

func InvitationOrderFields() []InvitationOrderField {
	return []InvitationOrderField{
		InvitationOrderFieldCreatedAt,
	}
}

func (v InvitationOrderField) IsValid() bool {
	switch v {
	case
		InvitationOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v InvitationOrderField) String() string {
	return string(v)
}

func (v InvitationOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InvitationOrderField) UnmarshalText(text []byte) error {
	val := InvitationOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InvitationOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p InvitationOrderField) Column() string {
	switch p {
	case InvitationOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
