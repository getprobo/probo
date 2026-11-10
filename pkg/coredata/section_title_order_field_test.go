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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

const sectionTitleCursorPlaceholder = "@cursor_field_value"

func sectionTitleOrderDirections() []page.OrderDirection {
	return []page.OrderDirection{page.OrderDirectionAsc, page.OrderDirectionDesc}
}

func TestControlOrderField_SectionTitle(t *testing.T) {
	t.Parallel()

	t.Run("section title maps the cursor value like the column", func(t *testing.T) {
		t.Parallel()

		field := coredata.ControlOrderFieldSectionTitle

		for _, direction := range sectionTitleOrderDirections() {
			assert.Equal(t, "section_title_sort_key(section_title)", field.DirectionalColumn(direction))
			assert.Equal(
				t,
				"section_title_sort_key("+sectionTitleCursorPlaceholder+")",
				field.DirectionalCursorValue(direction, sectionTitleCursorPlaceholder),
			)
		}
	})

	t.Run("other fields bind the cursor value unchanged", func(t *testing.T) {
		t.Parallel()

		field := coredata.ControlOrderFieldCreatedAt

		for _, direction := range sectionTitleOrderDirections() {
			assert.Equal(t, "created_at", field.DirectionalColumn(direction))
			assert.Equal(
				t,
				sectionTitleCursorPlaceholder,
				field.DirectionalCursorValue(direction, sectionTitleCursorPlaceholder),
			)
		}
	})
}

func TestApplicabilityStatementOrderField_ControlSectionTitle(t *testing.T) {
	t.Parallel()

	t.Run("section title maps the cursor value like the column", func(t *testing.T) {
		t.Parallel()

		field := coredata.ApplicabilityStatementOrderFieldControlSectionTitle

		for _, direction := range sectionTitleOrderDirections() {
			assert.Equal(t, "section_title_sort_key(section_title)", field.DirectionalColumn(direction))
			assert.Equal(
				t,
				"section_title_sort_key("+sectionTitleCursorPlaceholder+")",
				field.DirectionalCursorValue(direction, sectionTitleCursorPlaceholder),
			)
		}
	})

	t.Run("other fields bind the cursor value unchanged", func(t *testing.T) {
		t.Parallel()

		field := coredata.ApplicabilityStatementOrderFieldCreatedAt

		for _, direction := range sectionTitleOrderDirections() {
			assert.Equal(t, "created_at", field.DirectionalColumn(direction))
			assert.Equal(
				t,
				sectionTitleCursorPlaceholder,
				field.DirectionalCursorValue(direction, sectionTitleCursorPlaceholder),
			)
		}
	})
}
