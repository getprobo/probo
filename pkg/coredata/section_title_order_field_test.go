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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.probo.inc/probo/pkg/gid"
)

func TestSectionTitleSortKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title string
		want  string
	}{
		{title: "A.5", want: "A.0000000005"},
		{title: "A.10", want: "A.0000000010"},
		{title: "A.5.12 foo", want: "A.0000000005.0000000012 foo"},
		{title: "12abc", want: "0000000012abc"},
		{title: "a1b2", want: "a0000000001b0000000002"},
		{title: "hello", want: "hello"},
		{title: "", want: ""},
		{title: "A.12345678901", want: "A.1234567890"},
		{title: "10", want: "0000000010"},
		{title: "2", want: "0000000002"},
		{title: "é2", want: "é0000000002"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, sectionTitleSortKey(tt.title))
		})
	}
}

func TestControl_CursorKey_SectionTitle(t *testing.T) {
	t.Parallel()

	control := Control{
		ID:           gid.New(gid.NewTenantID(), ControlEntityType),
		SectionTitle: "A.10",
	}

	assert.Equal(t, "A.0000000010", control.CursorKey(ControlOrderFieldSectionTitle).Value)
	assert.Equal(t, "section_title_sort_key(section_title)", ControlOrderFieldSectionTitle.Column())
}

func TestApplicabilityStatement_CursorKey_SectionTitle(t *testing.T) {
	t.Parallel()

	statement := ApplicabilityStatement{
		ID:           gid.New(gid.NewTenantID(), ApplicabilityStatementEntityType),
		SectionTitle: "A.5.12 foo",
	}

	assert.Equal(
		t,
		"A.0000000005.0000000012 foo",
		statement.CursorKey(ApplicabilityStatementOrderFieldControlSectionTitle).Value,
	)
	assert.Equal(
		t,
		"section_title_sort_key(section_title)",
		ApplicabilityStatementOrderFieldControlSectionTitle.Column(),
	)
}
