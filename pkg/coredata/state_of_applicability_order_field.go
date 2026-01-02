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

type (
	StateOfApplicabilityOrderField string
)

const (
	StateOfApplicabilityOrderFieldName      StateOfApplicabilityOrderField = "NAME"
	StateOfApplicabilityOrderFieldCreatedAt StateOfApplicabilityOrderField = "CREATED_AT"
)

func (s StateOfApplicabilityOrderField) Column() string {
	switch s {
	case StateOfApplicabilityOrderFieldName:
		return "name"
	case StateOfApplicabilityOrderFieldCreatedAt:
		return "created_at"
	}
	panic(fmt.Sprintf("unsupported order by: %s", s))
}

func (s StateOfApplicabilityOrderField) String() string {
	return string(s)
}

func (s StateOfApplicabilityOrderField) IsValid() bool {
	switch s {
	case StateOfApplicabilityOrderFieldName, StateOfApplicabilityOrderFieldCreatedAt:
		return true
	}
	return false
}

func (s StateOfApplicabilityOrderField) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *StateOfApplicabilityOrderField) UnmarshalText(text []byte) error {
	*s = StateOfApplicabilityOrderField(text)
	if !s.IsValid() {
		return fmt.Errorf("%s is not a valid StateOfApplicabilityOrderField", string(text))
	}
	return nil
}
