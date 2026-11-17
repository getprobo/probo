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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// The built-in budget is 10 requests per 60 seconds for every @rateLimit field.
const (
	signUpLimit                  = 10
	forgotPasswordLimit          = 10
	resendVerificationEmailLimit = 10
)

const (
	forgotPasswordMutation = `
		mutation($input: ForgotPasswordInput!) {
			forgotPassword(input: $input) { success }
		}
	`

	forgotPasswordAliasMutation = `
		mutation($input: ForgotPasswordInput!) {
			first: forgotPassword(input: $input) { success }
			second: forgotPassword(input: $input) { success }
		}
	`

	resendVerificationEmailMutation = `
		mutation($input: ResendVerificationEmailInput!) {
			resendVerificationEmail(input: $input) { success }
		}
	`

	rateLimitSignUpMutation = `
		mutation($input: SignUpInput!) {
			signUp(input: $input) { identity { id } }
		}
	`
)

// TestConnectRateLimit_RejectsAtConfiguredLimits uses its own probod with
// the built-in 10/60 budget. The shared suite raises that budget because
// every test comes from one client address.
func TestConnectRateLimit_RejectsAtConfiguredLimits(t *testing.T) {
	t.Parallel()

	env := testutil.StartIsolatedEnv(
		t,
		testutil.IsolatedEnvOptions{UseBuiltInRateLimit: true},
	)
	client := testutil.NewUnauthenticatedClientFor(t, env.BaseURL)

	// Eight single calls plus one aliased pair reach the limit of 10. The
	// pair checks that each alias consumes the same field budget.
	for range forgotPasswordLimit - 2 {
		requireForgotPassword(t, client)
	}

	requireForgotPasswordAliases(t, client)

	err := forgotPassword(client)
	testutil.RequireErrorCode(t, err, "RATE_LIMITED")

	for range resendVerificationEmailLimit {
		requireResendVerificationEmail(t, client)
	}

	err = resendVerificationEmail(client)
	testutil.RequireErrorCode(t, err, "RATE_LIMITED")

	for range signUpLimit {
		requireSignUp(t, client)
	}

	err = signUp(client)
	testutil.RequireErrorCode(t, err, "RATE_LIMITED")

	// signIn has its own budget, so a password attempt still reaches the
	// resolver after the other fields are exhausted.
	err = client.SignIn(factory.SafeEmail(), "TestPassword123!")
	testutil.RequireErrorCode(t, err, "INVALID_CREDENTIALS")
}

func requireForgotPassword(t *testing.T, client *testutil.Client) {
	t.Helper()

	var result struct {
		ForgotPassword struct {
			Success bool `json:"success"`
		} `json:"forgotPassword"`
	}

	err := forgotPasswordInto(client, &result)
	require.NoError(t, err, "forgotPassword")
	require.True(t, result.ForgotPassword.Success, "forgotPassword")
}

func requireForgotPasswordAliases(t *testing.T, client *testutil.Client) {
	t.Helper()

	var result struct {
		First struct {
			Success bool `json:"success"`
		} `json:"first"`
		Second struct {
			Success bool `json:"success"`
		} `json:"second"`
	}

	err := client.ExecuteConnect(
		forgotPasswordAliasMutation,
		map[string]any{
			"input": map[string]any{"email": factory.SafeEmail()},
		},
		&result,
	)
	require.NoError(t, err, "aliased forgotPassword")
	require.True(t, result.First.Success, "first forgotPassword alias")
	require.True(t, result.Second.Success, "second forgotPassword alias")
}

func forgotPassword(client *testutil.Client) error {
	var result struct {
		ForgotPassword struct {
			Success bool `json:"success"`
		} `json:"forgotPassword"`
	}

	return forgotPasswordInto(client, &result)
}

func forgotPasswordInto(client *testutil.Client, result any) error {
	return client.ExecuteConnect(
		forgotPasswordMutation,
		map[string]any{
			"input": map[string]any{"email": factory.SafeEmail()},
		},
		result,
	)
}

func requireResendVerificationEmail(t *testing.T, client *testutil.Client) {
	t.Helper()

	var result struct {
		ResendVerificationEmail struct {
			Success bool `json:"success"`
		} `json:"resendVerificationEmail"`
	}

	err := resendVerificationEmailInto(client, &result)
	require.NoError(t, err, "resendVerificationEmail")
	require.True(t, result.ResendVerificationEmail.Success, "resendVerificationEmail")
}

func resendVerificationEmail(client *testutil.Client) error {
	var result struct {
		ResendVerificationEmail struct {
			Success bool `json:"success"`
		} `json:"resendVerificationEmail"`
	}

	return resendVerificationEmailInto(client, &result)
}

func resendVerificationEmailInto(client *testutil.Client, result any) error {
	return client.ExecuteConnect(
		resendVerificationEmailMutation,
		map[string]any{
			"input": map[string]any{"email": factory.SafeEmail()},
		},
		result,
	)
}

func requireSignUp(t *testing.T, client *testutil.Client) {
	t.Helper()

	var result struct {
		SignUp struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signUp"`
	}

	err := signUpInto(client, &result)
	require.NoError(t, err, "signUp")
	require.NotEmpty(t, result.SignUp.Identity.ID, "signUp")
}

func signUp(client *testutil.Client) error {
	var result struct {
		SignUp struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signUp"`
	}

	return signUpInto(client, &result)
}

func signUpInto(client *testutil.Client, result any) error {
	return client.ExecuteConnect(
		rateLimitSignUpMutation,
		map[string]any{
			"input": map[string]any{
				"email":    factory.SafeEmail(),
				"password": "TestPassword123!",
				"fullName": "Rate Limit",
			},
		},
		result,
	)
}
