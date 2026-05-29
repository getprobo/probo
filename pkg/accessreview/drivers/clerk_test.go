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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestClerkDriver(t *testing.T) {
	t.Parallel()

	const responseBody = `{"data":[{"id":"usr_000000000000000000000001","primary_email_address_id":"eml_primary_1","username":null,"first_name":"Jane","last_name":"Doe","password_enabled":true,"two_factor_enabled":false,"totp_enabled":false,"backup_code_enabled":false,"banned":false,"locked":false,"last_sign_in_at":1748471521000,"created_at":1748342000000,"email_addresses":[{"id":"eml_secondary_1","email_address":"jane+alt@example.com"},{"id":"eml_primary_1","email_address":"jane@example.com"}]},{"id":"usr_000000000000000000000002","primary_email_address_id":null,"username":"developer-user","first_name":null,"last_name":null,"password_enabled":false,"two_factor_enabled":true,"totp_enabled":false,"backup_code_enabled":false,"banned":false,"locked":false,"last_sign_in_at":null,"created_at":1748343000000,"email_addresses":[{"id":"eml_2","email_address":"developer@example.com"}]},{"id":"usr_000000000000000000000003","primary_email_address_id":"eml_primary_3","username":null,"first_name":null,"last_name":null,"password_enabled":false,"two_factor_enabled":false,"totp_enabled":false,"backup_code_enabled":false,"banned":true,"locked":false,"last_sign_in_at":null,"created_at":1748344000000,"email_addresses":[{"id":"eml_primary_3","email_address":"blocked@example.com"}]}],"total_count":3}`

	requestCount := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		require.Equal(t, http.MethodGet, req.Method)
		require.Equal(t, "https://api.clerk.com/v1/users?limit=100&offset=0", req.URL.String())
		require.Equal(t, "application/json", req.Header.Get("Accept"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})}

	driver := NewClerkDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)
	require.Equal(t, 1, requestCount)

	first := records[0]
	assert.Equal(t, "usr_000000000000000000000001", first.ExternalID)
	assert.Equal(t, "jane@example.com", first.Email)
	assert.Equal(t, "Jane Doe", first.FullName)
	require.NotNil(t, first.Active)
	assert.True(t, *first.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, first.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodPassword, first.AuthMethod)
	assert.NotNil(t, first.CreatedAt)
	assert.NotNil(t, first.LastLogin)

	second := records[1]
	assert.Equal(t, "developer-user", second.FullName)
	require.NotNil(t, second.Active)
	assert.True(t, *second.Active)
	assert.Equal(t, coredata.MFAStatusEnabled, second.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodUnknown, second.AuthMethod)

	third := records[2]
	assert.Equal(t, "blocked@example.com", third.FullName)
	require.NotNil(t, third.Active)
	assert.False(t, *third.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, third.MFAStatus)
}

func TestClerkPrimaryEmail(t *testing.T) {
	t.Parallel()

	user := clerkUser{
		PrimaryEmailAddressID: new("eml_primary"),
		EmailAddresses: []struct {
			ID           string `json:"id"`
			EmailAddress string `json:"email_address"`
		}{
			{ID: "eml_secondary", EmailAddress: "secondary@example.com"},
			{ID: "eml_primary", EmailAddress: "primary@example.com"},
		},
	}

	assert.Equal(t, "primary@example.com", clerkPrimaryEmail(user))
}
