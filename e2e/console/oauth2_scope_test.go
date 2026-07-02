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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestOAuth2_ScopeEnforcementOnConsoleGraphQL(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	factory.CreateThirdParty(owner, factory.Attrs{"name": "Scoped OAuth Vendor"})

	const redirectURI = "http://localhost:9999/callback"

	client := factory.CreateOAuth2ClientWithAPIScopes(
		owner,
		"openid v1:org:read",
		nil,
	)

	tokenResp := testutil.OAuth2PerformAuthorizationCodeFlowWithScopes(
		t,
		owner,
		client.ClientID,
		client.ClientSecret,
		redirectURI,
		"openid v1:org:read",
	)
	require.NotEmpty(t, tokenResp.AccessToken)

	const getOrganizationQuery = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}
	`

	allowedResp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		tokenResp.AccessToken,
		getOrganizationQuery,
		map[string]any{
			"id": owner.GetOrganizationID().String(),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, allowedResp)

	const listThirdPartiesQuery = `
		query ListThirdParties($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					thirdParties(first: 10) {
						totalCount
					}
				}
			}
		}
	`

	deniedResp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		tokenResp.AccessToken,
		listThirdPartiesQuery,
		map[string]any{
			"orgId": owner.GetOrganizationID().String(),
		},
	)
	require.Error(t, err)
	require.NotNil(t, deniedResp)
	require.NotEmpty(t, deniedResp.Errors)

	code := deniedResp.Errors[0].Code()
	msg := deniedResp.Errors[0].Message
	isForbidden := code == "FORBIDDEN" ||
		(code == "" && (strings.Contains(msg, "does not have sufficient permissions") || strings.Contains(msg, "insufficient permissions")))
	require.True(t, isForbidden, "expected FORBIDDEN error, got code=%q message=%q", code, msg)
	assert.Empty(t, deniedResp.DataString(), "expected no data on denied request")
}
