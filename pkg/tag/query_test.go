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

package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/validator"
)

func TestParseQuery(t *testing.T) {
	t.Parallel()

	t.Run(
		"two tags and a quoted name",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery(`tag:Production tag:"eu west" type:document`)
			require.NoError(t, err)
			require.Len(t, parsed.Tags, 2)
			assert.Equal(t, "Production", parsed.Tags[0].Name)
			assert.Equal(t, "eu west", parsed.Tags[1].Name)
			assert.Equal(t, []uint16{coredata.DocumentEntityType}, parsed.EntityTypes)
		},
	)

	t.Run(
		"type alone",
		func(t *testing.T) {
			t.Parallel()

			_, err := ParseQuery("type:document")
			assertQueryError(t, err, validator.ErrorCodeInvalidFormat, "at least one word or tag:name is required")
		},
	)

	t.Run(
		"text alone",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery("access control")
			require.NoError(t, err)
			assert.Empty(t, parsed.Tags)
			assert.Equal(t, "access control", parsed.Text)
		},
	)

	t.Run(
		"text with tag",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery("access tag:Production")
			require.NoError(t, err)
			require.Len(t, parsed.Tags, 1)
			assert.Equal(t, "access", parsed.Text)
		},
	)

	t.Run(
		"empty query",
		func(t *testing.T) {
			t.Parallel()

			_, err := ParseQuery("   ")
			assertQueryError(t, err, validator.ErrorCodeRequired, "field is required")
		},
	)

	t.Run(
		"unknown type",
		func(t *testing.T) {
			t.Parallel()

			_, err := ParseQuery("tag:Production type:widget")
			assertQueryError(t, err, validator.ErrorCodeInvalidFormat, `unknown type "widget"`)
		},
	)

	t.Run(
		"two different tags",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery("tag:Production tag:Staging")
			require.NoError(t, err)
			require.Len(t, parsed.Tags, 2)
			assert.Equal(t, "Production", parsed.Tags[0].Name)
			assert.Equal(t, "Staging", parsed.Tags[1].Name)
		},
	)

	t.Run(
		"identical repeated tag",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery("tag:Production tag:Production")
			require.NoError(t, err)
			require.Len(t, parsed.Tags, 1)
			assert.Equal(t, "Production", parsed.Tags[0].Name)
		},
	)

	t.Run(
		"repeated type is kept once",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery("tag:Production type:document type:risk type:document")
			require.NoError(t, err)
			assert.Equal(
				t,
				[]uint16{coredata.DocumentEntityType, coredata.RiskEntityType},
				parsed.EntityTypes,
			)
		},
	)

	t.Run(
		"key equals value syntax is rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := ParseQuery("tag:env=Production")
			assertQueryError(t, err, validator.ErrorCodeInvalidFormat, "unexpected token")
		},
	)

	t.Run(
		"quoted type name",
		func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseQuery(`tag:Production type:"risk-analysis"`)
			require.NoError(t, err)
			assert.Equal(t, []uint16{coredata.RiskAnalysisEntityType}, parsed.EntityTypes)
		},
	)
}

func TestEntityTypeNames_CoverAllowlist(t *testing.T) {
	t.Parallel()

	require.Len(t, entityTypeByName, len(TaggableEntityTypes))

	seen := make(map[uint16]struct{}, len(entityTypeByName))
	for _, entityType := range entityTypeByName {
		seen[entityType] = struct{}{}
	}

	for _, entityType := range TaggableEntityTypes {
		_, ok := seen[entityType]
		assert.True(t, ok, "entity type %d has no query name", entityType)
	}
}

func assertQueryError(t *testing.T, err error, code validator.ErrorCode, message string) {
	t.Helper()

	var validationErrors validator.ValidationErrors

	require.ErrorAs(t, err, &validationErrors)
	require.Len(t, validationErrors, 1)
	assert.Equal(t, "query", validationErrors[0].Field)
	assert.Equal(t, code, validationErrors[0].Code)
	assert.Equal(t, message, validationErrors[0].Message)
}
