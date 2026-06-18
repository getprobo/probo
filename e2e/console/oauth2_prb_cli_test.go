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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/cli/config"
)

func TestOAuth2_PrbCLIDeviceFlowWithAPIScopes(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	deviceResp, raw, err := testutil.OAuth2DeviceAuth(
		owner,
		config.CLIClientID,
		config.CLIClientScopes,
	)
	require.NoError(t, err)
	require.Equal(t, 200, raw.StatusCode, "device auth body: %s", string(raw.Body))
	require.NotNil(t, deviceResp)

	verifyResp, err := testutil.OAuth2DeviceVerify(owner, deviceResp.UserCode)
	require.NoError(t, err)
	require.Equal(t, 200, verifyResp.StatusCode)

	time.Sleep(time.Duration(deviceResp.Interval+1) * time.Second)

	tokenResp, _, pollRaw, err := testutil.OAuth2TokenWithDeviceCode(
		owner,
		config.CLIClientID,
		deviceResp.DeviceCode,
	)
	require.NoError(t, err)
	require.Equal(t, 200, pollRaw.StatusCode, "token poll body: %s", string(pollRaw.Body))
	require.NotNil(t, tokenResp)
	require.NotEmpty(t, tokenResp.AccessToken)
	assert.Contains(t, tokenResp.Scope, "v1:org:read")

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
}
