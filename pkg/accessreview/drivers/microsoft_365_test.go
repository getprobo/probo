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
	assert.Equal(t, []string{"Global Administrator"}, recordsByEmail["enabled@example.com"].Roles)
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
