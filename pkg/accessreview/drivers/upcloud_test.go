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
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

func TestUpCloudDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/upcloud", "UPCLOUD_API_KEY")
	// UpCloud authenticates with a Bearer API token. The matcher ignores
	// Authorization, so replay needs no auth.
	client := newVCRClient(rec, bearerAuth(os.Getenv("UPCLOUD_API_KEY")))

	driver := NewUpCloudDriver(client, log.NewLogger(log.WithName("test")))
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	main := records[0]
	assert.Equal(t, "test", main.ExternalID)
	assert.Equal(t, "Main Account", main.FullName)
	assert.Equal(t, "main@example.com", main.Email)
	assert.Equal(t, []string{"technical"}, main.Roles)
	assert.True(t, main.IsAdmin)
	assert.Nil(t, main.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, main.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, main.AccountType)

	sub := records[1]
	assert.Equal(t, "my_sub_account", sub.ExternalID)
	assert.Equal(t, "Sub Account", sub.FullName)
	assert.Equal(t, "sub@example.com", sub.Email)
	assert.Equal(t, []string{"technical"}, sub.Roles)
	assert.False(t, sub.IsAdmin)

	// no roles assigned; details fetch fails (404), so the record falls back
	// to list-only fields rather than being dropped.
	temp := records[2]
	assert.Equal(t, "my_temp_account", temp.ExternalID)
	assert.Equal(t, "my_temp_account", temp.FullName)
	assert.Empty(t, temp.Email)
	assert.Equal(t, []string{}, temp.Roles)
	assert.False(t, temp.IsAdmin)

	billing := records[3]
	assert.Equal(t, "my_billing_account", billing.ExternalID)
	assert.Equal(t, "Billing Account", billing.FullName)
	assert.Equal(t, "billing@example.com", billing.Email)
	assert.Equal(t, []string{"billing"}, billing.Roles)
	assert.False(t, billing.IsAdmin)
}
