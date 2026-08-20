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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestCalendlyDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/calendly", "CALENDLY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("CALENDLY_TOKEN")))

	driver := NewCalendlyDriver(client, "https://api.calendly.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	assert.Equal(t, "owner@example.com", records[0].Email)
	assert.Equal(t, "Owner User", records[0].FullName)
	assert.Equal(t, []string{"Owner"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Nil(t, records[0].Active)
	assert.Equal(t, "https://api.calendly.com/users/U1", records[0].ExternalID)
	assert.Equal(t, new(createdAt), records[0].CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)

	assert.Equal(t, []string{"User"}, records[1].Roles)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Nil(t, records[1].CreatedAt)
}

func TestCalendlyDriver_ListAccountsWithoutOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resource":{"current_organization":""}}`))
	}))
	t.Cleanup(server.Close)

	driver := NewCalendlyDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())

	require.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "authenticated user has no organization")
}
