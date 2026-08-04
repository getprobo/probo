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

	"go.probo.inc/probo/e2e/internal/journey"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmailVerification_PasswordSignInRequiresVerifiedEmail(t *testing.T) {
	t.Parallel()

	world := journey.New(t)
	newUser := world.NewUnauthenticatedActor("new user")
	client := newUser.Client()

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

	newUser.Step("signs up without verifying their email", func() error {
		err := client.ExecuteConnect(signUpMutation, map[string]any{
			"input": map[string]any{
				"email":    email,
				"password": password,
				"fullName": fullName,
			},
		}, &signUpResult)
		if err != nil {
			return fmt.Errorf("cannot sign up: %w", err)
		}
		if signUpResult.SignUp.Identity.ID == "" {
			return fmt.Errorf("sign up returned an empty identity ID")
		}

		return nil
	})

	newUser.Step("signs out of the signup session", func() error {
		client.SignOut()
		return nil
	})

	newUser.Step("cannot sign in before verifying their email", func() error {
		err := client.SignIn(email, password)
		testutil.RequireErrorCode(t, err, "EMAIL_NOT_VERIFIED")
		return nil
	})

	newUser.Step("requests another verification email", func() error {
		client.ResendVerificationEmail(email)
		return nil
	})

	var token string
	newUser.Step("receives the verification email", func() error {
		token = client.GetEmailConfirmationToken(email)
		if token == "" {
			return fmt.Errorf("verification email contained an empty token")
		}

		return nil
	})

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

	newUser.Step("verifies their email address", func() error {
		err := client.ExecuteConnect(verifyMutation, map[string]any{
			"input": map[string]any{
				"token": token,
			},
		}, &verifyResult)
		if err != nil {
			return fmt.Errorf("cannot verify email: %w", err)
		}
		if !verifyResult.VerifyEmail.Success {
			return fmt.Errorf("verify email returned success=false")
		}

		return nil
	})

	newUser.Step("signs in with their verified email", func() error {
		if err := client.SignIn(email, password); err != nil {
			return fmt.Errorf("cannot sign in after email verification: %w", err)
		}

		return nil
	})
}

func TestEmailVerification_ResendIsEnumerationSafe(t *testing.T) {
	t.Parallel()

	client := testutil.NewUnauthenticatedClient(t)

	client.ResendVerificationEmail(fmt.Sprintf("missing-%d@e2e.probo.test", time.Now().UnixNano()))
}
