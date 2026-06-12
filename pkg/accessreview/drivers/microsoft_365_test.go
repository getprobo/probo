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

func TestMicrosoft365DriverMFAStatus(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: microsoft365RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/v1.0/directoryRoles":
					return microsoft365Response(
						http.StatusOK,
						`{"value":[{"id":"role-global","displayName":"Global Administrator"}]}`,
					), nil
				case "/v1.0/directoryRoles/role-global/members":
					return microsoft365Response(
						http.StatusOK,
						`{"value":[{"id":"user-enabled","@odata.type":"#microsoft.graph.user"}]}`,
					), nil
				case "/v1.0/users":
					assert.Equal(t, "userType eq 'Member'", req.URL.Query().Get("$filter"))

					return microsoft365Response(
						http.StatusOK,
						`{"value":[{"id":"user-enabled","userPrincipalName":"enabled@example.com","mail":"enabled@example.com","displayName":"Enabled User","accountEnabled":true},{"id":"user-disabled","userPrincipalName":"disabled@example.com","mail":"disabled@example.com","displayName":"Disabled User","accountEnabled":true},{"id":"user-fallback","userPrincipalName":"fallback@example.com","mail":"fallback@example.com","displayName":"Fallback User","accountEnabled":true},{"id":"user-missing","userPrincipalName":"missing@example.com","mail":"missing@example.com","displayName":"Missing User","accountEnabled":true}]}`,
					), nil
				case "/v1.0/reports/authenticationMethods/userRegistrationDetails":
					assert.Empty(t, req.URL.RawQuery)

					return microsoft365Response(
						http.StatusOK,
						`{"value":[{"id":"user-enabled","userPrincipalName":"enabled@example.com","isMfaRegistered":true},{"id":"user-disabled","userPrincipalName":"disabled@example.com","isMfaRegistered":false},{"id":"different-id","userPrincipalName":"FALLBACK@example.com","isMfaRegistered":true}]}`,
					), nil
				default:
					t.Fatalf("unexpected Microsoft Graph request: %s", req.URL.String())
					return nil, nil
				}
			},
		),
	}

	driver := NewMicrosoft365Driver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	recordsByEmail := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		recordsByEmail[record.Email] = record
	}

	assert.Equal(t, coredata.MFAStatusEnabled, recordsByEmail["enabled@example.com"].MFAStatus)
	assert.True(t, recordsByEmail["enabled@example.com"].IsAdmin)
	assert.Equal(t, "Global Administrator", recordsByEmail["enabled@example.com"].Role)
	assert.Equal(t, coredata.MFAStatusDisabled, recordsByEmail["disabled@example.com"].MFAStatus)
	assert.Equal(t, coredata.MFAStatusEnabled, recordsByEmail["fallback@example.com"].MFAStatus)
	assert.Equal(t, coredata.MFAStatusUnknown, recordsByEmail["missing@example.com"].MFAStatus)
}

type microsoft365RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f microsoft365RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func microsoft365Response(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
