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

	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	membershipAccessCreateOrganizationMutation = `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
				organization { id }
			}
		}
	`

	membershipAccessConsoleProbeQuery = `
		query {
			__typename
		}
	`

	membershipAccessDeactivateUserMutation = `
		mutation($input: DeactivateUserInput!) {
			deactivateUser(input: $input) {
				success
			}
		}
	`
)

func createOrganizationAs(
	t *testing.T,
	client *testutil.Client,
	name string,
) (string, error) {
	t.Helper()

	var result struct {
		CreateOrganization struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"createOrganization"`
	}

	err := client.ExecuteConnect(
		membershipAccessCreateOrganizationMutation,
		map[string]any{
			"input": map[string]any{
				"name": name,
			},
		},
		&result,
	)
	if err != nil {
		return "", err
	}

	return result.CreateOrganization.Organization.ID, nil
}

func createOrganizationShouldFail(
	t *testing.T,
	client *testutil.Client,
	name string,
) error {
	t.Helper()

	return client.ExecuteConnectShouldFail(
		membershipAccessCreateOrganizationMutation,
		map[string]any{
			"input": map[string]any{
				"name": name,
			},
		},
	)
}

func consoleProbeShouldFail(t *testing.T, client *testutil.Client) error {
	t.Helper()

	return client.ExecuteShouldFail(membershipAccessConsoleProbeQuery, nil)
}

func deactivateUser(
	t *testing.T,
	owner *testutil.Client,
	member *testutil.Client,
) error {
	t.Helper()

	var result struct {
		DeactivateUser struct {
			Success bool `json:"success"`
		} `json:"deactivateUser"`
	}

	err := owner.ExecuteConnect(
		membershipAccessDeactivateUserMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"profileId":      member.GetProfileID().String(),
			},
		},
		&result,
	)
	if err != nil {
		return fmt.Errorf("cannot deactivate user: %w", err)
	}

	if !result.DeactivateUser.Success {
		return fmt.Errorf("deactivate user returned success=false")
	}

	return nil
}
