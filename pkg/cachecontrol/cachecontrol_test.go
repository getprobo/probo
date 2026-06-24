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

package cachecontrol_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cachecontrol"
)

func TestParseRequestDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		wantErr bool
		want    *cachecontrol.TokenPair
	}{
		{
			name:   "no-store flag",
			source: "no-store",
			want:   &cachecontrol.TokenPair{Name: "no-store"},
		},
		{
			name:   "max-age token",
			source: "max-age=4649",
			want:   &cachecontrol.TokenPair{Name: "max-age", Value: "4649"},
		},
		{
			name:    "max-age quoted rejected",
			source:  `max-age="4649"`,
			wantErr: true,
		},
		{
			name:    "no-store with argument rejected",
			source:  `no-store="foo"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := cachecontrol.ParseRequestDirective(tt.source)
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

func TestParseResponseDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		wantErr bool
		want    *cachecontrol.TokenPair
	}{
		{
			name:   "s-maxage token",
			source: "s-maxage=4649",
			want:   &cachecontrol.TokenPair{Name: "s-maxage", Value: "4649"},
		},
		{
			name:   "no-store flag",
			source: "no-store",
			want:   &cachecontrol.TokenPair{Name: "no-store"},
		},
		{
			name:   "extension with quoted value",
			source: `community="UCI"`,
			want:   &cachecontrol.TokenPair{Name: "community", Value: "UCI"},
		},
		{
			name:    "max-age quoted rejected",
			source:  `max-age="4649"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := cachecontrol.ParseResponseDirective(tt.source)
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

func TestParseRequestDirectives(t *testing.T) {
	t.Parallel()

	tokens, err := cachecontrol.ParseRequestDirectives(` max-age=4649 , no-store `)
	require.NoError(t, err)
	require.Len(t, tokens, 2)
	assert.Equal(t, &cachecontrol.TokenPair{Name: "max-age", Value: "4649"}, tokens[0])
	assert.Equal(t, &cachecontrol.TokenPair{Name: "no-store"}, tokens[1])
}

func TestParseResponseDirectives(t *testing.T) {
	t.Parallel()

	tokens, err := cachecontrol.ParseResponseDirectives(`max-age=4649, no-store, community="UCI"`)
	require.NoError(t, err)
	require.Len(t, tokens, 3)
	assert.Equal(t, &cachecontrol.TokenPair{Name: "max-age", Value: "4649"}, tokens[0])
	assert.Equal(t, &cachecontrol.TokenPair{Name: "no-store"}, tokens[1])
	assert.Equal(t, &cachecontrol.TokenPair{Name: "community", Value: "UCI"}, tokens[2])
}

func TestParseRequest(t *testing.T) {
	t.Parallel()

	t.Run(
		"max-age and no-store",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseRequest("max-age=4649, no-store")
			require.NoError(t, err)

			seconds, ok := dir.MaxAge()
			require.True(t, ok)
			assert.Equal(t, uint64(4649), seconds)
			assert.True(t, dir.NoStore())
		},
	)

	t.Run(
		"invalid max-age rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := cachecontrol.ParseRequest(`max-age="4649"`)
			require.Error(t, err)
		},
	)

	t.Run(
		"max-stale without value",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseRequest("max-stale")
			require.NoError(t, err)
			assert.True(t, dir.MaxStaleUnbounded())

			_, bounded, ok := dir.MaxStale()
			require.True(t, ok)
			assert.False(t, bounded)
		},
	)

	t.Run(
		"max-stale with value",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseRequest("max-stale=120")
			require.NoError(t, err)
			assert.False(t, dir.MaxStaleUnbounded())

			seconds, bounded, ok := dir.MaxStale()
			require.True(t, ok)
			assert.True(t, bounded)
			assert.Equal(t, uint64(120), seconds)
		},
	)
}

func TestParseResponse(t *testing.T) {
	t.Parallel()

	t.Run(
		"response directives and extension",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseResponse(`max-age=4649, no-store, community="UCI"`)
			require.NoError(t, err)

			seconds, ok := dir.MaxAge()
			require.True(t, ok)
			assert.Equal(t, uint64(4649), seconds)
			assert.True(t, dir.NoStore())
			assert.Equal(t, map[string]string{"community": "UCI"}, dir.Extensions())
		},
	)

	t.Run(
		"multiple max-age uses minimum",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseResponse("max-age=3600, max-age=60")
			require.NoError(t, err)

			seconds, ok := dir.MaxAge()
			require.True(t, ok)
			assert.Equal(t, uint64(60), seconds)
		},
	)

	t.Run(
		"s-maxage and flags",
		func(t *testing.T) {
			t.Parallel()

			dir, err := cachecontrol.ParseResponse("public, max-age=604800, s-maxage=86400, must-revalidate")
			require.NoError(t, err)

			maxAge, ok := dir.MaxAge()
			require.True(t, ok)
			assert.Equal(t, uint64(604800), maxAge)

			sMaxAge, ok := dir.SMaxAge()
			require.True(t, ok)
			assert.Equal(t, uint64(86400), sMaxAge)

			assert.True(t, dir.Public())
			assert.True(t, dir.MustRevalidate())
		},
	)

	t.Run(
		"invalid max-age rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := cachecontrol.ParseResponse(`max-age="4649"`)
			require.Error(t, err)
		},
	)
}

func TestResponseMaxAgeDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		header  string
		wantAge time.Duration
		wantOK  bool
	}{
		{
			name:   "empty header",
			header: "",
			wantOK: false,
		},
		{
			name:   "whitespace only",
			header: "   ",
			wantOK: false,
		},
		{
			name:   "no max-age directive",
			header: "public, private, no-cache",
			wantOK: false,
		},
		{
			name:    "single max-age",
			header:  "max-age=120",
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:    "max-age with other directives",
			header:  "public, max-age=120, private",
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:    "max-age zero",
			header:  "max-age=0",
			wantAge: 0,
			wantOK:  true,
		},
		{
			name:    "case insensitive directive name",
			header:  "Max-Age=90",
			wantAge: 90 * time.Second,
			wantOK:  true,
		},
		{
			name:    "whitespace around comma separators",
			header:  "public , max-age=120 , private",
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:    "whitespace around equals sign",
			header:  "max-age = 120",
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:    "multiple max-age uses minimum",
			header:  "max-age=3600, max-age=60",
			wantAge: 60 * time.Second,
			wantOK:  true,
		},
		{
			name:   "invalid max-age makes header invalid",
			header: "max-age=bad, max-age=30",
			wantOK: false,
		},
		{
			name:   "all max-age values invalid",
			header: "max-age=, max-age=abc",
			wantOK: false,
		},
		{
			name:   "negative max-age rejected",
			header: "max-age=-1",
			wantOK: false,
		},
		{
			name:   "decimal max-age rejected",
			header: "max-age=1.5",
			wantOK: false,
		},
		{
			name:   "quoted max-age rejected",
			header: `max-age="120"`,
			wantOK: false,
		},
		{
			name:    "leading zeros preserved",
			header:  "max-age=0060",
			wantAge: 60 * time.Second,
			wantOK:  true,
		},
		{
			name:   "s-maxage ignored by MaxAge helper",
			header: "s-maxage=3600",
			wantOK: false,
		},
		{
			name:    "s-maxage and max-age both present",
			header:  "s-maxage=3600, max-age=120",
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:   "directive name must match exactly",
			header: "foo-max-age=120",
			wantOK: false,
		},
		{
			name:    "comma inside quoted extension value",
			header:  `foo="bar,baz", max-age=120`,
			wantAge: 120 * time.Second,
			wantOK:  true,
		},
		{
			name:    "real world nginx style",
			header:  "max-age=31536000, public, immutable",
			wantAge: 365 * 24 * time.Hour,
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				dir, err := cachecontrol.ParseResponse(tt.header)
				if !tt.wantOK {
					if err == nil {
						_, gotOK := dir.MaxAgeDuration()
						assert.False(t, gotOK)
					}

					return
				}

				require.NoError(t, err)

				gotAge, gotOK := dir.MaxAgeDuration()
				assert.True(t, gotOK)
				assert.Equal(t, tt.wantAge, gotAge)
			},
		)
	}
}

func TestResponseMaxAgeDuration_Overflow(t *testing.T) {
	t.Parallel()

	dir, err := cachecontrol.ParseResponse("max-age=9223372036854775807")
	require.NoError(t, err)

	age, ok := dir.MaxAgeDuration()
	require.True(t, ok)
	assert.Equal(t, time.Duration(math.MaxInt64), age)
}

func TestParseResponseDirectives_NoSpaceAfterComma(t *testing.T) {
	t.Parallel()

	tokens, err := cachecontrol.ParseResponseDirectives("max-age=120,no-store")
	require.NoError(t, err)
	require.Len(t, tokens, 2)
}
