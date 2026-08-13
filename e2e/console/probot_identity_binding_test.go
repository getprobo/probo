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
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/baseurl"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func newProbotBindToken(
	t *testing.T,
	subject identitybinding.Subject,
) string {
	t.Helper()

	baseURL, err := baseurl.Parse("https://console.example.com")
	require.NoError(t, err)
	service := identitybinding.NewService(test.PGClient(t), baseURL)
	bindURL, err := service.BindURL(context.Background(), subject)
	require.NoError(t, err)
	parsed, err := url.Parse(bindURL)
	require.NoError(t, err)

	return parsed.Query().Get("token")
}

func TestProbotIdentityBinding_ConfirmWhileLoggedIn(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	externalTenantID := "T-e2e-" + factory.SafeName("team")
	externalUserID := "U-e2e-" + factory.SafeName("user")
	token := newProbotBindToken(t, identitybinding.Subject{
		Provider:           slackchannel.ProviderName,
		ExternalTenantID:   externalTenantID,
		ExternalUserID:     externalUserID,
		ExternalTenantName: "acme-workspace",
		ExternalUserName:   "ada",
	})

	const previewQuery = `
		query($token: String!) {
			probotIdentityBindPreview(token: $token) {
				provider
				externalTenantId
				externalUserId
				externalTenantName
				externalUserName
			}
		}
	`

	var previewResult struct {
		ProbotIdentityBindPreview struct {
			Provider           string `json:"provider"`
			ExternalTenantID   string `json:"externalTenantId"`
			ExternalUserID     string `json:"externalUserId"`
			ExternalTenantName string `json:"externalTenantName"`
			ExternalUserName   string `json:"externalUserName"`
		} `json:"probotIdentityBindPreview"`
	}

	err := owner.Execute(
		previewQuery,
		map[string]any{"token": token},
		&previewResult,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		slackchannel.ProviderName,
		previewResult.ProbotIdentityBindPreview.Provider,
	)
	assert.Equal(
		t,
		externalTenantID,
		previewResult.ProbotIdentityBindPreview.ExternalTenantID,
	)
	assert.Equal(
		t,
		externalUserID,
		previewResult.ProbotIdentityBindPreview.ExternalUserID,
	)
	assert.Equal(
		t,
		"acme-workspace",
		previewResult.ProbotIdentityBindPreview.ExternalTenantName,
	)
	assert.Equal(
		t,
		"ada",
		previewResult.ProbotIdentityBindPreview.ExternalUserName,
	)

	const confirmMutation = `
		mutation($input: ConfirmProbotIdentityBindingInput!) {
			confirmProbotIdentityBinding(input: $input) {
				probotIdentityBinding {
					id
					provider
					externalTenantId
					externalUserId
				}
			}
		}
	`

	var confirmResult struct {
		ConfirmProbotIdentityBinding struct {
			ProbotIdentityBinding struct {
				ID               string `json:"id"`
				Provider         string `json:"provider"`
				ExternalTenantID string `json:"externalTenantId"`
				ExternalUserID   string `json:"externalUserId"`
			} `json:"probotIdentityBinding"`
		} `json:"confirmProbotIdentityBinding"`
	}

	err = owner.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		externalTenantID,
		confirmResult.ConfirmProbotIdentityBinding.ProbotIdentityBinding.ExternalTenantID,
	)
	assert.Equal(
		t,
		externalUserID,
		confirmResult.ConfirmProbotIdentityBinding.ProbotIdentityBinding.ExternalUserID,
	)
	assert.NotEmpty(
		t,
		confirmResult.ConfirmProbotIdentityBinding.ProbotIdentityBinding.ID,
	)

	err = owner.Execute(
		previewQuery,
		map[string]any{"token": token},
		&previewResult,
	)
	testutil.RequireErrorCode(t, err, "INVALID")

	err = owner.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")

	err = other.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")
}
