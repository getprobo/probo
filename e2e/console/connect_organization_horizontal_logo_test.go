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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestConnectOrganization_HorizontalLogoDoesNotReplaceCompliancePageLogo(t *testing.T) {
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

	orgLogoPNG := []byte{
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

	horizontalLogoPNG := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0xbc, 0x18, 0x19,
		0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	const updateOrganizationMutation = `
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization {
					id
					logo {
						id
						fileName
					}
					horizontalLogo {
						id
						fileName
					}
				}
			}
		}
	`

	var orgLogoUploadResult struct {
		UpdateOrganization struct {
			Organization struct {
				Logo *struct {
					ID       string `json:"id"`
					FileName string `json:"fileName"`
				} `json:"logo"`
			} `json:"organization"`
		} `json:"updateOrganization"`
	}

	err = owner.ExecuteConnectWithFile(updateOrganizationMutation, map[string]any{
		"input": map[string]any{
			"organizationId": organizationID,
			"logoFile":       nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "org-logo.png",
		ContentType: "image/png",
		Content:     orgLogoPNG,
	}, &orgLogoUploadResult)
	require.NoError(t, err)
	require.NotNil(t, orgLogoUploadResult.UpdateOrganization.Organization.Logo)

	orgLogoID := orgLogoUploadResult.UpdateOrganization.Organization.Logo.ID

	const trustGraphQLQuery = `
		query {
			currentTrustCenter {
				logo {
					id
					fileName
				}
			}
		}
	`

	var trustResultAfterOrgLogo struct {
		CurrentTrustCenter struct {
			Logo *struct {
				ID       string `json:"id"`
				FileName string `json:"fileName"`
			} `json:"logo"`
		} `json:"currentTrustCenter"`
	}

	err = owner.ExecuteTrust(trustCenterID, trustGraphQLQuery, nil, &trustResultAfterOrgLogo)
	require.NoError(t, err)
	require.NotNil(t, trustResultAfterOrgLogo.CurrentTrustCenter.Logo)
	assert.Equal(t, orgLogoID, trustResultAfterOrgLogo.CurrentTrustCenter.Logo.ID)

	var horizontalLogoUploadResult struct {
		UpdateOrganization struct {
			Organization struct {
				HorizontalLogo *struct {
					ID       string `json:"id"`
					FileName string `json:"fileName"`
				} `json:"horizontalLogo"`
			} `json:"organization"`
		} `json:"updateOrganization"`
	}

	err = owner.ExecuteConnectWithFile(updateOrganizationMutation, map[string]any{
		"input": map[string]any{
			"organizationId":     organizationID,
			"horizontalLogoFile": nil,
		},
	}, "input.horizontalLogoFile", testutil.UploadFile{
		Filename:    "horizontal-logo.png",
		ContentType: "image/png",
		Content:     horizontalLogoPNG,
	}, &horizontalLogoUploadResult)
	require.NoError(t, err)
	require.NotNil(t, horizontalLogoUploadResult.UpdateOrganization.Organization.HorizontalLogo)

	horizontalLogoID := horizontalLogoUploadResult.UpdateOrganization.Organization.HorizontalLogo.ID
	assert.NotEqual(t, orgLogoID, horizontalLogoID)

	var trustResultAfterHorizontalLogo struct {
		CurrentTrustCenter struct {
			Logo *struct {
				ID       string `json:"id"`
				FileName string `json:"fileName"`
			} `json:"logo"`
		} `json:"currentTrustCenter"`
	}

	err = owner.ExecuteTrust(trustCenterID, trustGraphQLQuery, nil, &trustResultAfterHorizontalLogo)
	require.NoError(t, err)
	require.NotNil(t, trustResultAfterHorizontalLogo.CurrentTrustCenter.Logo)
	assert.Equal(t, orgLogoID, trustResultAfterHorizontalLogo.CurrentTrustCenter.Logo.ID)
	assert.NotEqual(t, horizontalLogoID, trustResultAfterHorizontalLogo.CurrentTrustCenter.Logo.ID)
}
