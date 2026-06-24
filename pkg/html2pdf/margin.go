// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
