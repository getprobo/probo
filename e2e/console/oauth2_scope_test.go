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
