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

package soagen

type FieldConfiguration struct {
	Field         string
	Columns       []string
	FilterColumns []string  // Specific columns that should have filters (subset of Columns)
	Width         []float64 // Width for each column
	DefaultWidth  float64   // Default width if not specified
	HasFilter     bool      // Whether to enable auto-filter
}

var (
	soaFieldConfigs = []FieldConfiguration{
		{
			Field:         "SectionTitle",
			Columns:       []string{"A"},
			FilterColumns: []string{},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     false,
		},
		{
			Field:         "ControlName",
			Columns:       []string{"B"},
			FilterColumns: []string{},
			Width:         []float64{35},
			DefaultWidth:  35,
			HasFilter:     false,
		},
		{
			Field:         "Applicability",
			Columns:       []string{"C"},
			FilterColumns: []string{"C"},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     true,
		},
		{
			Field:         "ExclusionJustification",
			Columns:       []string{"D"},
			FilterColumns: []string{},
			Width:         []float64{25},
			DefaultWidth:  25,
			HasFilter:     false,
		},
		{
			Field:         "Regulatory",
			Columns:       []string{"E"},
			FilterColumns: []string{},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     false,
		},
		{
			Field:         "Contractual",
			Columns:       []string{"F"},
			FilterColumns: []string{},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     false,
		},
		{
			Field:         "BestPractice",
			Columns:       []string{"G"},
			FilterColumns: []string{},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     false,
		},
		{
			Field:         "RiskAssessment",
			Columns:       []string{"H"},
			FilterColumns: []string{},
			Width:         []float64{12},
			DefaultWidth:  12,
			HasFilter:     false,
		},
		{
			Field:         "SecurityMeasures",
			Columns:       []string{"I"},
			FilterColumns: []string{},
			Width:         []float64{40},
			DefaultWidth:  40,
			HasFilter:     false,
		},
	}
)
