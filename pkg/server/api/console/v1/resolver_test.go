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

package console_v1

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVercelCallbackTeamID pins the exact OAuth-callback query-parameter name
// Vercel uses. It is camelCase `teamId`; a regression to the snake_case
// `team_id` used by most other providers would silently drop the team on every
// Vercel connect and leave the source resolving nobody.
func TestVercelCallbackTeamID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		query url.Values
		want  string
	}{
		{
			name:  "camelCase teamId is read",
			query: url.Values{"teamId": {"team_abc123"}},
			want:  "team_abc123",
		},
		{
			name:  "snake_case team_id is ignored",
			query: url.Values{"team_id": {"team_abc123"}},
			want:  "",
		},
		{
			name:  "absent parameter yields empty",
			query: url.Values{},
			want:  "",
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, vercelCallbackTeamID(tc.query))
			},
		)
	}
}
