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
	assert.Contains(t, tokenResp.Scope, "v1:org")

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
