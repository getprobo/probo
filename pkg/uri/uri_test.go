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

package uri

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    URI
		wantErr bool
	}{
		{
			name:  "valid https url",
			input: "https://example.com",
			want:  URI("https://example.com"),
		},
		{
			name:  "valid https url with path",
			input: "https://example.com/callback",
			want:  URI("https://example.com/callback"),
		},
		{
			name:  "valid https url with port",
			input: "https://localhost:8080/callback",
			want:  URI("https://localhost:8080/callback"),
		},
		{
			name:  "valid http url",
			input: "http://localhost:3000/auth/callback",
			want:  URI("http://localhost:3000/auth/callback"),
		},
		{
			name:  "valid url with query",
			input: "https://example.com/path?key=value",
			want:  URI("https://example.com/path?key=value"),
		},
		{
			name:  "valid custom scheme",
			input: "myapp://callback",
			want:  URI("myapp://callback"),
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no scheme",
			input:   "example.com/callback",
			wantErr: true,
		},
		{
			name:    "no host",
			input:   "/callback",
			wantErr: true,
		},
		{
			name:    "relative path",
			input:   "callback",
			wantErr: true,
		},
		{
			name:    "scheme only",
			input:   "https://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := Parse(tt.input)
				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestURIUnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run(
		"valid",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.UnmarshalText([]byte("https://example.com/callback"))
			require.NoError(t, err)
			assert.Equal(t, URI("https://example.com/callback"), u)
		},
	)

	t.Run(
		"invalid",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.UnmarshalText([]byte("not-a-url"))
			require.Error(t, err)
		},
	)
}

func TestURIMarshalText(t *testing.T) {
	t.Parallel()

	u := URI("https://example.com/callback")
	b, err := u.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("https://example.com/callback"), b)
}

func TestURIJSON(t *testing.T) {
	t.Parallel()

	t.Run(
		"marshal",
		func(t *testing.T) {
			t.Parallel()

			v := struct {
				URL URI `json:"url"`
			}{URL: URI("https://example.com")}

			data, err := json.Marshal(v)
			require.NoError(t, err)
			assert.JSONEq(t, `{"url":"https://example.com"}`, string(data))
		},
	)

	t.Run(
		"unmarshal valid",
		func(t *testing.T) {
			t.Parallel()

			var v struct {
				URL URI `json:"url"`
			}

			err := json.Unmarshal([]byte(`{"url":"https://example.com"}`), &v)
			require.NoError(t, err)
			assert.Equal(t, URI("https://example.com"), v.URL)
		},
	)

	t.Run(
		"unmarshal invalid",
		func(t *testing.T) {
			t.Parallel()

			var v struct {
				URL URI `json:"url"`
			}

			err := json.Unmarshal([]byte(`{"url":"not-a-url"}`), &v)
			require.Error(t, err)
		},
	)
}

func TestURIScan(t *testing.T) {
	t.Parallel()

	t.Run(
		"string",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.Scan("https://example.com")
			require.NoError(t, err)
			assert.Equal(t, URI("https://example.com"), u)
		},
	)

	t.Run(
		"bytes",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.Scan([]byte("https://example.com"))
			require.NoError(t, err)
			assert.Equal(t, URI("https://example.com"), u)
		},
	)

	t.Run(
		"invalid value",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.Scan("not-a-url")
			require.Error(t, err)
		},
	)

	t.Run(
		"unsupported type",
		func(t *testing.T) {
			t.Parallel()

			var u URI
			err := u.Scan(123)
			require.Error(t, err)
		},
	)
}

func TestURIValue(t *testing.T) {
	t.Parallel()

	u := URI("https://example.com")
	v, err := u.Value()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", v)
}
