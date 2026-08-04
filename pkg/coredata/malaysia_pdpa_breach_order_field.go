// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type MalaysiaPDPABreachOrderField string

const (
	MalaysiaPDPABreachOrderFieldCreatedAt   MalaysiaPDPABreachOrderField = "CREATED_AT"
	MalaysiaPDPABreachOrderFieldUpdatedAt   MalaysiaPDPABreachOrderField = "UPDATED_AT"
	MalaysiaPDPABreachOrderFieldAwarenessAt MalaysiaPDPABreachOrderField = "AWARENESS_AT"
	MalaysiaPDPABreachOrderFieldTitle       MalaysiaPDPABreachOrderField = "TITLE"
	MalaysiaPDPABreachOrderFieldStatus      MalaysiaPDPABreachOrderField = "STATUS"
)

var (
	_ page.OrderField          = MalaysiaPDPABreachOrderField("")
	_ fmt.Stringer             = MalaysiaPDPABreachOrderField("")
	_ encoding.TextMarshaler   = MalaysiaPDPABreachOrderField("")
	_ encoding.TextUnmarshaler = (*MalaysiaPDPABreachOrderField)(nil)
)

func MalaysiaPDPABreachOrderFields() []MalaysiaPDPABreachOrderField {
	return []MalaysiaPDPABreachOrderField{
		MalaysiaPDPABreachOrderFieldCreatedAt,
		MalaysiaPDPABreachOrderFieldUpdatedAt,
		MalaysiaPDPABreachOrderFieldAwarenessAt,
		MalaysiaPDPABreachOrderFieldTitle,
		MalaysiaPDPABreachOrderFieldStatus,
	}
}

func (v MalaysiaPDPABreachOrderField) IsValid() bool {
	switch v {
	case MalaysiaPDPABreachOrderFieldCreatedAt,
		MalaysiaPDPABreachOrderFieldUpdatedAt,
		MalaysiaPDPABreachOrderFieldAwarenessAt,
		MalaysiaPDPABreachOrderFieldTitle,
		MalaysiaPDPABreachOrderFieldStatus:
		return true
	}

	return false
}

func (v MalaysiaPDPABreachOrderField) String() string { return string(v) }

func (v MalaysiaPDPABreachOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MalaysiaPDPABreachOrderField) UnmarshalText(text []byte) error {
	value := MalaysiaPDPABreachOrderField(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid MalaysiaPDPABreachOrderField value: %q", string(text))
	}

	*v = value

	return nil
}

func (v MalaysiaPDPABreachOrderField) Column() string { return string(v) }
