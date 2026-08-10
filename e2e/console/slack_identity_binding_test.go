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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/slackbot"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const e2eAuthCookieSecret = "this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes"

func newSlackBindToken(t *testing.T, teamID, slackUserID string) string {
	t.Helper()

	token, err := statelesstoken.NewToken(
		e2eAuthCookieSecret,
		slackbot.TokenTypeSlackIdentityBind,
		30*time.Minute,
		slackbot.BindTokenData{
			TeamID:      teamID,
			SlackUserID: slackUserID,
		},
	)
	require.NoError(t, err)

	return token
}

func TestSlackIdentityBinding_ConfirmWhileLoggedIn(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	teamID := "T-e2e-" + factory.SafeName("team")
	slackUserID := "U-e2e-" + factory.SafeName("user")
	token := newSlackBindToken(t, teamID, slackUserID)

	const previewQuery = `
		query($token: String!) {
			slackIdentityBindPreview(token: $token) {
				teamId
				slackUserId
			}
		}
	`

	var previewResult struct {
		SlackIdentityBindPreview struct {
			TeamID      string `json:"teamId"`
			SlackUserID string `json:"slackUserId"`
		} `json:"slackIdentityBindPreview"`
	}

	err := owner.Execute(previewQuery, map[string]any{"token": token}, &previewResult)
	require.NoError(t, err)
	assert.Equal(t, teamID, previewResult.SlackIdentityBindPreview.TeamID)
	assert.Equal(t, slackUserID, previewResult.SlackIdentityBindPreview.SlackUserID)

	const confirmMutation = `
		mutation($input: ConfirmSlackIdentityBindingInput!) {
			confirmSlackIdentityBinding(input: $input) {
				slackIdentityBinding {
					id
					teamId
					slackUserId
				}
			}
		}
	`

	var confirmResult struct {
		ConfirmSlackIdentityBinding struct {
			SlackIdentityBinding struct {
				ID          string `json:"id"`
				TeamID      string `json:"teamId"`
				SlackUserID string `json:"slackUserId"`
			} `json:"slackIdentityBinding"`
		} `json:"confirmSlackIdentityBinding"`
	}

	err = owner.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	require.NoError(t, err)
	assert.Equal(t, teamID, confirmResult.ConfirmSlackIdentityBinding.SlackIdentityBinding.TeamID)
	assert.Equal(t, slackUserID, confirmResult.ConfirmSlackIdentityBinding.SlackIdentityBinding.SlackUserID)
	assert.NotEmpty(t, confirmResult.ConfirmSlackIdentityBinding.SlackIdentityBinding.ID)

	err = other.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")
}
