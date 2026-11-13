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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestControlOrderField_SectionTitle(t *testing.T) {
	t.Parallel()

	t.Run("section title sorts the raw cursor value with the column function", func(t *testing.T) {
		t.Parallel()

		id := gid.New(gid.NewTenantID(), coredata.ControlEntityType)
		key := page.NewCursorKey(id, "A.10")
		cursor := page.NewCursor(10, &key, page.Head, page.OrderBy[coredata.ControlOrderField]{
			Field:     coredata.ControlOrderFieldSectionTitle,
			Direction: page.OrderDirectionAsc,
		})

		assert.Contains(t, cursor.SQLFragment(), "section_title_sort_key(section_title) >= section_title_sort_key(@cursor_field_value)")
		args, err := cursor.SQLArguments()
		assert.NoError(t, err)
		assert.Equal(t, "A.10", args["cursor_field_value"])
	})

	t.Run("the cursor key stores the raw section title", func(t *testing.T) {
		t.Parallel()

		control := coredata.Control{
			ID:           gid.New(gid.NewTenantID(), coredata.ControlEntityType),
			SectionTitle: "A.10",
		}

		assert.Equal(t, "A.10", control.CursorKey(coredata.ControlOrderFieldSectionTitle).Value)
	})

	t.Run("other fields bind the cursor value unchanged", func(t *testing.T) {
		t.Parallel()

		id := gid.New(gid.NewTenantID(), coredata.ControlEntityType)
		key := page.NewCursorKey(id, "created")
		cursor := page.NewCursor(10, &key, page.Head, page.OrderBy[coredata.ControlOrderField]{
			Field:     coredata.ControlOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		})

		assert.Contains(t, cursor.SQLFragment(), "created_at >= @cursor_field_value")
		assert.NotContains(t, cursor.SQLFragment(), "section_title_sort_key(@cursor_field_value)")
	})
}

func TestApplicabilityStatementOrderField_ControlSectionTitle(t *testing.T) {
	t.Parallel()

	t.Run("section title sorts the raw cursor value with the column function", func(t *testing.T) {
		t.Parallel()

		id := gid.New(gid.NewTenantID(), coredata.ApplicabilityStatementEntityType)
		key := page.NewCursorKey(id, "A.5.12 foo")
		cursor := page.NewCursor(10, &key, page.Head, page.OrderBy[coredata.ApplicabilityStatementOrderField]{
			Field:     coredata.ApplicabilityStatementOrderFieldControlSectionTitle,
			Direction: page.OrderDirectionAsc,
		})

		assert.Contains(t, cursor.SQLFragment(), "section_title_sort_key(section_title) >= section_title_sort_key(@cursor_field_value)")
		args, err := cursor.SQLArguments()
		assert.NoError(t, err)
		assert.Equal(t, "A.5.12 foo", args["cursor_field_value"])
	})

	t.Run("the cursor key stores the raw section title", func(t *testing.T) {
		t.Parallel()

		statement := coredata.ApplicabilityStatement{
			ID:           gid.New(gid.NewTenantID(), coredata.ApplicabilityStatementEntityType),
			SectionTitle: "A.5.12 foo",
		}

		assert.Equal(
			t,
			"A.5.12 foo",
			statement.CursorKey(coredata.ApplicabilityStatementOrderFieldControlSectionTitle).Value,
		)
	})

	t.Run("other fields bind the cursor value unchanged", func(t *testing.T) {
		t.Parallel()

		id := gid.New(gid.NewTenantID(), coredata.ApplicabilityStatementEntityType)
		key := page.NewCursorKey(id, "created")
		cursor := page.NewCursor(10, &key, page.Head, page.OrderBy[coredata.ApplicabilityStatementOrderField]{
			Field:     coredata.ApplicabilityStatementOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		})

		assert.Contains(t, cursor.SQLFragment(), "created_at >= @cursor_field_value")
		assert.NotContains(t, cursor.SQLFragment(), "section_title_sort_key(@cursor_field_value)")
	})
}
