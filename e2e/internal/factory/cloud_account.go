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

package factory

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// Default mock credentials used by the cloud-account factory. None
// of these are valid against a real cloud, which is the point: the
// e2e binary has no AWS / GCP / Azure SDK injection seam, so the
// post-create Verify call deterministically fails and the row stays
// in PENDING_VERIFICATION with last_probe_error set. Tests that
// need to assert post-VERIFIED state should drive Verify with a
// stub registry in unit tests (see pkg/probo/cloud_account_*_test.go),
// not here.
const (
	mockAWSRoleARN     = "arn:aws:iam::123456789012:role/ProboCloudScannerE2E"
	mockAWSAccountID   = "123456789012"
	mockGCPProjectID   = "probo-e2e-mock-project"
	mockGCPOrgID       = "987654321098"
	mockAzureTenant    = "00000000-0000-0000-0000-000000000001"
	mockAzureClient    = "00000000-0000-0000-0000-000000000002"
	mockAzureMG        = "00000000-0000-0000-0000-000000000003"
	mockAzureSubscript = "00000000-0000-0000-0000-000000000004"
)

// randomExternalID mints a fresh AWS external_id per factory call.
// Cross-org reuse defeat means a hard-coded constant would only let
// the first test in the suite create an AWS row; subsequent calls
// would fail the second-org guard. 32 bytes of crypto/rand → 64 hex
// chars matches the production format set by
// pkg/cloudaccount.GenerateAWSExternalID.
func randomExternalID() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// CreateCloudAccount creates a cloud-account row through the public
// GraphQL API (`createCloudAccount`). The factory defaults to AWS
// with `AWS_ACCOUNT` scope and `AWS_ASSUME_ROLE` credentials so a
// caller that just wants "any cloud account" can write
// `factory.CreateCloudAccount(c)` and move on.
//
// Supported attrs (defaults shown):
//
//	"label"           string  // SafeName("Cloud Account")
//	"provider"        string  // "AWS" | "GCP" | "AZURE"
//	"credentialKind"  string  // derived from provider when omitted
//	"scopeKind"       string  // derived from provider when omitted
//	"scopeIdentifier" string  // mock id matching scopeKind
//	"awsRoleArn"      string  // mockAWSRoleARN
//	"awsExternalId"   string  // generated per call by randomExternalID()
//	"azureTenantId"   string  // mockAzureTenant
//	"azureClientId"   string  // mockAzureClient
//	"gcpProjectId"    string  // mockGCPProjectID
//	"gcpOrganizationId" string // mockGCPOrgID
//	"azureSubscriptionId" string
//	"azureManagementGroupId" string
//
// Verify failure note: the post-create Verify always fails in e2e
// (no real cloud SDK is reachable from the test harness), so the
// returned row is in PENDING_VERIFICATION with `last_probe_error`
// populated. This is intentional. Tests that need a VERIFIED row
// should run the registry seam in unit tests; e2e covers the
// API-level lifecycle and RBAC / tenant isolation only.
func CreateCloudAccount(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	provider := a.getString("provider", "AWS")

	input := buildCreateCloudAccountInput(c, a, provider)

	const query = `
		mutation($input: CreateCloudAccountInput!) {
			createCloudAccount(input: $input) {
				cloudAccount { id }
			}
		}
	`

	var result struct {
		CreateCloudAccount struct {
			CloudAccount struct {
				ID string `json:"id"`
			} `json:"cloudAccount"`
		} `json:"createCloudAccount"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createCloudAccount mutation failed")

	return result.CreateCloudAccount.CloudAccount.ID
}

// buildCreateCloudAccountInput assembles the GraphQL input map per
// provider. Centralised so test sites stay short.
func buildCreateCloudAccountInput(c *testutil.Client, a Attrs, provider string) map[string]any {
	input := map[string]any{
		"organizationId":      c.GetOrganizationID().String(),
		"label":               a.getString("label", SafeName("Cloud Account")),
		"provider":            provider,
		"enabledAuditModules": []string{"ACCESS_REVIEW"},
	}

	switch provider {
	case "AWS":
		input["credentialKind"] = a.getString("credentialKind", "AWS_ASSUME_ROLE")
		input["scopeKind"] = a.getString("scopeKind", "AWS_ACCOUNT")
		input["scopeIdentifier"] = a.getString("scopeIdentifier", mockAWSAccountID)
		input["awsRoleArn"] = a.getString("awsRoleArn", mockAWSRoleARN)
		input["awsExternalId"] = a.getString("awsExternalId", randomExternalID())

	case "GCP":
		input["credentialKind"] = a.getString("credentialKind", "GCP_SERVICE_ACCOUNT_KEY")
		input["scopeKind"] = a.getString("scopeKind", "GCP_PROJECT")
		input["scopeIdentifier"] = a.getString("scopeIdentifier", mockGCPProjectID)
		if v := a.getStringPtr("gcpProjectId"); v != nil {
			input["gcpProjectId"] = *v
		} else {
			input["gcpProjectId"] = mockGCPProjectID
		}
		if v := a.getStringPtr("gcpOrganizationId"); v != nil {
			input["gcpOrganizationId"] = *v
		}

	case "AZURE":
		input["credentialKind"] = a.getString("credentialKind", "AZURE_CLIENT_SECRET")
		input["scopeKind"] = a.getString("scopeKind", "AZURE_MANAGEMENT_GROUP")
		input["scopeIdentifier"] = a.getString("scopeIdentifier", mockAzureMG)
		input["azureTenantId"] = a.getString("azureTenantId", mockAzureTenant)
		input["azureClientId"] = a.getString("azureClientId", mockAzureClient)
		if v := a.getStringPtr("azureSubscriptionId"); v != nil {
			input["azureSubscriptionId"] = *v
		}
		if v := a.getStringPtr("azureManagementGroupId"); v != nil {
			input["azureManagementGroupId"] = *v
		} else {
			input["azureManagementGroupId"] = mockAzureMG
		}
	}

	return input
}

// CloudAccountBuilder is the fluent builder companion for
// CreateCloudAccount, matching the project's `NewVendor(...).WithX(...).Create()`
// shape.
type CloudAccountBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewCloudAccount(c *testutil.Client) *CloudAccountBuilder {
	return &CloudAccountBuilder{client: c, attrs: Attrs{}}
}

func (b *CloudAccountBuilder) WithProvider(provider string) *CloudAccountBuilder {
	b.attrs["provider"] = provider
	return b
}

func (b *CloudAccountBuilder) WithLabel(label string) *CloudAccountBuilder {
	b.attrs["label"] = label
	return b
}

func (b *CloudAccountBuilder) WithScopeKind(scopeKind string) *CloudAccountBuilder {
	b.attrs["scopeKind"] = scopeKind
	return b
}

func (b *CloudAccountBuilder) WithScopeIdentifier(identifier string) *CloudAccountBuilder {
	b.attrs["scopeIdentifier"] = identifier
	return b
}

func (b *CloudAccountBuilder) WithAWSRoleARN(arn string) *CloudAccountBuilder {
	b.attrs["awsRoleArn"] = arn
	return b
}

func (b *CloudAccountBuilder) WithAWSExternalID(id string) *CloudAccountBuilder {
	b.attrs["awsExternalId"] = id
	return b
}

// WithStatus is accepted for API symmetry with the original spec
// but has no effect in the current e2e harness: the e2e binary
// has no DB-side seam to flip status to VERIFIED, and no real
// cloud SDK injection point. Tests that need a VERIFIED row should
// rely on registry-level unit tests (pkg/probo/cloud_account_*_test.go).
// The option is preserved so a later harness gains a no-rewrite
// migration path.
func (b *CloudAccountBuilder) WithStatus(status string) *CloudAccountBuilder {
	b.attrs["__seed_status"] = status
	return b
}

func (b *CloudAccountBuilder) Create() string {
	return CreateCloudAccount(b.client, b.attrs)
}
