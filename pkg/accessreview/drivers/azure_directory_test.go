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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

func azureDirectorySession(transport http.RoundTripper) *cloudazure.Session {
	return cloudazure.NewSessionFromToken(
		"11111111-1111-4111-8111-111111111111",
		"token",
		cloudazure.WithHTTPClient(&http.Client{Transport: transport}),
	)
}

func azureGraphJSONOK(body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestAzureGetByIDsURL(t *testing.T) {
	t.Parallel()

	for _, env := range cloudazure.Environments() {
		t.Run(
			env.String(),
			func(t *testing.T) {
				t.Parallel()

				session := cloudazure.NewSessionFromToken(
					"11111111-1111-4111-8111-111111111111",
					"token",
					cloudazure.WithEnvironment(env),
				)
				got, err := azureGetByIDsURL(session.GraphBaseURL())
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(got, session.GraphBaseURL()))
				assert.True(t, strings.HasSuffix(got, "/v1.0/directoryObjects/getByIds"))
				assert.NotContains(t, got, "graph.microsoft.com/"+env.String())
			},
		)
	}
}

func TestAzureDirectoryObjectResolved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		object azureDirectoryObject
		want   bool
	}{
		{
			name:   "null-filled id only",
			object: azureDirectoryObject{ID: "22222222-2222-4222-8222-222222222222"},
		},
		{
			name: "resolved user",
			object: azureDirectoryObject{
				ID:             "22222222-2222-4222-8222-222222222222",
				DisplayName:    "Alice Chen",
				AccountEnabled: new(true),
			},
			want: true,
		},
		{
			name: "disabled service principal",
			object: azureDirectoryObject{
				ID:             "44444444-4444-4444-8444-444444444444",
				AccountEnabled: new(false),
			},
			want: true,
		},
		{name: "empty", object: azureDirectoryObject{}},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, azureDirectoryObjectResolved(tt.object))
			},
		)
	}
}

func TestAzureGraphResponseError_OmitsBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body: io.NopCloser(
			strings.NewReader(`{"error":{"code":"Authorization_RequestDenied","message":"alice@probo-azure.test"}}`),
		),
	}

	err := azureGraphResponseError(resp)
	require.Error(t, err)

	apiErr, ok := err.(*azcore.ResponseError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Equal(t, "Authorization_RequestDenied", apiErr.ErrorCode)
	assert.NotContains(t, err.Error(), "alice@probo-azure.test")
}

func TestResolveAzurePrincipals_IgnoresGUIDCasing(t *testing.T) {
	t.Parallel()

	session := cloudazure.NewSessionFromToken(
		"11111111-1111-4111-8111-111111111111",
		"token",
		cloudazure.WithHTTPClient(
			&http.Client{
				Transport: roundTripFunc(
					func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     http.Header{"Content-Type": []string{"application/json"}},
							Body: io.NopCloser(
								strings.NewReader(
									`{"value":[{"id":"AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA","displayName":"Alice Chen","mail":"alice@probo-azure.test","accountEnabled":true}]}`,
								),
							),
						}, nil
					},
				),
			},
		),
	)

	identities := []azureIdentity{
		{
			PrincipalID:   "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			PrincipalType: azurePrincipalUser,
		},
	}

	err := resolveAzurePrincipals(context.Background(), session, identities)
	require.NoError(t, err)
	assert.Equal(t, "Alice Chen", identities[0].DisplayName)
	assert.Equal(t, "alice@probo-azure.test", identities[0].Email)
	require.NotNil(t, identities[0].AccountEnabled)
	assert.True(t, *identities[0].AccountEnabled)
}

func TestAzureEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "alice@probo-azure.test", azureEmail(azurePrincipalUser, "alice@probo-azure.test", "upn@probo-azure.test"))
	assert.Equal(t, "upn@probo-azure.test", azureEmail(azurePrincipalUser, "", "upn@probo-azure.test"))
	assert.Equal(t, "eng@probo-azure.test", azureEmail(azurePrincipalGroup, "eng@probo-azure.test", ""))
	assert.Empty(t, azureEmail(azurePrincipalServicePrincipal, "sp@probo-azure.test", ""))
}

func TestGetAzureDirectoryObjectBatch_FollowsGetByIdsNextLink(t *testing.T) {
	t.Parallel()

	var second http.Request

	session := azureDirectorySession(
		roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				if req.URL.RawQuery != "" {
					second = *req

					return azureGraphJSONOK(
						`{"value":[{"id":"2","displayName":"Bob"}]}`,
					)
				}

				next := *req.URL
				q := next.Query()
				q.Set("$skiptoken", "abc")
				next.RawQuery = q.Encode()

				return azureGraphJSONOK(
					fmt.Sprintf(
						`{"value":[{"id":"1","displayName":"Alice"}],"@odata.nextLink":%q}`,
						next.String(),
					),
				)
			},
		),
	)

	objects, err := getAzureDirectoryObjectBatch(
		context.Background(),
		session,
		[]string{"1", "2"},
	)
	require.NoError(t, err)
	require.Len(t, objects, 2)
	assert.Equal(t, "Alice", objects[0].DisplayName)
	assert.Equal(t, "Bob", objects[1].DisplayName)
	assert.Equal(t, http.MethodGet, second.Method)
	assert.Equal(t, "/v1.0/directoryObjects/getByIds", second.URL.Path)
	assert.Equal(t, "abc", second.URL.Query().Get("$skiptoken"))
}

func TestGetAzureDirectoryObjectBatch_RefusesUnsafeNextLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		next      func(req *http.Request) string
		forbidden func(req *http.Request) bool
		wantErr   string
		echoHost  string
	}{
		{
			name: "same host outside getByIds",
			next: func(req *http.Request) string {
				next := *req.URL
				next.Path = "/v1.0/users"
				next.RawQuery = ""

				return next.String()
			},
			forbidden: func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, "/v1.0/users")
			},
			wantErr: "it leaves the collection path",
		},
		{
			name: "cross host",
			next: func(*http.Request) string {
				return (&url.URL{
					Scheme: "https",
					Host:   crossHostPaginationHost,
					Path:   "/v1.0/directoryObjects/getByIds",
				}).String()
			},
			forbidden: func(req *http.Request) bool {
				return strings.EqualFold(req.URL.Hostname(), crossHostPaginationHost)
			},
			wantErr:  "cross-host pagination is not allowed",
			echoHost: crossHostPaginationHost,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var forbiddenRequests int

				session := azureDirectorySession(
					roundTripFunc(
						func(req *http.Request) (*http.Response, error) {
							if tt.forbidden(req) {
								forbiddenRequests++
							}

							return azureGraphJSONOK(
								fmt.Sprintf(
									`{"value":[],"@odata.nextLink":%q}`,
									tt.next(req),
								),
							)
						},
					),
				)

				_, err := getAzureDirectoryObjectBatch(
					context.Background(),
					session,
					[]string{"1"},
				)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Zero(t, forbiddenRequests)

				if tt.echoHost != "" {
					assert.NotContains(t, err.Error(), tt.echoHost)
				}
			},
		)
	}
}
