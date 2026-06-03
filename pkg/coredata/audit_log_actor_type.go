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
)

type AuditLogActorType string

const (
	AuditLogActorTypeUser   AuditLogActorType = "USER"
	AuditLogActorTypeAPIKey AuditLogActorType = "API_KEY"
	AuditLogActorTypeSystem AuditLogActorType = "SYSTEM"
)

var (
	_ fmt.Stringer             = AuditLogActorType("")
	_ encoding.TextMarshaler   = AuditLogActorType("")
	_ encoding.TextUnmarshaler = (*AuditLogActorType)(nil)
)

func AuditLogActorTypes() []AuditLogActorType {
	return []AuditLogActorType{
		AuditLogActorTypeUser,
		AuditLogActorTypeAPIKey,
		AuditLogActorTypeSystem,
	}
}

func (v AuditLogActorType) IsValid() bool {
	switch v {
	case
		AuditLogActorTypeUser,
		AuditLogActorTypeAPIKey,
		AuditLogActorTypeSystem:
		return true
	}

	return false
}

func (v AuditLogActorType) String() string {
	return string(v)
}

func (v AuditLogActorType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AuditLogActorType) UnmarshalText(text []byte) error {
	val := AuditLogActorType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AuditLogActorType value: %q", string(text))
	}

	*v = val

	return nil
}
