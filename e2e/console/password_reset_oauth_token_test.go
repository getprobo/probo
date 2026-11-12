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
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	// signupPassword is the password NewClient sets during sign-up.
	signupPassword = "TestPassword123!"

	oauthAPIScopes = "openid offline_access v1:org:read"

	organizationQuery = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}
	`
)

func TestResetPassword_RevokesOAuth2Tokens(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClient(t, testutil.RoleOwner)

	ownerOrgID := owner.GetOrganizationID().String()
	otherOrgID := other.GetOrganizationID().String()

	personalToken := mintPersonalAccessToken(t, owner)
	otherToken := mintPersonalAccessToken(t, other)
	grant := mintAuthorizationCodeTokens(t, owner)

	assertAccessTokenAllowsOrganization(t, personalToken, ownerOrgID)
	assertAccessTokenAllowsOrganization(t, grant.accessToken, ownerOrgID)
	assertAccessTokenAllowsOrganization(t, otherToken, otherOrgID)

	const newPassword = "ResetPassword123!"
	resetPasswordByEmail(t, owner.GetEmail(), newPassword)

	assertAccessTokenRejected(t, personalToken, ownerOrgID)
	assertAccessTokenRejected(t, grant.accessToken, ownerOrgID)
	assertRefreshTokenRejected(t, owner, grant)

	assertAccessTokenAllowsOrganization(t, otherToken, otherOrgID)

	require.Error(t, testutil.NewUnauthenticatedClient(t).SignIn(owner.GetEmail(), signupPassword))

	signedIn := testutil.NewUnauthenticatedClient(t)
	require.NoError(t, signedIn.SignIn(owner.GetEmail(), newPassword))
	assertPersonalAccessTokenCount(t, signedIn, 0)
}

func TestResetPassword_RevokesUnredeemedOAuth2Grants(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClient(t, testutil.RoleOwner)

	// Register clients before any code is issued. E2E authorization codes live
	// for 5 seconds and device codes for 15, so registration must not consume
	// that lifetime.
	ownerAuthClient := factory.CreateOAuth2ClientWithAPIScopes(owner, oauthAPIScopes, nil)
	otherAuthClient := factory.CreateOAuth2ClientWithAPIScopes(other, oauthAPIScopes, nil)
	ownerDeviceClient := factory.CreatePublicOAuth2Client(owner, nil)
	otherDeviceClient := factory.CreatePublicOAuth2Client(other, nil)

	ownerDevice := authorizeDeviceCode(t, owner, ownerDeviceClient)
	otherDevice := authorizeDeviceCode(t, other, otherDeviceClient)
	ownerCode := issueAuthorizationCode(t, owner, ownerAuthClient)
	otherCode := issueAuthorizationCode(t, other, otherAuthClient)

	resetPasswordByEmail(t, owner.GetEmail(), "ResetPassword123!")

	// The other identity's code was issued just after the owner's. A successful
	// exchange shows the owner's code is still inside its lifetime, so its
	// rejection is deletion. The token endpoint reports both cases as
	// "invalid or expired code".
	otherTokens, otherCodeRaw := exchangeAuthorizationCode(t, other, otherCode)
	require.Equal(t, http.StatusOK, otherCodeRaw.StatusCode)
	require.NotNil(t, otherTokens)
	require.NotEmpty(t, otherTokens.AccessToken)

	_, ownerCodeRaw := exchangeAuthorizationCode(t, owner, ownerCode)
	assertOAuth2GrantRejected(t, ownerCodeRaw, "invalid or expired code")

	otherDeviceTokens, otherDeviceErr, otherDeviceRaw, err := testutil.OAuth2TokenWithDeviceCode(
		other,
		otherDevice.client.ClientID,
		otherDevice.deviceCode,
	)
	require.NoError(t, err)
	require.Nil(t, otherDeviceErr)
	require.Equal(t, http.StatusOK, otherDeviceRaw.StatusCode)
	require.NotNil(t, otherDeviceTokens)
	require.NotEmpty(t, otherDeviceTokens.AccessToken)

	_, ownerDeviceErr, _, err := testutil.OAuth2TokenWithDeviceCode(
		owner,
		ownerDevice.client.ClientID,
		ownerDevice.deviceCode,
	)
	require.NoError(t, err)
	require.NotNil(t, ownerDeviceErr)
	assert.Equal(t, "invalid_grant", ownerDeviceErr.Code)
	assert.Equal(t, "invalid device code", ownerDeviceErr.Description)
}

func TestChangePassword_PreservesOAuth2Tokens(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()
	personalToken := mintPersonalAccessToken(t, owner)
	grant := mintAuthorizationCodeTokens(t, owner)

	const newPassword = "ChangedPassword123!"

	const changePasswordMutation = `
		mutation($input: ChangePasswordInput!) {
			changePassword(input: $input) {
				success
			}
		}
	`

	var changed struct {
		ChangePassword struct {
			Success bool `json:"success"`
		} `json:"changePassword"`
	}

	err := owner.ExecuteConnect(changePasswordMutation, map[string]any{
		"input": map[string]any{
			"currentPassword": signupPassword,
			"newPassword":     newPassword,
		},
	}, &changed)
	require.NoError(t, err)
	assert.True(t, changed.ChangePassword.Success)

	assertAccessTokenAllowsOrganization(t, personalToken, organizationID)
	assertAccessTokenAllowsOrganization(t, grant.accessToken, organizationID)
	assertRefreshTokenWorks(t, owner, grant, organizationID)
	assertPersonalAccessTokenCount(t, owner, 1)
}

type authorizationCodeTokens struct {
	client       factory.OAuth2ClientResult
	accessToken  string
	refreshToken string
}

func mintPersonalAccessToken(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	const createMutation = `
		mutation CreateOAuth2AccessToken($input: CreateOAuth2AccessTokenInput!) {
			createOAuth2AccessToken(input: $input) {
				token
			}
		}
	`

	expiresAt := time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339)
	resp, err := owner.DoConnect(createMutation, map[string]any{
		"input": map[string]any{
			"name":      "password credential test",
			"expiresAt": expiresAt,
			"scopes":    []string{"v1:org:read"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var created struct {
		CreateOAuth2AccessToken struct {
			Token string `json:"token"`
		} `json:"createOAuth2AccessToken"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &created))
	require.NotEmpty(t, created.CreateOAuth2AccessToken.Token)

	return created.CreateOAuth2AccessToken.Token
}

type pendingAuthorizationCode struct {
	client   factory.OAuth2ClientResult
	code     string
	verifier string
}

type authorizedDeviceCode struct {
	client     factory.OAuth2ClientResult
	deviceCode string
}

func issueAuthorizationCode(
	t *testing.T,
	owner *testutil.Client,
	client factory.OAuth2ClientResult,
) pendingAuthorizationCode {
	t.Helper()

	const redirectURI = "http://localhost:9999/callback"

	verifier, challenge := testutil.GeneratePKCE()

	params := url.Values{
		"client_id":             {client.ClientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {oauthAPIScopes},
		"state":                 {"test-state"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}

	authResp, err := testutil.OAuth2Authorize(owner, params)
	require.NoError(t, err)

	var code string

	if testutil.IsConsentRedirect(authResp) {
		consentID, err := testutil.ExtractConsentIDFromResponse(authResp)
		require.NoError(t, err)

		consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
		require.NoError(t, err)
		require.Equal(t, http.StatusFound, consentResp.StatusCode)

		code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
		require.NoError(t, err)
	} else {
		require.Equal(t, http.StatusFound, authResp.StatusCode)

		code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
		require.NoError(t, err)
	}

	return pendingAuthorizationCode{
		client:   client,
		code:     code,
		verifier: verifier,
	}
}

func exchangeAuthorizationCode(
	t *testing.T,
	owner *testutil.Client,
	pending pendingAuthorizationCode,
) (*testutil.OAuth2TokenResponse, *testutil.OAuth2HTTPResponse) {
	t.Helper()

	tokens, raw, err := testutil.OAuth2TokenWithCode(
		owner,
		pending.client.ClientID,
		pending.client.ClientSecret,
		pending.code,
		"http://localhost:9999/callback",
		pending.verifier,
	)
	require.NoError(t, err)
	require.NotNil(t, raw)

	return tokens, raw
}

func authorizeDeviceCode(
	t *testing.T,
	owner *testutil.Client,
	client factory.OAuth2ClientResult,
) authorizedDeviceCode {
	t.Helper()

	deviceResp, raw, err := testutil.OAuth2DeviceAuth(
		owner,
		client.ClientID,
		"openid email profile offline_access",
	)
	require.NoError(t, err)
	require.NotNil(t, raw)
	require.Equal(t, http.StatusOK, raw.StatusCode, "device auth failed: %s", string(raw.Body))
	require.NotNil(t, deviceResp)
	require.NotEmpty(t, deviceResp.DeviceCode)
	require.NotEmpty(t, deviceResp.UserCode)

	verifyResp, err := testutil.OAuth2DeviceVerify(owner, deviceResp.UserCode)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, verifyResp.StatusCode, "device verify failed: %s", string(verifyResp.Body))

	return authorizedDeviceCode{
		client:     client,
		deviceCode: deviceResp.DeviceCode,
	}
}

func assertOAuth2GrantRejected(t *testing.T, raw *testutil.OAuth2HTTPResponse, description string) {
	t.Helper()

	require.NotEqual(t, http.StatusOK, raw.StatusCode)

	var oauthErr testutil.OAuth2ErrorResponse
	require.NoError(t, json.Unmarshal(raw.Body, &oauthErr))
	assert.Equal(t, "invalid_grant", oauthErr.Code)
	assert.Equal(t, description, oauthErr.Description)
}

func mintAuthorizationCodeTokens(t *testing.T, owner *testutil.Client) authorizationCodeTokens {
	t.Helper()

	const redirectURI = "http://localhost:9999/callback"

	client := factory.CreateOAuth2ClientWithAPIScopes(owner, oauthAPIScopes, nil)
	tokens := testutil.OAuth2PerformAuthorizationCodeFlowWithScopes(
		t,
		owner,
		client.ClientID,
		client.ClientSecret,
		redirectURI,
		oauthAPIScopes,
	)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)

	return authorizationCodeTokens{
		client:       client,
		accessToken:  tokens.AccessToken,
		refreshToken: tokens.RefreshToken,
	}
}

func resetPasswordByEmail(t *testing.T, email string, newPassword string) {
	t.Helper()

	anonymous := testutil.NewUnauthenticatedClient(t)

	const forgotPasswordMutation = `
		mutation($input: ForgotPasswordInput!) {
			forgotPassword(input: $input) {
				success
			}
		}
	`

	err := anonymous.ExecuteConnect(forgotPasswordMutation, map[string]any{
		"input": map[string]any{
			"email": email,
		},
	}, nil)
	require.NoError(t, err)

	resetToken := anonymous.PollForLinkToken(`to:` + email + ` subject:"Reset your password"`)

	const resetPasswordMutation = `
		mutation($input: ResetPasswordInput!) {
			resetPassword(input: $input) {
				success
			}
		}
	`

	var reset struct {
		ResetPassword struct {
			Success bool `json:"success"`
		} `json:"resetPassword"`
	}

	err = anonymous.ExecuteConnect(resetPasswordMutation, map[string]any{
		"input": map[string]any{
			"token":    resetToken,
			"password": newPassword,
		},
	}, &reset)
	require.NoError(t, err)
	require.True(t, reset.ResetPassword.Success)
}

func assertAccessTokenAllowsOrganization(t *testing.T, accessToken string, organizationID string) {
	t.Helper()

	resp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		accessToken,
		organizationQuery,
		map[string]any{"id": organizationID},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func assertAccessTokenRejected(t *testing.T, accessToken string, organizationID string) {
	t.Helper()

	resp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		accessToken,
		organizationQuery,
		map[string]any{"id": organizationID},
	)
	require.Error(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Errors)
	assert.Equal(t, "UNAUTHENTICATED", resp.Errors[0].Code())
	assert.Empty(t, resp.DataString())
}

func assertRefreshTokenRejected(t *testing.T, owner *testutil.Client, grant authorizationCodeTokens) {
	t.Helper()

	_, raw, err := testutil.OAuth2TokenWithRefreshToken(
		owner,
		grant.client.ClientID,
		grant.client.ClientSecret,
		grant.refreshToken,
	)
	require.NoError(t, err)
	require.NotNil(t, raw)
	assert.NotEqual(t, http.StatusOK, raw.StatusCode)
}

func assertRefreshTokenWorks(
	t *testing.T,
	owner *testutil.Client,
	grant authorizationCodeTokens,
	organizationID string,
) {
	t.Helper()

	refreshed, raw, err := testutil.OAuth2TokenWithRefreshToken(
		owner,
		grant.client.ClientID,
		grant.client.ClientSecret,
		grant.refreshToken,
	)
	require.NoError(t, err)
	require.NotNil(t, raw)
	require.Equal(t, http.StatusOK, raw.StatusCode)
	require.NotNil(t, refreshed)
	require.NotEmpty(t, refreshed.AccessToken)

	assertAccessTokenAllowsOrganization(t, refreshed.AccessToken, organizationID)
}

func assertPersonalAccessTokenCount(t *testing.T, owner *testutil.Client, want int) {
	t.Helper()

	const listQuery = `
		query ListOAuth2AccessTokens {
			viewer {
				oauth2AccessTokens(first: 10) {
					totalCount
				}
			}
		}
	`

	resp, err := owner.DoConnect(listQuery, map[string]any{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var listed struct {
		Viewer struct {
			OAuth2AccessTokens struct {
				TotalCount int `json:"totalCount"`
			} `json:"oauth2AccessTokens"`
		} `json:"viewer"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &listed))
	assert.Equal(t, want, listed.Viewer.OAuth2AccessTokens.TotalCount)
}
