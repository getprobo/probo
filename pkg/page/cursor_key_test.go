// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package page

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
)

// b64 encodes a raw JSON payload the way a cursor string is wire-encoded, so
// error-path tests can feed deliberately malformed cursors.
func b64(payload string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func TestCursorKeyRoundTrip(t *testing.T) {
	t.Parallel()

	id := mkID(7)

	tests := []struct {
		name      string
		value     any
		wantValue any // value after a wire round-trip (JSON-decoded type)
	}{
		{"nil value", nil, nil},
		{"string value", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z"},
		// JSON has no integer type: numbers decode back as float64.
		{"int value", 42, float64(42)},
		{"float value", 3.5, 3.5},
		{"bool value", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key := NewCursorKey(id, tt.value)

			// String <-> ParseCursorKey.
			parsed, err := ParseCursorKey(key.String())
			require.NoError(t, err)
			assert.Equal(t, id, parsed.ID)
			assert.Equal(t, tt.wantValue, parsed.Value)

			// Bytes == MarshalBinary, and UnmarshalBinary reverses it.
			bin, err := key.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, bin, key.Bytes())

			var fromBin CursorKey
			require.NoError(t, fromBin.UnmarshalBinary(bin))
			assert.Equal(t, id, fromBin.ID)
			assert.Equal(t, tt.wantValue, fromBin.Value)

			// MarshalText <-> UnmarshalText.
			txt, err := key.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, key.String(), string(txt))

			var fromTxt CursorKey
			require.NoError(t, fromTxt.UnmarshalText(txt))
			assert.Equal(t, parsed, fromTxt)

			// MarshalJSON wraps the opaque string; UnmarshalJSON reverses it.
			js, err := json.Marshal(key)
			require.NoError(t, err)

			var fromJS CursorKey
			require.NoError(t, json.Unmarshal(js, &fromJS))
			assert.Equal(t, parsed, fromJS)

			assert.Equal(t, tt.value, key.FieldValue())
		})
	}
}

// TestParseCursorKeyErrors covers every rejection path: a client-supplied
// cursor must fail with ErrInvalidFormat rather than panicking or silently
// decoding to a bogus key.
func TestParseCursorKeyErrors(t *testing.T) {
	t.Parallel()

	validID := mkID(1).String()

	tests := []struct {
		name  string
		input string
	}{
		{"not base64", "not*base64*"},
		{"not a json array", b64(`"just a string"`)},
		{"array too short", b64(`["` + validID + `"]`)},
		{"array too long", b64(`["` + validID + `", 1, 2]`)},
		{"id not a string", b64(`[123, 1]`)},
		{"id not a valid gid", b64(`["not-a-gid", 1]`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseCursorKey(tt.input)
			assert.ErrorIs(t, err, ErrInvalidFormat)
		})
	}
}

func TestCursorKeyUnmarshalRejectsInvalid(t *testing.T) {
	t.Parallel()

	var ck CursorKey

	assert.Error(t, ck.UnmarshalText([]byte("not*base64*")))
	assert.Error(t, ck.UnmarshalBinary([]byte(`["only-one"]`)))
	assert.Error(t, ck.UnmarshalJSON([]byte(`{"not":"a string"}`)))
}

func TestParseGIDCursorKeyValue(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	key := NewCursorKey(id, "x")

	parsed, err := ParseCursorKey(key.String())
	require.NoError(t, err)
	assert.Equal(t, id, parsed.ID)
}
