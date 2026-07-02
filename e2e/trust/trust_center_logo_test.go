// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package trust_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTrustCenter_LogoFileDownloadURL(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const trustCenterQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter {
						id
					}
				}
			}
		}
	`

	var trustCenterLookup struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(trustCenterQuery, map[string]any{
		"organizationId": organizationID,
	}, &trustCenterLookup)
	require.NoError(t, err)
	require.NotEmpty(t, trustCenterLookup.Node.TrustCenter.ID)

	trustCenterID := trustCenterLookup.Node.TrustCenter.ID

	const activateMutation = `
		mutation($input: UpdateTrustCenterInput!) {
			updateTrustCenter(input: $input) {
				trustCenter {
					id
					active
				}
			}
		}
	`

	err = owner.Execute(activateMutation, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"active":        true,
		},
	}, nil)
	require.NoError(t, err)

	const uploadMutation = `
		mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
			updateTrustCenterBrand(input: $input) {
				trustCenter {
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
		UpdateTrustCenterBrand struct {
			TrustCenter struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"trustCenter"`
		} `json:"updateTrustCenterBrand"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"logoFile":      nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "trust-center-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)
	require.NotNil(t, uploadResult.UpdateTrustCenterBrand.TrustCenter.Logo)

	const trustGraphQLQuery = `
		query {
			currentTrustCenter {
				logo {
					id
					fileName
					downloadUrl
				}
			}
		}
	`

	var trustResult struct {
		CurrentTrustCenter struct {
			Logo *struct {
				ID          string `json:"id"`
				FileName    string `json:"fileName"`
				DownloadURL string `json:"downloadUrl"`
			} `json:"logo"`
		} `json:"currentTrustCenter"`
	}

	err = owner.ExecuteTrust(trustCenterID, trustGraphQLQuery, nil, &trustResult)
	require.NoError(t, err)
	require.NotNil(t, trustResult.CurrentTrustCenter.Logo)
	assert.Equal(t, uploadResult.UpdateTrustCenterBrand.TrustCenter.Logo.ID, trustResult.CurrentTrustCenter.Logo.ID)
	assert.True(
		t,
		strings.Contains(trustResult.CurrentTrustCenter.Logo.DownloadURL, "/api/files/v1/public/"),
		"downloadUrl must route through the public files API, got %q",
		trustResult.CurrentTrustCenter.Logo.DownloadURL,
	)
}
