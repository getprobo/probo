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
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_LLMsTxt(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	portalID := compliancePortalID(t, owner)

	const (
		description  = "Private inference inside secure hardware enclaves."
		website      = "https://tinfoil.example"
		email        = "security@tinfoil.example"
		headquarters = "135 Pierce St"
	)

	err := owner.Execute(
		`
		mutation UpdateCompliancePortal($input: UpdateCompliancePortalInput!) {
			updateCompliancePortal(input: $input) {
				compliancePortal { id }
			}
		}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"description":        description,
				"websiteUrl":         website,
				"email":              email,
				"headquarterAddress": headquarters,
			},
		},
		nil,
	)
	require.NoError(t, err)

	orgName := llmsTxtOrganizationName(t, owner)
	frameworkName := factory.SafeName("SOC2")
	frameworkID := factory.NewFramework(owner).
		WithName(frameworkName).
		Create()

	err = owner.Execute(
		`
		mutation($input: CreateComplianceFrameworkInput!) {
			createComplianceFramework(input: $input) {
				complianceFrameworkEdge { node { id } }
			}
		}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"frameworkId":        frameworkID,
			},
		},
		nil,
	)
	require.NoError(t, err)

	vendorName := factory.SafeName("Acme")
	vendorID := factory.NewThirdParty(owner).
		WithName(vendorName).
		WithCategory("CLOUD_PROVIDER").
		WithWebsiteUrl("https://acme.example").
		Create()
	publishLLMsTxtSubprocessor(t, owner, portalID, vendorID, []string{"US"})

	linkName := factory.SafeName("Status")
	err = owner.Execute(
		`
		mutation($input: CreateComplianceCustomLinkInput!) {
			createComplianceCustomLink(input: $input) {
				complianceCustomLinkEdge { node { id } }
			}
		}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"name":               linkName,
				"url":                "https://status.example",
			},
		},
		nil,
	)
	require.NoError(t, err)

	trustHost := llmsTxtTrustHost(t, owner, portalID)
	// Certificate polling waits out the provisioning lease before the CA
	// validates the managed domain, so the shared 30s helper is too short.
	waitForLLMsTxtPortal(t, trustHost)

	client := testutil.TrustHTTPClient(trustHost)
	resp, err := client.Get("https://" + trustHost + "/llms.txt")
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	markdown := string(body)
	require.Equal(t, http.StatusOK, resp.StatusCode, markdown)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")
	assert.Contains(t, markdown, "# "+orgName+" — Compliance")
	assert.Contains(t, markdown, description)
	assert.Contains(t, markdown, "- **Website**: "+website)
	assert.Contains(t, markdown, "- **Email**: "+email)
	assert.Contains(t, markdown, "- **Headquarters**: "+headquarters)
	assert.Contains(t, markdown, "- "+frameworkName)
	assert.Contains(t, markdown, "## Subprocessors")
	assert.Contains(
		t,
		markdown,
		"| "+vendorName+" | CLOUD_PROVIDER | US | https://acme.example |",
	)
	assert.Contains(t, markdown, "## External Links")
	assert.Contains(t, markdown, "| "+linkName+" | https://status.example |")
	assert.NotContains(t, markdown, "internal server error")
}

func waitForLLMsTxtPortal(t *testing.T, host string) {
	t.Helper()

	client := testutil.TrustHTTPClient(host)

	require.Eventually(
		t,
		func() bool {
			resp, err := client.Get("https://" + host + "/.well-known/oauth-client-metadata")
			if err != nil {
				return false
			}

			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode == http.StatusOK
		},
		3*time.Minute,
		time.Second,
		"compliance portal did not become servable on the dedicated listener",
	)
}

func llmsTxtOrganizationName(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	var result struct {
		Node struct {
			Name string `json:"name"`
		} `json:"node"`
	}

	err := owner.Execute(
		`
		query($id: ID!) {
			node(id: $id) {
				... on Organization { name }
			}
		}
		`,
		map[string]any{"id": owner.GetOrganizationID().String()},
		&result,
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.Name)

	return result.Node.Name
}

func llmsTxtTrustHost(t *testing.T, owner *testutil.Client, portalID string) string {
	t.Helper()

	testutil.ActivateCompliancePortal(t, owner, portalID)

	var result struct {
		Node struct {
			PublicURL string `json:"publicUrl"`
		} `json:"node"`
	}

	err := owner.Execute(
		`
		query($compliancePortalId: ID!) {
			node(id: $compliancePortalId) {
				... on CompliancePortal { publicUrl }
			}
		}
		`,
		map[string]any{"compliancePortalId": portalID},
		&result,
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.PublicURL)

	publicURL, err := url.Parse(result.Node.PublicURL)
	require.NoError(t, err)
	require.NotEmpty(t, publicURL.Host)

	return publicURL.Host
}

func publishLLMsTxtSubprocessor(
	t *testing.T,
	owner *testutil.Client,
	portalID string,
	thirdPartyID string,
	countries []string,
) {
	t.Helper()

	err := owner.Execute(
		`
		mutation($input: UpdateThirdPartyInput!) {
			updateThirdParty(input: $input) {
				thirdParty { id }
			}
		}
		`,
		map[string]any{
			"input": map[string]any{
				"id":        thirdPartyID,
				"countries": countries,
			},
		},
		nil,
	)
	require.NoError(t, err)

	err = owner.Execute(
		`
		mutation($input: UpdateCompliancePortalThirdPartyPublishedInput!) {
			updateCompliancePortalThirdPartyPublished(input: $input) {
				catalogThirdParty { thirdParty { id } }
			}
		}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"thirdPartyId":       thirdPartyID,
				"published":          true,
			},
		},
		nil,
	)
	require.NoError(t, err)
}
