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
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type BusinessImpact string

const (
	BusinessImpactLow      BusinessImpact = "LOW"
	BusinessImpactMedium   BusinessImpact = "MEDIUM"
	BusinessImpactHigh     BusinessImpact = "HIGH"
	BusinessImpactCritical BusinessImpact = "CRITICAL"
)

func BusinessImpacts() []BusinessImpact {
	return []BusinessImpact{
		BusinessImpactLow,
		BusinessImpactMedium,
		BusinessImpactHigh,
		BusinessImpactCritical,
	}
}

func (i BusinessImpact) String() string {
	return string(i)
}

func (i *BusinessImpact) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		switch v {
		case "LOW":
			*i = BusinessImpactLow
		case "MEDIUM":
			*i = BusinessImpactMedium
		case "HIGH":
			*i = BusinessImpactHigh
		case "CRITICAL":
			*i = BusinessImpactCritical
		default:
			return fmt.Errorf("invalid BusinessImpact value: %q", v)
		}
	default:
		return fmt.Errorf("unsupported type for BusinessImpact: %T", value)
	}
	return nil
}

func (i BusinessImpact) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i BusinessImpact) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *BusinessImpact) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "LOW":
		*i = BusinessImpactLow
	case "MEDIUM":
		*i = BusinessImpactMedium
	case "HIGH":
		*i = BusinessImpactHigh
	case "CRITICAL":
		*i = BusinessImpactCritical
	default:
		return fmt.Errorf("invalid BusinessImpact value: %q", s)
	}
	return nil
}

func (i *BusinessImpact) UnmarshalText(text []byte) error {
	s := string(text)
	switch s {
	case "LOW":
		*i = BusinessImpactLow
	case "MEDIUM":
		*i = BusinessImpactMedium
	case "HIGH":
		*i = BusinessImpactHigh
	case "CRITICAL":
		*i = BusinessImpactCritical
	default:
		return fmt.Errorf("invalid BusinessImpact value: %q", s)
	}
	return nil
}
