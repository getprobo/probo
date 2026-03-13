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

package html2pdf

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	MarginUnit string

	Margin struct {
		Value float64
		Unit  MarginUnit
	}
)

const (
	MarginUnitInch       MarginUnit = "in"
	MarginUnitMillimeter MarginUnit = "mm"
	MarginUnitCentimeter MarginUnit = "cm"
	MarginUnitPoint      MarginUnit = "pt"
)

// NewMargin creates a new margin with the specified value and unit
func NewMargin(value float64, unit MarginUnit) Margin {
	return Margin{Value: value, Unit: unit}
}

// NewMarginInches creates a new margin in inches
func NewMarginInches(value float64) Margin {
	return Margin{Value: value, Unit: MarginUnitInch}
}

// NewMarginMillimeters creates a new margin in millimeters
func NewMarginMillimeters(value float64) Margin {
	return Margin{Value: value, Unit: MarginUnitMillimeter}
}

// NewMarginCentimeters creates a new margin in centimeters
func NewMarginCentimeters(value float64) Margin {
	return Margin{Value: value, Unit: MarginUnitCentimeter}
}

// NewMarginPoints creates a new margin in points
func NewMarginPoints(value float64) Margin {
	return Margin{Value: value, Unit: MarginUnitPoint}
}

// ParseMargin parses a CSS margin string into a Margin
func ParseMargin(margin string) Margin {
	if margin == "" {
		return NewMarginInches(1.0) // Default 1 inch
	}

	margin = strings.TrimSpace(margin)

	// Handle different units
	if before, ok := strings.CutSuffix(margin, "in"); ok {
		if val, err := strconv.ParseFloat(before, 64); err == nil {
			return NewMarginInches(val)
		}
	} else if before, ok := strings.CutSuffix(margin, "mm"); ok {
		if val, err := strconv.ParseFloat(before, 64); err == nil {
			return NewMarginMillimeters(val)
		}
	} else if before, ok := strings.CutSuffix(margin, "cm"); ok {
		if val, err := strconv.ParseFloat(before, 64); err == nil {
			return NewMarginCentimeters(val)
		}
	} else if before, ok := strings.CutSuffix(margin, "pt"); ok {
		if val, err := strconv.ParseFloat(before, 64); err == nil {
			return NewMarginPoints(val)
		}
	} else {
		// Try to parse as plain number (assume inches)
		if val, err := strconv.ParseFloat(margin, 64); err == nil {
			return NewMarginInches(val)
		}
	}

	return NewMarginInches(1.0) // Default fallback
}

// ToInches converts the margin to inches (required by Chrome DevTools Protocol)
func (m Margin) ToInches() float64 {
	switch m.Unit {
	case MarginUnitInch:
		return m.Value
	case MarginUnitMillimeter:
		return m.Value / 25.4
	case MarginUnitCentimeter:
		return m.Value / 2.54
	case MarginUnitPoint:
		return m.Value / 72.0
	default:
		return m.Value // Assume inches if unknown unit
	}
}

// String returns the margin as a CSS string
func (m Margin) String() string {
	return fmt.Sprintf("%.2f%s", m.Value, string(m.Unit))
}
