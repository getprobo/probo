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

package scim

import (
	"testing"

	scimfilter "github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUserFilter(t *testing.T) {
	t.Run("nil expression returns empty filter", func(t *testing.T) {
		filter, err := ParseUserFilter(nil)
		require.NoError(t, err)
		require.NotNil(t, filter)
		assert.Nil(t, filter.Email())
	})

	t.Run("simple userName eq filter", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`userName eq "test@example.com"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		require.NoError(t, err)
		require.NotNil(t, filter)
		require.NotNil(t, filter.Email())
		assert.Equal(t, "test@example.com", filter.Email().String())
	})

	t.Run("userName filter is case insensitive", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`UserName eq "test@example.com"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		require.NoError(t, err)
		require.NotNil(t, filter)
		require.NotNil(t, filter.Email())
		assert.Equal(t, "test@example.com", filter.Email().String())
	})

	t.Run("logical AND expression", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`userName eq "user1@example.com" and userName eq "user2@example.com"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		require.NoError(t, err)
		require.NotNil(t, filter)
		// Last processed value wins
		require.NotNil(t, filter.Email())
	})

	t.Run("unsupported operator returns error", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`userName co "test"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.Contains(t, err.Error(), "operator")
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("unsupported attribute returns error", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`displayName eq "John"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.Contains(t, err.Error(), "displayName")
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("OR operator returns error", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`userName eq "a@b.com" or userName eq "c@d.com"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.Contains(t, err.Error(), "logical operator")
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("NOT expression returns error", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`not (userName eq "test@example.com")`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.Contains(t, err.Error(), "NOT expressions are not supported")
	})

	t.Run("nested AND expressions", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`(userName eq "a@b.com" and userName eq "c@d.com") and userName eq "e@f.com"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		require.NoError(t, err)
		require.NotNil(t, filter)
		require.NotNil(t, filter.Email())
	})

	t.Run("invalid email format returns error", func(t *testing.T) {
		expr, err := scimfilter.ParseFilter([]byte(`userName eq "not-an-email"`))
		require.NoError(t, err)

		filter, err := ParseUserFilter(expr)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}
