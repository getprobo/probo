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

const (
	signUpMutation = `
		mutation($input: SignUpInput!) {
			signUp(input: $input) {
				identity { id }
			}
		}
	`

	verifyEmailMutation = `
		mutation($input: VerifyEmailInput!) {
			verifyEmail(input: $input) {
				success
			}
		}
	`

	viewerQuery = `
		query {
			viewer { id }
		}
	`
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

	var signUpResult struct {
		SignUp struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signUp"`
	}

	newUser.Step(
		"signs up without verifying their email",
		func() error {
			err := client.ExecuteConnect(
				signUpMutation,
				map[string]any{
					"input": map[string]any{
						"email":    email,
						"password": password,
						"fullName": fullName,
					},
				},
				&signUpResult,
			)
			if err != nil {
				return fmt.Errorf("cannot sign up: %w", err)
			}

			if signUpResult.SignUp.Identity.ID == "" {
				return fmt.Errorf("sign up returned an empty identity ID")
			}

			return nil
		},
	)

	newUser.Step(
		"is not signed in after signup",
		func() error {
			err := client.ExecuteConnectShouldFail(viewerQuery, nil)
			testutil.RequireErrorCode(t, err, "UNAUTHENTICATED")

			return nil
		},
	)

	newUser.Step(
		"cannot sign in before verifying their email",
		func() error {
			err := client.SignIn(email, password)
			testutil.RequireErrorCode(t, err, "EMAIL_NOT_VERIFIED")

			return nil
		},
	)

	newUser.Step(
		"requests another verification email",
		func() error {
			client.ResendVerificationEmail(email)
			return nil
		},
	)

	var token string

	newUser.Step(
		"receives the verification email",
		func() error {
			token = client.GetEmailConfirmationToken(email)

			return nil
		},
	)

	var verifyResult struct {
		VerifyEmail struct {
			Success bool `json:"success"`
		} `json:"verifyEmail"`
	}

	newUser.Step(
		"verifies their email address and is signed in",
		func() error {
			err := client.ExecuteConnect(
				verifyEmailMutation,
				map[string]any{
					"input": map[string]any{
						"token": token,
					},
				},
				&verifyResult,
			)
			if err != nil {
				return fmt.Errorf("cannot verify email: %w", err)
			}

			if !verifyResult.VerifyEmail.Success {
				return fmt.Errorf("verify email returned success=false")
			}

			var viewer struct {
				Viewer struct {
					ID string `json:"id"`
				} `json:"viewer"`
			}

			if err := client.ExecuteConnect(viewerQuery, nil, &viewer); err != nil {
				return fmt.Errorf("expected a session after first verify: %w", err)
			}

			if viewer.Viewer.ID != signUpResult.SignUp.Identity.ID {
				return fmt.Errorf(
					"viewer %q does not match signed-up identity %q",
					viewer.Viewer.ID,
					signUpResult.SignUp.Identity.ID,
				)
			}

			return nil
		},
	)

	newUser.Step(
		"can verify the same token again without error",
		func() error {
			verifyResult.VerifyEmail.Success = false

			err := client.ExecuteConnect(
				verifyEmailMutation,
				map[string]any{
					"input": map[string]any{
						"token": token,
					},
				},
				&verifyResult,
			)
			if err != nil {
				return fmt.Errorf("replay verify email failed: %w", err)
			}

			if !verifyResult.VerifyEmail.Success {
				return fmt.Errorf("replay verify email returned success=false")
			}

			return nil
		},
	)

	replay := world.NewUnauthenticatedActor("replay client")
	replayClient := replay.Client()

	replay.Step(
		"cannot open a session by replaying the confirmation token",
		func() error {
			verifyResult.VerifyEmail.Success = false

			err := replayClient.ExecuteConnect(
				verifyEmailMutation,
				map[string]any{
					"input": map[string]any{
						"token": token,
					},
				},
				&verifyResult,
			)
			if err != nil {
				return fmt.Errorf("replay verify from a new client failed: %w", err)
			}

			if !verifyResult.VerifyEmail.Success {
				return fmt.Errorf("replay verify from a new client returned success=false")
			}

			err = replayClient.ExecuteConnectShouldFail(viewerQuery, nil)
			testutil.RequireErrorCode(t, err, "UNAUTHENTICATED")

			return nil
		},
	)

	newUser.Step(
		"can still sign in with their verified email",
		func() error {
			if err := client.SignIn(email, password); err != nil {
				return fmt.Errorf("cannot sign in after email verification: %w", err)
			}

			return nil
		},
	)
}

func TestEmailVerification_ResendIsEnumerationSafe(t *testing.T) {
	t.Parallel()

	client := testutil.NewUnauthenticatedClient(t)

	client.ResendVerificationEmail(fmt.Sprintf("missing-%d@e2e.probo.test", time.Now().UnixNano()))
}
