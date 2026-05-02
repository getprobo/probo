// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package equal_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/equal"
)

func TestPtr_String(t *testing.T) {
	t.Parallel()

	a := "hello"
	b := "hello"
	c := "world"

	assert.True(t, equal.Ptr[string](nil, nil))
	assert.False(t, equal.Ptr(&a, nil))
	assert.False(t, equal.Ptr[string](nil, &a))
	assert.True(t, equal.Ptr(&a, &b))
	assert.False(t, equal.Ptr(&a, &c))
}

func TestPtr_Int(t *testing.T) {
	t.Parallel()

	a := 42
	b := 42
	c := 99

	assert.True(t, equal.Ptr[int](nil, nil))
	assert.False(t, equal.Ptr(&a, nil))
	assert.False(t, equal.Ptr[int](nil, &a))
	assert.True(t, equal.Ptr(&a, &b))
	assert.False(t, equal.Ptr(&a, &c))
}

func TestPtr_Bool(t *testing.T) {
	t.Parallel()

	tt := true
	tt2 := true
	ff := false

	assert.True(t, equal.Ptr(&tt, &tt2))
	assert.False(t, equal.Ptr(&tt, &ff))
}

func TestJSON(t *testing.T) {
	t.Parallel()

	t.Run("identical bytes", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"foo":"bar","n":1}`)
		b := json.RawMessage(`{"foo":"bar","n":1}`)

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.True(t, eq)
	})

	t.Run("whitespace and key order differences", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"foo":"bar","n":1}`)
		b := json.RawMessage("{\n  \"n\": 1,\n  \"foo\": \"bar\"\n}")

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.True(t, eq)
	})

	t.Run("nested objects with reordered keys", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"a":{"x":1,"y":2},"b":[1,2,3]}`)
		b := json.RawMessage(`{"b":[1,2,3],"a":{"y":2,"x":1}}`)

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.True(t, eq)
	})

	t.Run("array order matters", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"items":[1,2,3]}`)
		b := json.RawMessage(`{"items":[3,2,1]}`)

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.False(t, eq)
	})

	t.Run("real content change", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"title":"Cookies"}`)
		b := json.RawMessage(`{"title":"Cookies updated"}`)

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.False(t, eq)
	})

	t.Run("integers and floats compare equal", func(t *testing.T) {
		t.Parallel()

		a := json.RawMessage(`{"n":1}`)
		b := json.RawMessage(`{"n":1.0}`)

		eq, err := equal.JSON(a, b)
		require.NoError(t, err)
		assert.True(t, eq, "json.Unmarshal yields float64 for both")
	})

	t.Run("invalid first blob returns error", func(t *testing.T) {
		t.Parallel()

		_, err := equal.JSON(json.RawMessage(`not json`), json.RawMessage(`{}`))
		assert.Error(t, err)
	})

	t.Run("invalid second blob returns error", func(t *testing.T) {
		t.Parallel()

		_, err := equal.JSON(json.RawMessage(`{}`), json.RawMessage(`not json`))
		assert.Error(t, err)
	})
}
