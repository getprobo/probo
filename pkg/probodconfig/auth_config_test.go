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

package probodconfig_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/probodconfig"
)

func TestParseCookieSameSite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    probodconfig.CookieSameSite
		wantErr bool
	}{
		{name: "empty", input: "", want: probodconfig.CookieSameSiteLax},
		{name: "lax", input: "lax", want: probodconfig.CookieSameSiteLax},
		{name: "Lax", input: "Lax", want: probodconfig.CookieSameSiteLax},
		{name: "strict", input: "strict", want: probodconfig.CookieSameSiteStrict},
		{name: "none", input: "none", want: probodconfig.CookieSameSiteNone},
		{name: "invalid", input: "cross-site", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := probodconfig.ParseCookieSameSite(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCookieConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cookie  probodconfig.CookieConfig
		wantErr bool
	}{
		{
			name: "lax without secure",
			cookie: probodconfig.CookieConfig{
				SameSite: probodconfig.CookieSameSiteLax,
				Secure:   false,
			},
		},
		{
			name: "none requires secure",
			cookie: probodconfig.CookieConfig{
				SameSite: probodconfig.CookieSameSiteNone,
				Secure:   false,
			},
			wantErr: true,
		},
		{
			name: "none with secure",
			cookie: probodconfig.CookieConfig{
				SameSite: probodconfig.CookieSameSiteNone,
				Secure:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cookie.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCookieConfig_HTTPSameSite(t *testing.T) {
	t.Parallel()

	cookie := probodconfig.CookieConfig{SameSite: probodconfig.CookieSameSiteStrict}

	got, err := cookie.HTTPSameSite()
	require.NoError(t, err)
	assert.Equal(t, http.SameSiteStrictMode, got)
}
