// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMargin(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		unit  MarginUnit
		want  Margin
	}{
		{
			name:  "inches",
			value: 1.5,
			unit:  MarginUnitInch,
			want:  Margin{Value: 1.5, Unit: MarginUnitInch},
		},
		{
			name:  "millimeters",
			value: 25.4,
			unit:  MarginUnitMillimeter,
			want:  Margin{Value: 25.4, Unit: MarginUnitMillimeter},
		},
		{
			name:  "centimeters",
			value: 2.54,
			unit:  MarginUnitCentimeter,
			want:  Margin{Value: 2.54, Unit: MarginUnitCentimeter},
		},
		{
			name:  "points",
			value: 72.0,
			unit:  MarginUnitPoint,
			want:  Margin{Value: 72.0, Unit: MarginUnitPoint},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMargin(tt.value, tt.unit)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMarginInches(t *testing.T) {
	value := 2.0
	expected := Margin{Value: 2.0, Unit: MarginUnitInch}
	got := NewMarginInches(value)
	assert.Equal(t, expected, got)
}

func TestNewMarginMillimeters(t *testing.T) {
	value := 50.8
	expected := Margin{Value: 50.8, Unit: MarginUnitMillimeter}
	got := NewMarginMillimeters(value)
	assert.Equal(t, expected, got)
}

func TestNewMarginCentimeters(t *testing.T) {
	value := 5.08
	expected := Margin{Value: 5.08, Unit: MarginUnitCentimeter}
	got := NewMarginCentimeters(value)
	assert.Equal(t, expected, got)
}

func TestNewMarginPoints(t *testing.T) {
	value := 144.0
	expected := Margin{Value: 144.0, Unit: MarginUnitPoint}
	got := NewMarginPoints(value)
	assert.Equal(t, expected, got)
}

func TestParseMargin(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Margin
	}{
		{
			name:  "empty string returns default",
			input: "",
			want:  NewMarginInches(1.0),
		},
		{
			name:  "inches with unit",
			input: "2.5in",
			want:  NewMarginInches(2.5),
		},
		{
			name:  "millimeters with unit",
			input: "25.4mm",
			want:  NewMarginMillimeters(25.4),
		},
		{
			name:  "centimeters with unit",
			input: "2.54cm",
			want:  NewMarginCentimeters(2.54),
		},
		{
			name:  "points with unit",
			input: "72pt",
			want:  NewMarginPoints(72),
		},
		{
			name:  "plain number assumes inches",
			input: "1.5",
			want:  NewMarginInches(1.5),
		},
		{
			name:  "whitespace is trimmed",
			input: "  1.5in  ",
			want:  NewMarginInches(1.5),
		},
		{
			name:  "integer values",
			input: "1in",
			want:  NewMarginInches(1.0),
		},
		{
			name:  "zero values",
			input: "0mm",
			want:  NewMarginMillimeters(0.0),
		},
		{
			name:  "invalid input returns default",
			input: "invalid",
			want:  NewMarginInches(1.0),
		},
		{
			name:  "invalid number with valid unit returns default",
			input: "invalidmm",
			want:  NewMarginInches(1.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMargin(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMargin_ToInches(t *testing.T) {
	tests := []struct {
		name   string
		margin Margin
		want   float64
	}{
		{
			name:   "inches to inches",
			margin: NewMarginInches(2.0),
			want:   2.0,
		},
		{
			name:   "millimeters to inches",
			margin: NewMarginMillimeters(25.4),
			want:   1.0,
		},
		{
			name:   "centimeters to inches",
			margin: NewMarginCentimeters(2.54),
			want:   1.0,
		},
		{
			name:   "points to inches",
			margin: NewMarginPoints(72.0),
			want:   1.0,
		},
		{
			name:   "fractional values",
			margin: NewMarginMillimeters(12.7), // 0.5 inches
			want:   0.5,
		},
		{
			name:   "zero value",
			margin: NewMarginInches(0.0),
			want:   0.0,
		},
		{
			name:   "unknown unit defaults to value",
			margin: Margin{Value: 2.5, Unit: MarginUnit("unknown")},
			want:   2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.margin.ToInches()
			assert.InDelta(t, tt.want, got, 0.0001) // Allow small floating point errors
		})
	}
}

func TestMargin_String(t *testing.T) {
	tests := []struct {
		name   string
		margin Margin
		want   string
	}{
		{
			name:   "inches",
			margin: NewMarginInches(1.5),
			want:   "1.50in",
		},
		{
			name:   "millimeters",
			margin: NewMarginMillimeters(25.4),
			want:   "25.40mm",
		},
		{
			name:   "centimeters",
			margin: NewMarginCentimeters(2.54),
			want:   "2.54cm",
		},
		{
			name:   "points",
			margin: NewMarginPoints(72.0),
			want:   "72.00pt",
		},
		{
			name:   "zero value",
			margin: NewMarginInches(0.0),
			want:   "0.00in",
		},
		{
			name:   "integer value shows decimal",
			margin: NewMarginInches(1.0),
			want:   "1.00in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.margin.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMarginConstants(t *testing.T) {
	// Test that the margin unit constants are properly defined
	assert.Equal(t, MarginUnit("in"), MarginUnitInch)
	assert.Equal(t, MarginUnit("mm"), MarginUnitMillimeter)
	assert.Equal(t, MarginUnit("cm"), MarginUnitCentimeter)
	assert.Equal(t, MarginUnit("pt"), MarginUnitPoint)
}

func TestMarginConversions(t *testing.T) {
	// Test conversion accuracy between units
	tests := []struct {
		name     string
		original Margin
		expected float64 // expected inches
	}{
		{
			name:     "1 inch = 25.4 mm",
			original: NewMarginMillimeters(25.4),
			expected: 1.0,
		},
		{
			name:     "1 inch = 2.54 cm",
			original: NewMarginCentimeters(2.54),
			expected: 1.0,
		},
		{
			name:     "1 inch = 72 points",
			original: NewMarginPoints(72.0),
			expected: 1.0,
		},
		{
			name:     "0.5 inch = 12.7 mm",
			original: NewMarginMillimeters(12.7),
			expected: 0.5,
		},
		{
			name:     "0.5 inch = 1.27 cm",
			original: NewMarginCentimeters(1.27),
			expected: 0.5,
		},
		{
			name:     "0.5 inch = 36 points",
			original: NewMarginPoints(36.0),
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.original.ToInches()
			assert.InDelta(t, tt.expected, got, 0.0001)
		})
	}
}

// Benchmark tests
func BenchmarkParseMargin(b *testing.B) {
	testCases := []string{
		"1.5in",
		"25.4mm",
		"2.54cm",
		"72pt",
		"1.5",
		"",
		"invalid",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			ParseMargin(tc)
		}
	}
}

func BenchmarkMarginToInches(b *testing.B) {
	margins := []Margin{
		NewMarginInches(1.5),
		NewMarginMillimeters(25.4),
		NewMarginCentimeters(2.54),
		NewMarginPoints(72.0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, margin := range margins {
			margin.ToInches()
		}
	}
}

func BenchmarkMarginString(b *testing.B) {
	margins := []Margin{
		NewMarginInches(1.5),
		NewMarginMillimeters(25.4),
		NewMarginCentimeters(2.54),
		NewMarginPoints(72.0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, margin := range margins {
			_ = margin.String()
		}
	}
}
