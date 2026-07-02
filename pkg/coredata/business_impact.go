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
)

type BusinessImpact string

const (
	BusinessImpactLow      BusinessImpact = "LOW"
	BusinessImpactMedium   BusinessImpact = "MEDIUM"
	BusinessImpactHigh     BusinessImpact = "HIGH"
	BusinessImpactCritical BusinessImpact = "CRITICAL"
)

var (
	_ fmt.Stringer             = BusinessImpact("")
	_ encoding.TextMarshaler   = BusinessImpact("")
	_ encoding.TextUnmarshaler = (*BusinessImpact)(nil)
)

func BusinessImpacts() []BusinessImpact {
	return []BusinessImpact{
		BusinessImpactLow,
		BusinessImpactMedium,
		BusinessImpactHigh,
		BusinessImpactCritical,
	}
}

func (v BusinessImpact) IsValid() bool {
	switch v {
	case
		BusinessImpactLow,
		BusinessImpactMedium,
		BusinessImpactHigh,
		BusinessImpactCritical:
		return true
	}

	return false
}

func (v BusinessImpact) String() string {
	return string(v)
}

func (v BusinessImpact) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *BusinessImpact) UnmarshalText(text []byte) error {
	val := BusinessImpact(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid BusinessImpact value: %q", string(text))
	}

	*v = val

	return nil
}
