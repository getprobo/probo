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

package azidentityx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientAssertionCredential_PostsRequestedScopeOnly(t *testing.T) {
	t.Parallel()

	const (
		tenantID  = "a1d8168a-79aa-4686-9f04-ab6cb4469fa5"
		clientID  = "124e037a-15d4-4d45-89ef-89b4c9c0a935"
		scope     = "https://management.azure.com/.default"
		assertion = "header.payload.signature"
	)

	var (
		posts        int
		mints        int
		gotURL       string
		gotMethod    string
		gotType      string
		gotRequestID string
		gotReturnID  string
		gotSKU       string
		gotOS        string
		gotCPU       string
		gotVersion   string
		gotForm      url.Values
	)

	credential := testClientAssertionCredential(
		t,
		cloud.AzurePublic.ActiveDirectoryAuthorityHost,
		tenantID,
		clientID,
		func(context.Context) (string, error) {
			mints++

			return assertion, nil
		},
		func(request *http.Request) (*http.Response, error) {
			posts++
			gotURL = request.URL.String()
			gotMethod = request.Method
			gotType = request.Header.Get("Content-Type")
			gotRequestID = request.Header.Get("client-request-id")
			gotReturnID = request.Header.Get("Return-Client-Request-Id")
			gotSKU = request.Header.Get("x-client-sku")
			gotOS = request.Header.Get("x-client-os")
			gotCPU = request.Header.Get("x-client-cpu")
			gotVersion = request.Header.Get("x-client-ver")

			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)

			gotForm, err = url.ParseQuery(string(body))
			require.NoError(t, err)

			return tokenHTTPResponse(`{"access_token":"arm-token","expires_in":3600,"token_type":"Bearer"}`), nil
		},
	)

	options := policy.TokenRequestOptions{Scopes: []string{scope}}

	first, err := credential.GetToken(context.Background(), options)
	require.NoError(t, err)
	assert.Equal(t, "arm-token", first.Token)

	second, err := credential.GetToken(context.Background(), options)
	require.NoError(t, err)
	assert.Equal(t, first.Token, second.Token)

	assert.Equal(t, 1, posts)
	assert.Equal(t, 1, mints)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, tokenContentType, gotType)

	_, err = uuid.Parse(gotRequestID)
	require.NoError(t, err)
	assert.Equal(t, "false", gotReturnID)
	assert.Empty(t, gotSKU)
	assert.Empty(t, gotOS)
	assert.Empty(t, gotCPU)
	assert.Empty(t, gotVersion)
	assert.Equal(
		t,
		"https://login.microsoftonline.com/"+tenantID+"/oauth2/v2.0/token",
		gotURL,
	)
	assert.Equal(t, "client_credentials", gotForm.Get("grant_type"))
	assert.Equal(t, clientID, gotForm.Get("client_id"))
	assert.Equal(t, scope, gotForm.Get("scope"))
	assert.Equal(t, clientAssertionType, gotForm.Get("client_assertion_type"))
	assert.Equal(t, assertion, gotForm.Get("client_assertion"))
	assert.Equal(t, "1", gotForm.Get("client_info"))
	assert.NotContains(t, gotForm.Get("scope"), "openid")
	assert.NotContains(t, gotForm.Get("scope"), "offline_access")
	assert.NotContains(t, gotForm.Get("scope"), "profile")
	assert.Len(t, gotForm, 6)
}

func TestClientAssertionCredential_RefreshesAtRefreshOn(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	now := fixed
	posts := 0

	credential := testClientAssertionCredential(
		t,
		cloud.AzurePublic.ActiveDirectoryAuthorityHost,
		"a1d8168a-79aa-4686-9f04-ab6cb4469fa5",
		"b2222222-2222-4222-8222-222222222222",
		func(context.Context) (string, error) {
			return "header.payload.signature", nil
		},
		func(*http.Request) (*http.Response, error) {
			posts++

			return tokenHTTPResponse(`{"access_token":"arm-token","expires_in":3600,"refresh_in":1800}`), nil
		},
	)
	credential.now = func() time.Time { return now }

	options := policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}}

	_, err := credential.GetToken(context.Background(), options)
	require.NoError(t, err)

	now = fixed.Add(30*time.Minute - time.Second)
	_, err = credential.GetToken(context.Background(), options)
	require.NoError(t, err)
	assert.Equal(t, 1, posts)

	now = fixed.Add(30 * time.Minute)
	_, err = credential.GetToken(context.Background(), options)
	require.NoError(t, err)
	assert.Equal(t, 2, posts)
}

func TestClientAssertionCredential_PrefersExpiresOn(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

	credential := testClientAssertionCredential(
		t,
		cloud.AzurePublic.ActiveDirectoryAuthorityHost,
		"a1d8168a-79aa-4686-9f04-ab6cb4469fa5",
		"b2222222-2222-4222-8222-222222222222",
		func(context.Context) (string, error) {
			return "header.payload.signature", nil
		},
		func(*http.Request) (*http.Response, error) {
			return tokenHTTPResponse(fmt.Sprintf(
				`{"access_token":"arm-token","expires_in":"3600","expires_on":%d,"refresh_in":1800}`,
				fixed.Add(2*time.Hour).Unix(),
			)), nil
		},
	)
	credential.now = func() time.Time { return fixed }

	token, err := credential.GetToken(
		context.Background(),
		policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}},
	)
	require.NoError(t, err)
	assert.Equal(t, fixed.Add(2*time.Hour).UTC(), token.ExpiresOn)
	assert.Equal(t, fixed.Add(30*time.Minute).UTC(), token.RefreshOn)
}

func TestClientAssertionCredential_GovernmentAuthority(t *testing.T) {
	t.Parallel()

	const tenantID = "a1d8168a-79aa-4686-9f04-ab6cb4469fa5"

	var gotURL string

	credential := testClientAssertionCredential(
		t,
		cloud.AzureGovernment.ActiveDirectoryAuthorityHost,
		tenantID,
		"b2222222-2222-4222-8222-222222222222",
		func(context.Context) (string, error) {
			return "header.payload.signature", nil
		},
		func(request *http.Request) (*http.Response, error) {
			gotURL = request.URL.String()

			return tokenHTTPResponse(`{"access_token":"gov-token","expires_in":3600}`), nil
		},
	)

	_, err := credential.GetToken(
		context.Background(),
		policy.TokenRequestOptions{Scopes: []string{"https://management.usgovcloudapi.net/.default"}},
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://login.microsoftonline.us/"+tenantID+"/oauth2/v2.0/token",
		gotURL,
	)
}

func TestClientAssertionCredential_RequiresScope(t *testing.T) {
	t.Parallel()

	credential := testClientAssertionCredential(
		t,
		cloud.AzurePublic.ActiveDirectoryAuthorityHost,
		"a1d8168a-79aa-4686-9f04-ab6cb4469fa5",
		"b2222222-2222-4222-8222-222222222222",
		func(context.Context) (string, error) {
			t.Fatal("assertion must not be minted without a scope")

			return "", nil
		},
		func(*http.Request) (*http.Response, error) {
			t.Fatal("token endpoint must not be called without a scope")

			return nil, nil
		},
	)

	_, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope is required")
}

type tokenRoundTripFunc func(*http.Request) (*http.Response, error)

func (f tokenRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func testClientAssertionCredential(
	t *testing.T,
	authority string,
	tenantID string,
	clientID string,
	assertion func(context.Context) (string, error),
	roundTrip tokenRoundTripFunc,
) *ClientAssertionCredential {
	t.Helper()

	credential, err := NewClientAssertionCredential(
		&http.Client{Transport: roundTrip},
		authority,
		tenantID,
		clientID,
		assertion,
	)
	require.NoError(t, err)

	return credential
}

func tokenHTTPResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
