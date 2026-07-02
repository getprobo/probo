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
