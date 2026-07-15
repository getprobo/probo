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

package trust_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTrustCenter_CIMDMetadataDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	trustCenterID := lookupTrustCenterID(t, owner)
	trustHost := lookupTrustHost(t, owner, trustCenterID)
	testutil.WaitForTrustCenterHTTPS(t, trustHost)

	client := testutil.TrustHTTPClient(trustHost)
	resp, err := client.Get("https://" + trustHost + "/.well-known/oauth-client-metadata")
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var doc struct {
		ClientID                string   `json:"client_id"`
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
	}
	require.NoError(t, json.Unmarshal(body, &doc))

	assert.Equal(
		t,
		"https://"+trustHost+"/.well-known/oauth-client-metadata",
		doc.ClientID,
	)
	assert.Equal(
		t,
		[]string{"https://" + trustHost + "/callback"},
		doc.RedirectURIs,
	)
	assert.Equal(t, "none", doc.TokenEndpointAuthMethod)
	assert.Contains(t, doc.GrantTypes, "authorization_code")
	assert.Equal(t, []string{"code"}, doc.ResponseTypes)
	assert.NotEmpty(t, doc.ClientName)
}

func TestTrustCenter_VisitorConnectViaCIMD(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	trustCenterID := lookupTrustCenterID(t, owner)
	trustHost := lookupTrustHost(t, owner, trustCenterID)

	visitor := testutil.SelfProvisionTrustCenterVisitor(t, trustHost)

	const query = `
		query {
			currentTrustCenter {
				title
			}
		}
	`

	var result struct {
		CurrentTrustCenter struct {
			Title string `json:"title"`
		} `json:"currentTrustCenter"`
	}

	err := visitor.ExecuteTrust(trustHost, query, nil, &result)
	require.NoError(t, err, "visitor session must authenticate trust GraphQL after CIMD connect")
	assert.NotEmpty(t, result.CurrentTrustCenter.Title)
}

func TestTrustCenter_UnknownCIMDClientRejected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	_, challenge := testutil.GeneratePKCE()

	clientID := "https://unknown-cimd.example.com/.well-known/oauth-client-metadata"

	resp, err := testutil.OAuth2Authorize(owner, url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"https://unknown-cimd.example.com/callback"},
		"response_type":         {"code"},
		"scope":                 {"openid profile email"},
		"state":                 {"rejected"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"nonce":                 {"test-nonce"},
	})
	require.NoError(t, err)

	require.False(
		t,
		testutil.IsConsentRedirect(resp),
		"unknown CIMD client must not reach consent screen",
	)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var oauthErr struct {
		Code string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(resp.Body, &oauthErr))
	assert.Equal(t, "invalid_client", oauthErr.Code)
}
