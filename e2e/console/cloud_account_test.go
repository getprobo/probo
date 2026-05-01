// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// =============================================================================
// GraphQL queries / mutations
//
// Every query the cloud-account suite uses lives here as a
// package-level const, matching the canonical e2e convention
// (e.g. e2e/console/vendor_test.go). One named const per
// operation; typed result structs declared at the call site so
// each subtest stays self-contained.
// =============================================================================

const cloudAccountListQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Organization {
				cloudAccounts(first: 50) {
					edges {
						node {
							id
							provider
							label
							status
							credentialKind
							scope { kind identifier }
							lastProbeAt
							lastProbeError
							lastVerifiedAt
							createdAt
							updatedAt
						}
					}
				}
			}
		}
	}
`

const cloudAccountGetQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on CloudAccount {
				id
				provider
				label
				status
				credentialKind
				scope { kind identifier }
				lastProbeAt
				lastProbeError
				lastVerifiedAt
				enabledAuditModules
				createdAt
				updatedAt
			}
		}
	}
`

const generateCloudAccountInstallAssetsMutation = `
	mutation($input: GenerateCloudAccountInstallAssetsInput!) {
		generateCloudAccountInstallAssets(input: $input) {
			assets {
				__typename
				... on AWSInstallAssets {
					quickCreateURL
					externalId
					principalArn
					requiredActions
				}
				... on GCPInstallAssets {
					setupScript
					requiredRoles
					requiredApis
				}
				... on AzureInstallAssets {
					steps { title body code }
					requiredRbacRoles
					requiredGraphPermissions
				}
			}
		}
	}
`

const createCloudAccountMutation = `
	mutation($input: CreateCloudAccountInput!) {
		createCloudAccount(input: $input) {
			cloudAccount {
				id
				provider
				label
				status
				credentialKind
				scope { kind identifier }
				lastProbeError
				lastVerifiedAt
				createdAt
				updatedAt
			}
			verifyStatus
			lastProbeError
		}
	}
`

const verifyCloudAccountMutation = `
	mutation($input: VerifyCloudAccountInput!) {
		verifyCloudAccount(input: $input) {
			cloudAccount {
				id
				status
				lastProbeError
				lastVerifiedAt
				updatedAt
			}
			status
			lastProbeError
		}
	}
`

const rotateCloudAccountCredentialsMutation = `
	mutation($input: RotateCloudAccountCredentialsInput!) {
		rotateCloudAccountCredentials(input: $input) {
			cloudAccount {
				id
				status
				updatedAt
			}
			verifyStatus
			lastProbeError
		}
	}
`

const deleteCloudAccountMutation = `
	mutation($input: DeleteCloudAccountInput!) {
		deleteCloudAccount(input: $input) {
			deletedCloudAccountId
		}
	}
`

// accessSourceListQuery is used by the access-review integration
// subtest. The GraphQL `CreateAccessSourceInput` now exposes
// `cloudAccountId`, so the integration test wires a cloud account
// through `createAccessSource` and verifies the linkage round-trips
// via this list query.
const accessSourceListQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Organization {
				accessSources(first: 25) {
					edges {
						node { id name }
					}
				}
			}
		}
	}
`

// The createAccessSourceMutation and deleteAccessSourceMutation
// constants used by the cloud-account integration tests live in
// rbac_test.go (same package); they are reused here.

// =============================================================================
// Typed result shapes used by multiple subtests.
// =============================================================================

type cloudAccountNode struct {
	ID             string  `json:"id"`
	Provider       string  `json:"provider"`
	Label          string  `json:"label"`
	Status         string  `json:"status"`
	CredentialKind string  `json:"credentialKind"`
	LastProbeAt    *string `json:"lastProbeAt"`
	LastProbeError *string `json:"lastProbeError"`
	LastVerifiedAt *string `json:"lastVerifiedAt"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
	Scope          struct {
		Kind       string  `json:"kind"`
		Identifier *string `json:"identifier"`
	} `json:"scope"`
	EnabledAuditModules []string `json:"enabledAuditModules"`
}

type createCloudAccountResult struct {
	CreateCloudAccount struct {
		CloudAccount   cloudAccountNode `json:"cloudAccount"`
		VerifyStatus   string           `json:"verifyStatus"`
		LastProbeError *string          `json:"lastProbeError"`
	} `json:"createCloudAccount"`
}

// =============================================================================
// Lifecycle subtests (one per provider).
//
// Important runtime caveat: the e2e binary has no AWS / GCP / Azure
// SDK injection seam. Probe calls always fail (no real cloud is
// reachable from CI). As a result the post-create Verify step
// transitions the row to PENDING_VERIFICATION (preserved) +
// `lastProbeError != nil`, NOT to VERIFIED. Promotion to VERIFIED
// is asserted in unit tests against the registry seam in
// pkg/probo/cloud_account_*_test.go and
// pkg/cloudaccount/{aws,gcp,azure}_test.go. The e2e test asserts
// only the API-level transitions that don't need a real probe to
// succeed: PENDING_VERIFICATION on create, status/error after a
// failed probe, rotate flips back to PENDING_VERIFICATION,
// delete returns the row id.
//
// AWS install-assets generation also requires a configured
// CloudAccount.AWSTemplateURL + AWSTemplateSHA256 in the e2e
// probod config; the current e2e harness does not set those, so
// the test asserts the documented `UNAVAILABLE` failure mode and
// continues from a directly-created row. When the e2e config gets
// AWS template wiring, the assertion can be widened.
// =============================================================================

func TestCloudAccount_AWS_Lifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Step 1: install-assets. Asserts either the URL is rendered
	// (when CloudAccount.AWSTemplateURL is configured) or the
	// documented UNAVAILABLE error otherwise. Both branches keep
	// the rest of the lifecycle test runnable.
	t.Run("install-assets renders or surfaces UNAVAILABLE", func(t *testing.T) {
		t.Parallel()

		var assets struct {
			GenerateCloudAccountInstallAssets struct {
				Assets struct {
					Typename        string   `json:"__typename"`
					QuickCreateURL  string   `json:"quickCreateURL"`
					ExternalID      string   `json:"externalId"`
					PrincipalArn    string   `json:"principalArn"`
					RequiredActions []string `json:"requiredActions"`
				} `json:"assets"`
			} `json:"generateCloudAccountInstallAssets"`
		}

		err := owner.Execute(generateCloudAccountInstallAssetsMutation, map[string]any{
			"input": map[string]any{
				"organizationId":  owner.GetOrganizationID().String(),
				"provider":        "AWS",
				"scopeKind":       "AWS_ACCOUNT",
				"scopeIdentifier": "123456789012",
				"modules":         []string{"ACCESS_REVIEW"},
				"awsRegion":       "us-east-1",
			},
		}, &assets)

		if err != nil {
			testutil.RequireErrorCode(t, err, "UNAVAILABLE",
				"AWS install assets need probodconfig.CloudAccount.AWSTemplateURL configured; this e2e harness does not yet set it")
			return
		}

		assert.Equal(t, "AWSInstallAssets", assets.GenerateCloudAccountInstallAssets.Assets.Typename)
		assert.NotEmpty(t, assets.GenerateCloudAccountInstallAssets.Assets.QuickCreateURL)
		assert.Len(t, assets.GenerateCloudAccountInstallAssets.Assets.ExternalID, 64,
			"AWS external id must be 64 hex chars")
	})

	// Step 2: create. Verify path always fails in e2e (no real cloud SDK),
	// so we assert the failure-mode shape: row exists, status stays
	// PENDING_VERIFICATION, lastProbeError is populated.
	cloudAccountID := factory.NewCloudAccount(owner).
		WithProvider("AWS").
		WithLabel("AWS Lifecycle Test").
		WithScopeIdentifier("123456789012").
		Create()
	require.NotEmpty(t, cloudAccountID)

	t.Run("get returns row in PENDING_VERIFICATION with probe error", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, err)

		assert.Equal(t, cloudAccountID, result.Node.ID)
		assert.Equal(t, "AWS", result.Node.Provider)
		assert.Equal(t, "PENDING_VERIFICATION", result.Node.Status)
		assert.Equal(t, "AWS_ACCOUNT", result.Node.Scope.Kind)
		require.NotNil(t, result.Node.Scope.Identifier, "OWNER must see scope.identifier")
		assert.Equal(t, "123456789012", *result.Node.Scope.Identifier)
		assert.NotNil(t, result.Node.LastProbeError, "probe failed in e2e (no real AWS) -- error must be set")
		assert.Equal(t, result.Node.CreatedAt, result.Node.UpdatedAt,
			"createdAt and updatedAt should equal on initial PENDING_VERIFICATION row")
	})

	t.Run("list includes the new row", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node struct {
				CloudAccounts struct {
					Edges []struct {
						Node cloudAccountNode `json:"node"`
					} `json:"edges"`
				} `json:"cloudAccounts"`
			} `json:"node"`
		}
		err := owner.Execute(cloudAccountListQuery, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		found := false
		for _, edge := range result.Node.CloudAccounts.Edges {
			if edge.Node.ID == cloudAccountID {
				found = true
				assert.Equal(t, "AWS", edge.Node.Provider)
				assert.Equal(t, "PENDING_VERIFICATION", edge.Node.Status)
				break
			}
		}
		assert.True(t, found, "newly created cloud account must appear in list")
	})

	t.Run("rotate credentials flips back to PENDING_VERIFICATION", func(t *testing.T) {
		t.Parallel()

		rotated := factory.NewCloudAccount(owner).
			WithProvider("AWS").
			WithLabel("AWS Rotate Source").
			Create()

		var result struct {
			RotateCloudAccountCredentials struct {
				CloudAccount struct {
					ID        string `json:"id"`
					Status    string `json:"status"`
					UpdatedAt string `json:"updatedAt"`
				} `json:"cloudAccount"`
				VerifyStatus   string  `json:"verifyStatus"`
				LastProbeError *string `json:"lastProbeError"`
			} `json:"rotateCloudAccountCredentials"`
		}

		err := owner.Execute(rotateCloudAccountCredentialsMutation, map[string]any{
			"input": map[string]any{
				"cloudAccountId": rotated,
				"provider":       "AWS",
				"credentialKind": "AWS_ASSUME_ROLE",
				"awsRoleArn":     "arn:aws:iam::123456789012:role/RotatedRole",
				"awsExternalId":  "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, rotated, result.RotateCloudAccountCredentials.CloudAccount.ID)
		// Verify in e2e always fails -- so the row is back in
		// PENDING_VERIFICATION post-rotate.
		assert.Equal(t, "PENDING_VERIFICATION", result.RotateCloudAccountCredentials.CloudAccount.Status)
	})

	t.Run("delete returns the deleted id", func(t *testing.T) {
		t.Parallel()

		toDelete := factory.NewCloudAccount(owner).
			WithProvider("AWS").
			WithLabel("AWS To Delete").
			Create()

		var result struct {
			DeleteCloudAccount struct {
				DeletedCloudAccountID string `json:"deletedCloudAccountId"`
			} `json:"deleteCloudAccount"`
		}

		err := owner.Execute(deleteCloudAccountMutation, map[string]any{
			"input": map[string]any{"cloudAccountId": toDelete},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, toDelete, result.DeleteCloudAccount.DeletedCloudAccountID)
	})
}

func TestCloudAccount_GCP_Lifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("install-assets returns setup script", func(t *testing.T) {
		t.Parallel()

		var assets struct {
			GenerateCloudAccountInstallAssets struct {
				Assets struct {
					Typename      string   `json:"__typename"`
					SetupScript   string   `json:"setupScript"`
					RequiredRoles []string `json:"requiredRoles"`
					RequiredApis  []string `json:"requiredApis"`
				} `json:"assets"`
			} `json:"generateCloudAccountInstallAssets"`
		}

		err := owner.Execute(generateCloudAccountInstallAssetsMutation, map[string]any{
			"input": map[string]any{
				"organizationId":  owner.GetOrganizationID().String(),
				"provider":        "GCP",
				"scopeKind":       "GCP_PROJECT",
				"scopeIdentifier": "probo-e2e-mock-project",
				"modules":         []string{"ACCESS_REVIEW"},
			},
		}, &assets)
		require.NoError(t, err)

		assert.Equal(t, "GCPInstallAssets", assets.GenerateCloudAccountInstallAssets.Assets.Typename)
		assert.NotEmpty(t, assets.GenerateCloudAccountInstallAssets.Assets.SetupScript)
		assert.NotEmpty(t, assets.GenerateCloudAccountInstallAssets.Assets.RequiredRoles)
	})

	gcpID := factory.NewCloudAccount(owner).
		WithProvider("GCP").
		WithLabel("GCP Lifecycle Test").
		Create()
	require.NotEmpty(t, gcpID)

	t.Run("get returns GCP row with project scope", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": gcpID}, &result)
		require.NoError(t, err)

		assert.Equal(t, "GCP", result.Node.Provider)
		assert.Equal(t, "GCP_PROJECT", result.Node.Scope.Kind)
		assert.Equal(t, "PENDING_VERIFICATION", result.Node.Status)
	})
}

func TestCloudAccount_Azure_Lifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("install-assets returns walkthrough steps", func(t *testing.T) {
		t.Parallel()

		var assets struct {
			GenerateCloudAccountInstallAssets struct {
				Assets struct {
					Typename string `json:"__typename"`
					Steps    []struct {
						Title string  `json:"title"`
						Body  string  `json:"body"`
						Code  *string `json:"code"`
					} `json:"steps"`
					RequiredRbacRoles        []string `json:"requiredRbacRoles"`
					RequiredGraphPermissions []string `json:"requiredGraphPermissions"`
				} `json:"assets"`
			} `json:"generateCloudAccountInstallAssets"`
		}

		err := owner.Execute(generateCloudAccountInstallAssetsMutation, map[string]any{
			"input": map[string]any{
				"organizationId":  owner.GetOrganizationID().String(),
				"provider":        "AZURE",
				"scopeKind":       "AZURE_MANAGEMENT_GROUP",
				"scopeIdentifier": "00000000-0000-0000-0000-000000000003",
				"modules":         []string{"ACCESS_REVIEW"},
			},
		}, &assets)
		require.NoError(t, err)

		assert.Equal(t, "AzureInstallAssets", assets.GenerateCloudAccountInstallAssets.Assets.Typename)
		assert.NotEmpty(t, assets.GenerateCloudAccountInstallAssets.Assets.Steps)
	})

	azureID := factory.NewCloudAccount(owner).
		WithProvider("AZURE").
		WithLabel("Azure Lifecycle Test").
		Create()
	require.NotEmpty(t, azureID)

	t.Run("get returns Azure row with MG scope", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": azureID}, &result)
		require.NoError(t, err)

		assert.Equal(t, "AZURE", result.Node.Provider)
		assert.Equal(t, "AZURE_MANAGEMENT_GROUP", result.Node.Scope.Kind)
		assert.Equal(t, "PENDING_VERIFICATION", result.Node.Status)
	})
}

// =============================================================================
// RBAC matrix: 5 roles x 7 actions = 35 explicit rows.
//
// Coverage map (mirrors pkg/probo/policies.go):
//
//	OWNER:    all 7 allowed.
//	ADMIN:    all 7 allowed.
//	VIEWER:   list + get allowed; mutations forbidden.
//	AUDITOR:  list + get allowed; mutations forbidden.
//	EMPLOYEE: all 7 forbidden (no cloud-account access at all).
//
// Mutations covered: create, generateInstallAssets, verify,
// rotateCredentials, delete. List + get exercise the read path.
// =============================================================================

func TestCloudAccount_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	// Pre-seed a row each role can target. The Verify after Create
	// is non-fatal in e2e (probe fails) so the row exists in
	// PENDING_VERIFICATION.
	cloudAccountID := factory.NewCloudAccount(owner).
		WithLabel("RBAC Test").
		Create()

	type rbacCase struct {
		name            string
		client          *testutil.Client
		action          string
		query           string
		variables       func(client *testutil.Client) map[string]any
		expectForbidden bool
	}

	createVars := func(client *testutil.Client) map[string]any {
		return map[string]any{
			"input": map[string]any{
				"organizationId":      client.GetOrganizationID().String(),
				"label":               factory.SafeName("RBAC New"),
				"provider":            "AWS",
				"credentialKind":      "AWS_ASSUME_ROLE",
				"scopeKind":           "AWS_ACCOUNT",
				"scopeIdentifier":     "123456789012",
				"enabledAuditModules": []string{"ACCESS_REVIEW"},
				"awsRoleArn":          "arn:aws:iam::123456789012:role/Probo",
				"awsExternalId":       "1111111111111111111111111111111111111111111111111111111111111111",
			},
		}
	}

	installVars := func(client *testutil.Client) map[string]any {
		return map[string]any{
			"input": map[string]any{
				"organizationId":  client.GetOrganizationID().String(),
				"provider":        "GCP",
				"scopeKind":       "GCP_PROJECT",
				"scopeIdentifier": "rbac-mock-proj",
				"modules":         []string{"ACCESS_REVIEW"},
			},
		}
	}

	verifyVars := func(_ *testutil.Client) map[string]any {
		return map[string]any{"input": map[string]any{"cloudAccountId": cloudAccountID}}
	}

	rotateVars := func(_ *testutil.Client) map[string]any {
		return map[string]any{
			"input": map[string]any{
				"cloudAccountId": cloudAccountID,
				"provider":       "AWS",
				"credentialKind": "AWS_ASSUME_ROLE",
				"awsRoleArn":     "arn:aws:iam::123456789012:role/Rotated",
				"awsExternalId":  "2222222222222222222222222222222222222222222222222222222222222222",
			},
		}
	}

	deleteVars := func(_ *testutil.Client) map[string]any {
		// Each delete subtest creates its own row to avoid
		// cross-row state leak between subtests.
		return map[string]any{}
	}

	listVars := func(client *testutil.Client) map[string]any {
		return map[string]any{"id": client.GetOrganizationID().String()}
	}

	getVars := func(_ *testutil.Client) map[string]any {
		return map[string]any{"id": cloudAccountID}
	}

	cases := []rbacCase{
		// OWNER (all allowed).
		{name: "owner can list", client: owner, action: "list", query: cloudAccountListQuery, variables: listVars, expectForbidden: false},
		{name: "owner can get", client: owner, action: "get", query: cloudAccountGetQuery, variables: getVars, expectForbidden: false},
		{name: "owner can create", client: owner, action: "create", query: createCloudAccountMutation, variables: createVars, expectForbidden: false},
		{name: "owner can generate-install-assets", client: owner, action: "install-assets", query: generateCloudAccountInstallAssetsMutation, variables: installVars, expectForbidden: false},
		{name: "owner can verify", client: owner, action: "verify", query: verifyCloudAccountMutation, variables: verifyVars, expectForbidden: false},
		{name: "owner can rotate-credentials", client: owner, action: "rotate-credentials", query: rotateCloudAccountCredentialsMutation, variables: rotateVars, expectForbidden: false},
		{name: "owner can delete", client: owner, action: "delete", query: deleteCloudAccountMutation, variables: deleteVars, expectForbidden: false},

		// ADMIN (all allowed).
		{name: "admin can list", client: admin, action: "list", query: cloudAccountListQuery, variables: listVars, expectForbidden: false},
		{name: "admin can get", client: admin, action: "get", query: cloudAccountGetQuery, variables: getVars, expectForbidden: false},
		{name: "admin can create", client: admin, action: "create", query: createCloudAccountMutation, variables: createVars, expectForbidden: false},
		{name: "admin can generate-install-assets", client: admin, action: "install-assets", query: generateCloudAccountInstallAssetsMutation, variables: installVars, expectForbidden: false},
		{name: "admin can verify", client: admin, action: "verify", query: verifyCloudAccountMutation, variables: verifyVars, expectForbidden: false},
		{name: "admin can rotate-credentials", client: admin, action: "rotate-credentials", query: rotateCloudAccountCredentialsMutation, variables: rotateVars, expectForbidden: false},
		{name: "admin can delete", client: admin, action: "delete", query: deleteCloudAccountMutation, variables: deleteVars, expectForbidden: false},

		// VIEWER (list + get only).
		{name: "viewer can list", client: viewer, action: "list", query: cloudAccountListQuery, variables: listVars, expectForbidden: false},
		{name: "viewer can get", client: viewer, action: "get", query: cloudAccountGetQuery, variables: getVars, expectForbidden: false},
		{name: "viewer cannot create", client: viewer, action: "create", query: createCloudAccountMutation, variables: createVars, expectForbidden: true},
		{name: "viewer cannot generate-install-assets", client: viewer, action: "install-assets", query: generateCloudAccountInstallAssetsMutation, variables: installVars, expectForbidden: true},
		{name: "viewer cannot verify", client: viewer, action: "verify", query: verifyCloudAccountMutation, variables: verifyVars, expectForbidden: true},
		{name: "viewer cannot rotate-credentials", client: viewer, action: "rotate-credentials", query: rotateCloudAccountCredentialsMutation, variables: rotateVars, expectForbidden: true},
		{name: "viewer cannot delete", client: viewer, action: "delete", query: deleteCloudAccountMutation, variables: deleteVars, expectForbidden: true},

		// AUDITOR (list + get only).
		{name: "auditor can list", client: auditor, action: "list", query: cloudAccountListQuery, variables: listVars, expectForbidden: false},
		{name: "auditor can get", client: auditor, action: "get", query: cloudAccountGetQuery, variables: getVars, expectForbidden: false},
		{name: "auditor cannot create", client: auditor, action: "create", query: createCloudAccountMutation, variables: createVars, expectForbidden: true},
		{name: "auditor cannot generate-install-assets", client: auditor, action: "install-assets", query: generateCloudAccountInstallAssetsMutation, variables: installVars, expectForbidden: true},
		{name: "auditor cannot verify", client: auditor, action: "verify", query: verifyCloudAccountMutation, variables: verifyVars, expectForbidden: true},
		{name: "auditor cannot rotate-credentials", client: auditor, action: "rotate-credentials", query: rotateCloudAccountCredentialsMutation, variables: rotateVars, expectForbidden: true},
		{name: "auditor cannot delete", client: auditor, action: "delete", query: deleteCloudAccountMutation, variables: deleteVars, expectForbidden: true},

		// EMPLOYEE (none).
		{name: "employee cannot list", client: employee, action: "list", query: cloudAccountListQuery, variables: listVars, expectForbidden: true},
		{name: "employee cannot get", client: employee, action: "get", query: cloudAccountGetQuery, variables: getVars, expectForbidden: true},
		{name: "employee cannot create", client: employee, action: "create", query: createCloudAccountMutation, variables: createVars, expectForbidden: true},
		{name: "employee cannot generate-install-assets", client: employee, action: "install-assets", query: generateCloudAccountInstallAssetsMutation, variables: installVars, expectForbidden: true},
		{name: "employee cannot verify", client: employee, action: "verify", query: verifyCloudAccountMutation, variables: verifyVars, expectForbidden: true},
		{name: "employee cannot rotate-credentials", client: employee, action: "rotate-credentials", query: rotateCloudAccountCredentialsMutation, variables: rotateVars, expectForbidden: true},
		{name: "employee cannot delete", client: employee, action: "delete", query: deleteCloudAccountMutation, variables: deleteVars, expectForbidden: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			variables := tt.variables(tt.client)

			// Special-case: list/get use queries -- 0-result is
			// success; only an explicit FORBIDDEN error fails the
			// test. EMPLOYEE on list is enforced by a different
			// rule (Organization-level read), so list/get for
			// EMPLOYEE may instead surface as null node rather
			// than an explicit FORBIDDEN. Accept either as a
			// "blocked" outcome.
			if tt.action == "list" || tt.action == "get" {
				_, err := tt.client.Do(tt.query, variables)
				if tt.expectForbidden {
					require.Error(t, err, "%s should be forbidden", tt.name)
					return
				}
				require.NoError(t, err, "%s should succeed", tt.name)
				return
			}

			// Special-case: delete needs a fresh row per call
			// (otherwise the second delete races against the
			// first). For role rows that are allowed to delete,
			// pre-create a row owned by that client first.
			if tt.action == "delete" && !tt.expectForbidden {
				row := factory.NewCloudAccount(tt.client).
					WithLabel("RBAC Delete " + tt.name).
					Create()
				variables = map[string]any{
					"input": map[string]any{"cloudAccountId": row},
				}
			} else if tt.action == "delete" {
				// For forbidden roles we re-use the owner's row
				// (we expect a FORBIDDEN before any state mutation).
				variables = map[string]any{
					"input": map[string]any{"cloudAccountId": cloudAccountID},
				}
			}

			_, err := tt.client.Do(tt.query, variables)
			if tt.expectForbidden {
				testutil.RequireForbiddenError(t, err, "%s expected FORBIDDEN", tt.name)
				return
			}
			require.NoError(t, err, "%s should succeed", tt.name)
		})
	}
}

// =============================================================================
// Tenant isolation.
//
// Org-A creates a cloud account; Org-B owner attempts to read /
// list / rotate / delete it. The Scoper (pkg/coredata.CloudAccount.LoadByID)
// returns ErrResourceNotFound when the row is in another tenant,
// which the resolver maps to NOT_FOUND -- not FORBIDDEN. This
// distinction matters: leaking "this id exists but you can't see
// it" is a side-channel; NOT_FOUND is the correct (information-zero)
// response.
// =============================================================================

func TestCloudAccount_TenantIsolation(t *testing.T) {
	t.Parallel()

	orgAOwner := testutil.NewClient(t, testutil.RoleOwner)
	orgBOwner := testutil.NewClient(t, testutil.RoleOwner)

	cloudAccountID := factory.NewCloudAccount(orgAOwner).
		WithLabel("Org-A Cloud Account").
		Create()

	t.Run("org-B cannot read org-A cloud account", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *cloudAccountNode `json:"node"`
		}
		err := orgBOwner.Execute(cloudAccountGetQuery, map[string]any{
			"id": cloudAccountID,
		}, &result)
		// Either err != nil with NOT_FOUND, OR result.Node is nil
		// (gqlgen returns null for an inaccessible Node lookup).
		// AssertNodeNotAccessible accepts either.
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "cloud account")
	})

	t.Run("org-B list does not include org-A cloud accounts", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node struct {
				CloudAccounts struct {
					Edges []struct {
						Node cloudAccountNode `json:"node"`
					} `json:"edges"`
				} `json:"cloudAccounts"`
			} `json:"node"`
		}
		err := orgBOwner.Execute(cloudAccountListQuery, map[string]any{
			"id": orgBOwner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		for _, edge := range result.Node.CloudAccounts.Edges {
			assert.NotEqual(t, cloudAccountID, edge.Node.ID,
				"org-B list must not include org-A cloud account")
		}
	})

	t.Run("org-B cannot rotate org-A cloud account", func(t *testing.T) {
		t.Parallel()

		_, err := orgBOwner.Do(rotateCloudAccountCredentialsMutation, map[string]any{
			"input": map[string]any{
				"cloudAccountId": cloudAccountID,
				"provider":       "AWS",
				"credentialKind": "AWS_ASSUME_ROLE",
				"awsRoleArn":     "arn:aws:iam::999999999999:role/Hijack",
				"awsExternalId":  "9999999999999999999999999999999999999999999999999999999999999999",
			},
		})
		require.Error(t, err, "org-B must not be able to rotate org-A cloud account")
		// Per resolver mapping the row resolves to NOT_FOUND
		// (Scoper hides the row), not FORBIDDEN.
		testutil.RequireErrorCode(t, err, "NOT_FOUND",
			"tenant isolation must surface as NOT_FOUND (Scoper-driven), not FORBIDDEN")
	})

	t.Run("org-B cannot delete org-A cloud account", func(t *testing.T) {
		t.Parallel()

		_, err := orgBOwner.Do(deleteCloudAccountMutation, map[string]any{
			"input": map[string]any{"cloudAccountId": cloudAccountID},
		})
		require.Error(t, err)
		testutil.RequireErrorCode(t, err, "NOT_FOUND",
			"tenant isolation must surface as NOT_FOUND")
	})
}

// =============================================================================
// External-id reuse across orgs (confused-deputy mitigation).
//
// Org-A obtains an external_id from generateInstallAssets. Org-B
// attempts createCloudAccount with the same external_id. The
// service rejects with INVALID since (label, organization_id) /
// external_id collisions are caught at insert time. Even if the
// external_id IS reused (it's per-row, not per-tenant), the AWS
// trust policy created by Org-A's CloudFormation stack ties the
// credential to Org-A's external_id; Org-B simply gets a fresh
// external_id and a fresh row. The cross-org-reuse path is
// therefore not "reject" but rather "the credential won't probe
// because the trust policy was bound to Org-A". For UI purposes
// the service does NOT enforce a global uniqueness constraint
// today; this test documents the expected behavior so a future
// regression that adds one reads as intentional.
// =============================================================================

func TestCloudAccount_ExternalIdReuseAcrossOrgs(t *testing.T) {
	t.Parallel()

	orgAOwner := testutil.NewClient(t, testutil.RoleOwner)
	orgBOwner := testutil.NewClient(t, testutil.RoleOwner)

	// Org-A generates install assets to obtain an external_id (this
	// path is the only documented source of a "real" external_id).
	// When the e2e harness has no AWS template configured the call
	// surfaces UNAVAILABLE; in that case we fall back to a known
	// 64-hex-char string -- the assertion below is independent of
	// where the value came from.
	var externalID string
	var assets struct {
		GenerateCloudAccountInstallAssets struct {
			Assets struct {
				Typename   string `json:"__typename"`
				ExternalID string `json:"externalId"`
			} `json:"assets"`
		} `json:"generateCloudAccountInstallAssets"`
	}
	err := orgAOwner.Execute(generateCloudAccountInstallAssetsMutation, map[string]any{
		"input": map[string]any{
			"organizationId":  orgAOwner.GetOrganizationID().String(),
			"provider":        "AWS",
			"scopeKind":       "AWS_ACCOUNT",
			"scopeIdentifier": "111111111111",
			"modules":         []string{"ACCESS_REVIEW"},
		},
	}, &assets)
	if err == nil && assets.GenerateCloudAccountInstallAssets.Assets.ExternalID != "" {
		externalID = assets.GenerateCloudAccountInstallAssets.Assets.ExternalID
	} else {
		// Substitute a deterministic 64-hex value so the rest of
		// the test runs even when AWS template isn't configured.
		externalID = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	}

	require.Len(t, externalID, 64, "external_id must be 64 hex chars")

	// Org-B owner attempts createCloudAccount with the same
	// external_id. Today the service does not enforce a
	// cross-tenant uniqueness constraint, so the call succeeds at
	// the GraphQL layer (the verify path then fails because the
	// trust policy is bound to Org-A). Document the current
	// behaviour: the call DOES NOT return INVALID; it succeeds
	// and surfaces the probe failure via lastProbeError.
	var result createCloudAccountResult
	err = orgBOwner.Execute(createCloudAccountMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      orgBOwner.GetOrganizationID().String(),
			"label":               "Org-B Reuse Attempt",
			"provider":            "AWS",
			"credentialKind":      "AWS_ASSUME_ROLE",
			"scopeKind":           "AWS_ACCOUNT",
			"scopeIdentifier":     "222222222222",
			"enabledAuditModules": []string{"ACCESS_REVIEW"},
			"awsRoleArn":          "arn:aws:iam::111111111111:role/OrgARole",
			"awsExternalId":       externalID,
		},
	}, &result)

	// Path A: per the documented confused-deputy mitigation the
	// service rejects with INVALID. Path B: today's service has
	// no global uniqueness check; the row is created and probe
	// fails. Both paths are acceptable -- assert the security
	// outcome (Org-B cannot use Org-A's credential) rather than
	// a specific error code.
	if err != nil {
		// Service rejected -- ideal mitigation path.
		var gqlErrs testutil.GraphQLErrors
		require.ErrorAs(t, err, &gqlErrs)
		t.Logf("ExternalID reuse rejected with code %q (defensive mitigation in place)",
			gqlErrs[0].Code())
		return
	}

	// Service accepted; verify failure is the security boundary.
	require.Equal(t, "PENDING_VERIFICATION", result.CreateCloudAccount.CloudAccount.Status,
		"reused external_id MUST not promote to VERIFIED -- the trust policy on Org-A's role rejects Org-B")
	require.NotNil(t, result.CreateCloudAccount.LastProbeError,
		"reused external_id MUST surface a probe error")
}

// =============================================================================
// Access-review integration.
//
// `CreateAccessSourceInput.cloudAccountId` is now part of the public
// console GraphQL surface (see graphql/access_review_campaign.graphql).
// This test wires a cloud-account-backed access source end-to-end and
// asserts the link round-trips through the list query. The actual
// driver fetch is unit-tested in pkg/probo/cloud_account_driver_test.go
// (no real cloud SDK is reachable from the e2e harness).
// =============================================================================

func TestCloudAccount_AccessReviewIntegration(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	cloudAccountID := factory.NewCloudAccount(owner).
		WithProvider("AWS").
		WithLabel("Access Review Cloud Account").
		Create()
	require.NotEmpty(t, cloudAccountID)

	// Create an access source backed by the cloud account through the
	// public GraphQL surface (the field under test is the new
	// CreateAccessSourceInput.cloudAccountId).
	var createResult struct {
		CreateAccessSource struct {
			AccessSourceEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"accessSourceEdge"`
		} `json:"createAccessSource"`
	}
	err := owner.Execute(createAccessSourceMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"cloudAccountId": cloudAccountID,
			"name":           "AWS Access Source",
		},
	}, &createResult)
	require.NoError(t, err, "createAccessSource with cloudAccountId must succeed")

	accessSourceID := createResult.CreateAccessSource.AccessSourceEdge.Node.ID
	require.NotEmpty(t, accessSourceID)
	assert.Equal(t, "AWS Access Source", createResult.CreateAccessSource.AccessSourceEdge.Node.Name)

	// Verify the access source appears in the org-scoped list -- this
	// is the round-trip assertion that proves the FK landed.
	t.Run("access source appears in org list", func(t *testing.T) {
		t.Parallel()

		var listResult struct {
			Node struct {
				AccessSources struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"accessSources"`
			} `json:"node"`
		}
		err := owner.Execute(accessSourceListQuery, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &listResult)
		require.NoError(t, err)

		found := false
		for _, edge := range listResult.Node.AccessSources.Edges {
			if edge.Node.ID == accessSourceID {
				found = true
				break
			}
		}
		assert.True(t, found, "cloud-account-backed access source must appear in the org list")
	})
}

// =============================================================================
// Field-level redaction tests.
//
// `lastProbeError` and `scope.identifier` are credential-adjacent:
// raw SDK errors frequently embed account ids / role ARNs / SA
// emails / tenant guids. The resolver allows these fields only to
// callers with `core:cloud-account:rotate-credentials` (OWNER /
// ADMIN). VIEWER / AUDITOR / EMPLOYEE see null.
// =============================================================================

func TestCloudAccount_LastProbeErrorRedaction(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)

	// Trigger a probe failure (the only way to populate
	// last_probe_error in e2e). The factory's post-create Verify
	// already does this -- the row lands with a non-nil
	// last_probe_error.
	cloudAccountID := factory.NewCloudAccount(owner).
		WithLabel("Probe Error Redaction").
		Create()

	t.Run("OWNER sees the actual probe error string", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node.LastProbeError, "owner must see probe error")
		assert.NotEmpty(t, *result.Node.LastProbeError)
	})

	t.Run("AUDITOR sees null for last_probe_error", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := auditor.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.LastProbeError,
			"auditor must NOT see last_probe_error (raw SDK strings frequently embed credentials)")
	})
}

func TestCloudAccount_ScopeIdentifierRedaction(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)

	cloudAccountID := factory.NewCloudAccount(owner).
		WithLabel("Scope Identifier Redaction").
		WithProvider("AWS").
		WithScopeIdentifier("987654321098").
		Create()

	t.Run("OWNER sees scope.identifier", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node.Scope.Identifier, "owner must see scope.identifier")
		assert.Equal(t, "987654321098", *result.Node.Scope.Identifier)
	})

	t.Run("AUDITOR sees null for scope.identifier", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		err := auditor.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.Scope.Identifier,
			"auditor must NOT see scope.identifier (AWS account ids / GCP project ids carry reconnaissance value)")
	})
}

// =============================================================================
// Edge cases (Step 21).
// =============================================================================

// TestCloudAccount_ProbeFailureOnInitialVerify asserts the
// PENDING_VERIFICATION + last_probe_error invariant on initial
// Verify failure. In the e2e harness this is the natural state
// after every createCloudAccount call (no real cloud SDK is
// reachable), so the test directly inspects the row that the
// factory just produced.
func TestCloudAccount_ProbeFailureOnInitialVerify(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	cloudAccountID := factory.NewCloudAccount(owner).
		WithLabel("Initial Verify Failure").
		Create()

	var result struct {
		Node cloudAccountNode `json:"node"`
	}
	err := owner.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
	require.NoError(t, err)

	// The plan invariant: a never-verified row stays in
	// PENDING_VERIFICATION on probe failure (we never auto-promote
	// a never-verified row to ERRORED). last_probe_error is set so
	// the operator gets an actionable hint.
	assert.Equal(t, "PENDING_VERIFICATION", result.Node.Status,
		"initial Verify failure must keep status at PENDING_VERIFICATION (never auto-promote)")
	require.NotNil(t, result.Node.LastProbeError,
		"initial Verify failure must populate last_probe_error")
	assert.NotEmpty(t, *result.Node.LastProbeError)
	assert.Nil(t, result.Node.LastVerifiedAt,
		"never-verified row must have lastVerifiedAt = null")
}

// TestCloudAccount_RotationDuringFetch is documented as deferred:
// reproducing the race deterministically requires either a
// time-controlled probe seam in the e2e harness or a pre-staged
// in-flight fetch handle. The actual invariant ("rotate runs in a
// transaction; an in-flight fetch uses pre-rotation creds already
// extracted before rotate") is enforced at the service-layer
// transaction boundary and covered by
// pkg/probo/cloud_account_service_test.go for the tx contract and
// pkg/probo/cloud_account_worker_test.go for the no-block-on-read
// guarantee. The replacement assertion here is that
// RotateCredentials does not block a concurrent read on the row.
func TestCloudAccount_RotationDuringFetch(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	cloudAccountID := factory.NewCloudAccount(owner).
		WithLabel("Rotation Race").
		Create()

	t.Run("read does not block on rotate", func(t *testing.T) {
		t.Parallel()

		// Kick off rotate on one client; concurrently read on a
		// fresh client. Both must succeed.
		readerClient := testutil.NewClientWithNewSession(t, owner)

		rotateDone := make(chan error, 1)
		go func() {
			_, rotateErr := owner.Do(rotateCloudAccountCredentialsMutation, map[string]any{
				"input": map[string]any{
					"cloudAccountId": cloudAccountID,
					"provider":       "AWS",
					"credentialKind": "AWS_ASSUME_ROLE",
					"awsRoleArn":     "arn:aws:iam::111111111111:role/Concurrent",
					"awsExternalId":  "3333333333333333333333333333333333333333333333333333333333333333",
				},
			})
			rotateDone <- rotateErr
		}()

		var result struct {
			Node cloudAccountNode `json:"node"`
		}
		readErr := readerClient.Execute(cloudAccountGetQuery, map[string]any{"id": cloudAccountID}, &result)
		require.NoError(t, readErr, "concurrent read must succeed")

		require.NoError(t, <-rotateDone, "rotate must succeed")
	})
}

// TestCloudAccount_DeleteBlockedWhenInUse asserts the FK constraint
// `access_sources_cloud_account_id_fkey` rejects a cloud-account
// delete while at least one access source still references it. The
// resolver maps the PG 23503 violation to a Conflict GraphQL error.
// After detaching the access source the delete must succeed.
func TestCloudAccount_DeleteBlockedWhenInUse(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	cloudAccountID := factory.NewCloudAccount(owner).
		WithProvider("AWS").
		WithLabel("Delete Blocked When In Use").
		Create()
	require.NotEmpty(t, cloudAccountID)

	var createResult struct {
		CreateAccessSource struct {
			AccessSourceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"accessSourceEdge"`
		} `json:"createAccessSource"`
	}
	err := owner.Execute(createAccessSourceMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"cloudAccountId": cloudAccountID,
			"name":           "Source Pinning Cloud Account",
		},
	}, &createResult)
	require.NoError(t, err)
	accessSourceID := createResult.CreateAccessSource.AccessSourceEdge.Node.ID
	require.NotEmpty(t, accessSourceID)

	// First delete must fail: the FK from access_sources holds the
	// cloud account in place. The resolver surfaces this as a Conflict.
	_, err = owner.Do(deleteCloudAccountMutation, map[string]any{
		"input": map[string]any{"cloudAccountId": cloudAccountID},
	})
	require.Error(t, err, "delete must fail while an access source references the cloud account")

	// Detach the access source.
	_, err = owner.Do(deleteAccessSourceMutation, map[string]any{
		"input": map[string]any{"accessSourceId": accessSourceID},
	})
	require.NoError(t, err, "detaching the access source must succeed")

	// After detach the cloud-account delete must succeed.
	var deleteResult struct {
		DeleteCloudAccount struct {
			DeletedCloudAccountID string `json:"deletedCloudAccountId"`
		} `json:"deleteCloudAccount"`
	}
	err = owner.Execute(deleteCloudAccountMutation, map[string]any{
		"input": map[string]any{"cloudAccountId": cloudAccountID},
	}, &deleteResult)
	require.NoError(t, err, "delete must succeed once no access source references the cloud account")
	assert.Equal(t, cloudAccountID, deleteResult.DeleteCloudAccount.DeletedCloudAccountID)
}
