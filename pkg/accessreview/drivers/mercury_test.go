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

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestMercuryDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/mercury", "MERCURY_API_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("MERCURY_API_TOKEN")))

	driver := NewMercuryDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	admin := records[0]
	assert.Equal(t, "8f1a6f1e-0000-4000-8000-000000000001", admin.ExternalID)
	assert.Equal(t, "ada@example.com", admin.Email)
	assert.Equal(t, "Ada Admin", admin.FullName)
	assert.Equal(t, []string{"Administrator"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)

	bookkeeper := records[1]
	assert.Equal(t, []string{"Bookkeeper"}, bookkeeper.Roles)
	assert.False(t, bookkeeper.IsAdmin)

	employee := records[2]
	assert.Equal(t, []string{"Employee"}, employee.Roles)
	assert.False(t, employee.IsAdmin)
}

func TestMercuryRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want []string
	}{
		{"administrator", []string{"Administrator"}},
		{"bookkeeper", []string{"Bookkeeper"}},
		{"customUser", []string{"Custom User"}},
		{"cardOnlyUser", []string{"Card Only User"}},
		{"employee", []string{"Employee"}},
		{"unknown_future_role", []string{"unknown_future_role"}},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, mercuryRoles(c.in))
		})
	}
}
