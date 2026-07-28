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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmailVerification_PasswordSignInRequiresVerifiedEmail(t *testing.T) {
	t.Parallel()

	client := testutil.NewUnauthenticatedClient(t)

	uniqueID := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("unverified-%s@e2e.probo.test", uniqueID)
	password := "TestPassword123!"
	fullName := fmt.Sprintf("Unverified User %s", uniqueID)

	const signUpMutation = `
		mutation($input: SignUpInput!) {
			signUp(input: $input) {
				identity { id }
			}
		}
	`

	var signUpResult struct {
		SignUp struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signUp"`
	}

	err := client.ExecuteConnect(signUpMutation, map[string]any{
		"input": map[string]any{
			"email":    email,
			"password": password,
			"fullName": fullName,
		},
	}, &signUpResult)
	require.NoError(t, err, "signUp should succeed for unverified identity")
	require.NotEmpty(t, signUpResult.SignUp.Identity.ID)

	client.SignOut()

	err = client.SignIn(email, password)
	testutil.RequireErrorCode(t, err, "EMAIL_NOT_VERIFIED")

	client.ResendVerificationEmail(email)

	token := client.GetEmailConfirmationToken(email)
	require.NotEmpty(t, token)

	const verifyMutation = `
		mutation($input: VerifyEmailInput!) {
			verifyEmail(input: $input) {
				success
			}
		}
	`

	var verifyResult struct {
		VerifyEmail struct {
			Success bool `json:"success"`
		} `json:"verifyEmail"`
	}

	err = client.ExecuteConnect(verifyMutation, map[string]any{
		"input": map[string]any{
			"token": token,
		},
	}, &verifyResult)
	require.NoError(t, err, "verifyEmail should succeed")
	assert.True(t, verifyResult.VerifyEmail.Success)

	err = client.SignIn(email, password)
	require.NoError(t, err, "signIn should succeed after email verification")
}

func TestEmailVerification_ResendIsEnumerationSafe(t *testing.T) {
	t.Parallel()

	client := testutil.NewUnauthenticatedClient(t)

	client.ResendVerificationEmail(fmt.Sprintf("missing-%d@e2e.probo.test", time.Now().UnixNano()))
}
