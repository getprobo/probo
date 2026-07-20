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

package trust_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_LogoFileDownloadURL(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const compliancePortalQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					compliancePortal {
						id
					}
				}
			}
		}
	`

	var compliancePortalLookup struct {
		Node struct {
			CompliancePortal struct {
				ID string `json:"id"`
			} `json:"compliancePortal"`
		} `json:"node"`
	}

	err := owner.Execute(compliancePortalQuery, map[string]any{
		"organizationId": organizationID,
	}, &compliancePortalLookup)
	require.NoError(t, err)
	require.NotEmpty(t, compliancePortalLookup.Node.CompliancePortal.ID)

	compliancePortalID := compliancePortalLookup.Node.CompliancePortal.ID

	const activateMutation = `
		mutation($input: UpdateCompliancePortalInput!) {
			updateCompliancePortal(input: $input) {
				compliancePortal {
					id
					active
					publicUrl
				}
			}
		}
	`

	var activateResult struct {
		UpdateCompliancePortal struct {
			CompliancePortal struct {
				ID        string `json:"id"`
				Active    bool   `json:"active"`
				PublicURL string `json:"publicUrl"`
			} `json:"compliancePortal"`
		} `json:"updateCompliancePortal"`
	}

	err = owner.Execute(activateMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"active":             true,
		},
	}, &activateResult)
	require.NoError(t, err)

	// Publishing the page provisions a managed {slug}.probopage.localhost
	// domain; the effective public URL resolves to it while no customer
	// custom domain is primary.
	require.NotEmpty(t, activateResult.UpdateCompliancePortal.CompliancePortal.PublicURL)

	publicURL, err := url.Parse(activateResult.UpdateCompliancePortal.CompliancePortal.PublicURL)
	require.NoError(t, err)

	trustHost := publicURL.Host
	require.NotEmpty(t, trustHost)

	const uploadMutation = `
		mutation UpdateCompliancePortalBrand($input: UpdateCompliancePortalBrandInput!) {
			updateCompliancePortalBrand(input: $input) {
				compliancePortal {
					id
					logo {
						id
						fileName
						downloadUrl
					}
				}
			}
		}
	`

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var uploadResult struct {
		UpdateCompliancePortalBrand struct {
			CompliancePortal struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"compliancePortal"`
		} `json:"updateCompliancePortalBrand"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"logoFile":           nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "trust-center-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)
	require.NotNil(t, uploadResult.UpdateCompliancePortalBrand.CompliancePortal.Logo)

	const trustGraphQLQuery = `
		query {
			currentCompliancePortal {
				logo {
					id
					fileName
					downloadUrl
				}
			}
		}
	`

	var trustResult struct {
		CurrentCompliancePortal struct {
			Logo *struct {
				ID          string `json:"id"`
				FileName    string `json:"fileName"`
				DownloadURL string `json:"downloadUrl"`
			} `json:"logo"`
		} `json:"currentCompliancePortal"`
	}

	// The dedicated HTTPS listener only serves the page once the managed
	// domain's certificate has been provisioned (async, ~1s poll in e2e), so
	// retry until the TLS handshake and query succeed.
	require.Eventually(t, func() bool {
		trustResult.CurrentCompliancePortal.Logo = nil
		if err := owner.ExecuteTrust(trustHost, trustGraphQLQuery, nil, &trustResult); err != nil {
			return false
		}

		return trustResult.CurrentCompliancePortal.Logo != nil
	}, 30*time.Second, 500*time.Millisecond, "compliance portal did not become servable on the dedicated listener")

	require.NotNil(t, trustResult.CurrentCompliancePortal.Logo)
	assert.Equal(t, uploadResult.UpdateCompliancePortalBrand.CompliancePortal.Logo.ID, trustResult.CurrentCompliancePortal.Logo.ID)

	downloadURL := trustResult.CurrentCompliancePortal.Logo.DownloadURL
	assert.True(
		t,
		strings.Contains(downloadURL, "/api/files/v1/public/"),
		"downloadUrl must route through the public files API, got %q",
		downloadURL,
	)

	// Match the e2e HTTP client convention (see internal/testutil) so a hung
	// server fails the request instead of blocking the parallel suite forever.
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// The public endpoint streams the file bytes directly (no presigned
	// redirect) with cache headers, so the stable URL is CDN/browser cacheable.
	resp, err := httpClient.Get(downloadURL)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, pngContent, body, "public endpoint must stream the original file bytes")

	cacheControl := resp.Header.Get("Cache-Control")
	assert.Contains(t, cacheControl, "public")
	assert.Contains(t, cacheControl, "max-age=31536000")
	assert.Contains(t, cacheControl, "immutable")

	etag := resp.Header.Get("ETag")
	require.NotEmpty(t, etag, "public file response must carry an ETag for revalidation")

	// Untrusted, unauthenticated content served from the app origin must be
	// hardened against MIME sniffing and script execution (e.g. SVG uploads).
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "sandbox")

	// A matching If-None-Match must revalidate to 304 without transferring
	// the body again.
	revalidateReq, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	require.NoError(t, err)
	revalidateReq.Header.Set("If-None-Match", etag)

	revalidateResp, err := httpClient.Do(revalidateReq)
	require.NoError(t, err)

	defer func() { _ = revalidateResp.Body.Close() }()

	assert.Equal(t, http.StatusNotModified, revalidateResp.StatusCode)

	// A matching If-Modified-Since (using the returned Last-Modified) must also
	// revalidate to 304.
	lastModified := resp.Header.Get("Last-Modified")
	require.NotEmpty(t, lastModified, "public file response must carry Last-Modified")

	sinceReq, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	require.NoError(t, err)
	sinceReq.Header.Set("If-Modified-Since", lastModified)

	sinceResp, err := httpClient.Do(sinceReq)
	require.NoError(t, err)

	defer func() { _ = sinceResp.Body.Close() }()

	assert.Equal(t, http.StatusNotModified, sinceResp.StatusCode)
}
