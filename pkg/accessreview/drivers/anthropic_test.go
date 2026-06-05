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
	assert.Equal(t, "User", first.Role)
	assert.False(t, first.IsAdmin)
	assert.NotNil(t, first.CreatedAt)

	assert.Equal(t, "Developer", records[1].Role)
	assert.False(t, records[1].IsAdmin)

	admin := records[2]
	assert.Equal(t, "Admin", admin.Role)
	assert.True(t, admin.IsAdmin)
}

func TestAnthropicRole(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"admin", "Admin"},
		{"billing", "Billing"},
		{"developer", "Developer"},
		{"claude_code_user", "Claude Code User"},
		{"user", "User"},
		{"unknown_future_role", "unknown_future_role"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, anthropicRole(c.in))
		})
	}
}
