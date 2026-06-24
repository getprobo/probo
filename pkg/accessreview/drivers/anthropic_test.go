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

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/anthropic", "ANTHROPIC_ADMIN_TOKEN")
	// Anthropic authenticates via x-api-key, not Authorization: Bearer.
	client := newVCRClientWithHeader(rec, "x-api-key", os.Getenv("ANTHROPIC_ADMIN_TOKEN"))

	driver := NewAnthropicDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Len(t, records, 3)

	first := records[0]
	assert.NotEmpty(t, first.Email)
	assert.NotEmpty(t, first.FullName)
	assert.NotEmpty(t, first.ExternalID)
	assert.Equal(t, []string{"User"}, first.Roles)
	assert.False(t, first.IsAdmin)
	assert.NotNil(t, first.CreatedAt)

	assert.Equal(t, []string{"Developer"}, records[1].Roles)
	assert.False(t, records[1].IsAdmin)

	admin := records[2]
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
}

func TestAnthropicRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want []string
	}{
		{"admin", []string{"Admin"}},
		{"billing", []string{"Billing"}},
		{"developer", []string{"Developer"}},
		{"claude_code_user", []string{"Claude Code User"}},
		{"user", []string{"User"}},
		{"unknown_future_role", []string{"unknown_future_role"}},
		{"", []string{}},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, anthropicRoles(c.in))
		})
	}
}
