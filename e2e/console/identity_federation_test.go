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

package console_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type identityFederationDiscovery struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

type identityFederationJWKS struct {
	Keys []struct {
		KeyType   string `json:"kty"`
		Use       string `json:"use"`
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
		N         string `json:"n"`
		E         string `json:"e"`
	} `json:"keys"`
}

func identityFederationGet(t testing.TB, client *testutil.Client, path string) (int, []byte) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, client.BaseURL()+path, nil)
	require.NoError(t, err)

	response, err := client.HTTPClient().Do(request)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, body
}

func identityFederationIssuerPath(organizationID string) string {
	return "/federation/" + organizationID
}

func TestIdentityFederation_Discovery(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	status, body := identityFederationGet(
		t,
		owner,
		identityFederationIssuerPath(organizationID)+"/.well-known/openid-configuration",
	)

	require.Equal(t, http.StatusOK, status, string(body))

	discovery := identityFederationDiscovery{}
	require.NoError(t, json.Unmarshal(body, &discovery))

	// The e2e harness sets no issuer base URL, so the issuer is derived from the
	// application base URL.
	wantIssuer := owner.BaseURL() + identityFederationIssuerPath(organizationID)

	assert.Equal(t, wantIssuer, discovery.Issuer)
	assert.Equal(t, wantIssuer+"/jwks", discovery.JWKSURI)
	assert.Equal(t, []string{"id_token"}, discovery.ResponseTypesSupported)
	assert.Equal(t, []string{"public"}, discovery.SubjectTypesSupported)
	assert.Equal(t, []string{"RS256"}, discovery.IDTokenSigningAlgValuesSupported)
}

func TestIdentityFederation_JWKS(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	status, body := identityFederationGet(
		t,
		owner,
		identityFederationIssuerPath(owner.GetOrganizationID().String())+"/jwks",
	)

	require.Equal(t, http.StatusOK, status, string(body))

	jwks := identityFederationJWKS{}
	require.NoError(t, json.Unmarshal(body, &jwks))

	require.NotEmpty(t, jwks.Keys)

	for _, key := range jwks.Keys {
		assert.Equal(t, "RSA", key.KeyType)
		assert.Equal(t, "sig", key.Use)
		assert.Equal(t, "RS256", key.Algorithm)
		assert.NotEmpty(t, key.KeyID)
		assert.NotEmpty(t, key.N)
		assert.NotEmpty(t, key.E)
	}
}

func TestIdentityFederation_UnknownOrganizationIsNotFound(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	unknownOrganizationID := gid.New(gid.TenantID{}, coredata.OrganizationEntityType)

	status, _ := identityFederationGet(
		t,
		owner,
		identityFederationIssuerPath(unknownOrganizationID.String())+"/.well-known/openid-configuration",
	)

	assert.Equal(t, http.StatusNotFound, status)
}

func TestIdentityFederation_NoTokenEndpointOrRFC8414Variant(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	// Only the appended OIDC discovery form exists. The RFC 8414 variant the
	// OAuth2 server serves has no identity federation equivalent.
	status, _ := identityFederationGet(
		t,
		owner,
		"/federation/.well-known/oauth-authorization-server/"+organizationID,
	)
	assert.Equal(t, http.StatusNotFound, status)

	// Tokens are minted in-process, so no HTTP surface hands one out.
	status, _ = identityFederationGet(t, owner, identityFederationIssuerPath(organizationID)+"/token")
	assert.Equal(t, http.StatusNotFound, status)
}
