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

package mcp_v1

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/probo"
)

// mockAzureGuide returns a deterministic AzureInstallGuide fixture
// covering the empty-Code branch (step 0) and the populated-Code
// branch (step 1) of the converter under test.
func mockAzureGuide() *cloudaccount.AzureInstallGuide {
	return &cloudaccount.AzureInstallGuide{
		Steps: []cloudaccount.AzureInstallStep{
			{
				Title: "Create the App Registration",
				Body:  "narrative-only step",
			},
			{
				Title: "Assign the Reader role",
				Body:  "with CLI snippet",
				Code:  "az role assignment create",
			},
		},
		RequiredRBACRoles:        []string{"Reader"},
		RequiredGraphPermissions: []string{"Directory.Read.All"},
	}
}

// MCP resolver tests are hard to drive end-to-end without a real
// iam.Service (MustAuthorize calls into the real Authorizer; there is
// no interface seam to inject a fake). Instead, this file pins the
// observable contracts that other agents and reviewers care about:
//
//  1. the wire-level OpenAPI specification served to MCP clients
//     (specification.yaml) and the generated Go input types contain
//     no secret-credential fields for create/rotate -- secrets must
//     travel exclusively via the multipart upload endpoint;
//
//  2. each MCP resolver method dispatches to MustAuthorize with the
//     correct ActionCloudAccount* constant (verified by source-text
//     scan of the generated resolvers file -- a brittle but cheap
//     guard against a stale action wiring);
//
//  3. the IAM policy set permits AUDITOR list+get, denies AUDITOR on
//     mutating actions, and denies EMPLOYEE on every cloud-account
//     action.
//
// (1) and (3) are hermetic; (2) trades a tiny bit of brittleness for
// hermeticity -- no DB, no panics.

// --- (1) Specification contract: secrets are NOT in create/rotate ---

type specComponentsSchemas struct {
	Components struct {
		Schemas map[string]struct {
			Type       string                    `yaml:"type"`
			Required   []string                  `yaml:"required"`
			Properties map[string]map[string]any `yaml:"properties"`
		} `yaml:"schemas"`
	} `yaml:"components"`
}

type specToolsDoc struct {
	Tools []struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Hints       struct {
			Readonly    *bool `yaml:"readonly"`
			Idempotent  *bool `yaml:"idempotent"`
			Destructive *bool `yaml:"destructive"`
		} `yaml:"hints"`
		InputSchema struct {
			Ref string `yaml:"$ref"`
		} `yaml:"inputSchema"`
		OutputSchema struct {
			Ref string `yaml:"$ref"`
		} `yaml:"outputSchema"`
	} `yaml:"tools"`
}

func loadSpec(t *testing.T) ([]byte, *specComponentsSchemas, *specToolsDoc) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	path := filepath.Join(wd, "specification.yaml")
	raw, err := os.ReadFile(path)
	require.NoError(t, err, "specification.yaml must be readable from the test working directory")

	var schemas specComponentsSchemas
	require.NoError(t, yaml.Unmarshal(raw, &schemas))

	var tools specToolsDoc
	require.NoError(t, yaml.Unmarshal(raw, &tools))

	return raw, &schemas, &tools
}

func TestCloudAccountMCP_SecretFieldsAbsentFromInputSchemas(t *testing.T) {
	t.Parallel()

	_, schemas, _ := loadSpec(t)

	cases := []struct {
		schema string
	}{
		{schema: "CreateCloudAccountMCPInput"},
		{schema: "RotateCloudAccountCredentialsMCPInput"},
	}

	// Field names that, if ever introduced, would mean the MCP API
	// is leaking secret credential bytes through tool variables.
	// Both snake_case (yaml) and camelCase (Go) are checked.
	forbidden := []string{
		"gcp_service_account_key", "gcpServiceAccountKey",
		"azure_client_secret", "azureClientSecret",
		"client_secret", "clientSecret",
		"service_account_json", "serviceAccountJson", "serviceAccountJSON",
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.schema, func(t *testing.T) {
			t.Parallel()

			s, ok := schemas.Components.Schemas[tc.schema]
			require.Truef(t, ok, "schema %q must exist in specification.yaml", tc.schema)
			for _, banned := range forbidden {
				_, present := s.Properties[banned]
				assert.Falsef(t,
					present,
					"%s must NOT carry the %q property -- secret credential bytes travel via the upload endpoint",
					tc.schema, banned,
				)
			}
		})
	}
}

func TestCloudAccountMCP_ToolListPinsExpectedDispatch(t *testing.T) {
	t.Parallel()

	_, _, tools := loadSpec(t)

	want := map[string]string{
		"cloud_account_list":                    "#/components/schemas/ListCloudAccountsInput",
		"cloud_account_get":                     "#/components/schemas/GetCloudAccountInput",
		"cloud_account_install_assets_generate": "#/components/schemas/GenerateCloudAccountInstallAssetsMCPInput",
		"cloud_account_create":                  "#/components/schemas/CreateCloudAccountMCPInput",
		"cloud_account_verify":                  "#/components/schemas/VerifyCloudAccountMCPInput",
		"cloud_account_rotate_credentials":      "#/components/schemas/RotateCloudAccountCredentialsMCPInput",
		"cloud_account_delete":                  "#/components/schemas/DeleteCloudAccountMCPInput",
	}

	got := make(map[string]string)
	for _, tool := range tools.Tools {
		if _, ok := want[tool.Name]; ok {
			got[tool.Name] = tool.InputSchema.Ref
		}
	}

	for name, ref := range want {
		assert.Equalf(t, ref, got[name], "tool %q must declare inputSchema=%s", name, ref)
	}
}

// --- (2) Resolver-source contract: ActionCloudAccount* per tool ---

func TestCloudAccountMCP_ResolverActionWiring(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	src, err := os.ReadFile(filepath.Join(wd, "schema.resolvers.go"))
	require.NoError(t, err)
	body := string(src)

	cases := []struct {
		method string
		action string
	}{
		{method: "CloudAccountListTool", action: "ActionCloudAccountList"},
		{method: "CloudAccountGetTool", action: "ActionCloudAccountGet"},
		{method: "CloudAccountInstallAssetsGenerateTool", action: "ActionCloudAccountGenerateInstallAssets"},
		{method: "CloudAccountCreateTool", action: "ActionCloudAccountCreate"},
		{method: "CloudAccountVerifyTool", action: "ActionCloudAccountVerify"},
		{method: "CloudAccountRotateCredentialsTool", action: "ActionCloudAccountRotateCredentials"},
		{method: "CloudAccountDeleteTool", action: "ActionCloudAccountDelete"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()

			// Locate `func (r *Resolver) <method>(...) ...` and capture
			// its body up to the first `^}` line. The resolver must
			// invoke MustAuthorize with the expected action constant
			// somewhere in that body.
			pattern := regexp.MustCompile(
				`(?ms)func \(r \*Resolver\) ` + regexp.QuoteMeta(tc.method) + `\b.*?\n\}\n`,
			)
			match := pattern.FindString(body)
			require.NotEmptyf(t, match, "resolver method %q must exist in schema.resolvers.go", tc.method)

			expected := "probo." + tc.action
			assert.Containsf(t, match, "r.MustAuthorize(", "method %q must call MustAuthorize", tc.method)
			assert.Containsf(t, match, expected, "method %q must reference %s", tc.method, expected)
		})
	}
}

// --- (3) Policy contract: AUDITOR & EMPLOYEE matrix ---

// statementGrants returns true if the policy contains an Allow statement
// whose Actions slice includes the supplied action.
func statementGrants(p *policy.Policy, action string) bool {
	if p == nil {
		return false
	}
	for _, s := range p.Statements {
		if s.Effect != policy.EffectAllow {
			continue
		}
		for _, a := range s.Actions {
			if a == action {
				return true
			}
		}
	}
	return false
}

func TestCloudAccountMCP_AuditorPolicyMatrix(t *testing.T) {
	t.Parallel()

	allowed := []string{
		probo.ActionCloudAccountList,
		probo.ActionCloudAccountGet,
	}
	forbidden := []string{
		probo.ActionCloudAccountGenerateInstallAssets,
		probo.ActionCloudAccountCreate,
		probo.ActionCloudAccountVerify,
		probo.ActionCloudAccountRotateCredentials,
		probo.ActionCloudAccountDelete,
	}

	for _, a := range allowed {
		a := a
		t.Run("auditor allowed: "+a, func(t *testing.T) {
			t.Parallel()
			assert.True(t,
				statementGrants(probo.AuditorPolicy, a),
				"AUDITOR must allow %s",
				a,
			)
		})
	}

	for _, a := range forbidden {
		a := a
		t.Run("auditor forbidden: "+a, func(t *testing.T) {
			t.Parallel()
			assert.False(t,
				statementGrants(probo.AuditorPolicy, a),
				"AUDITOR must NOT allow %s -- mutating cloud-account action",
				a,
			)
		})
	}
}

func TestCloudAccountMCP_EmployeePolicyMatrix(t *testing.T) {
	t.Parallel()

	all := []string{
		probo.ActionCloudAccountList,
		probo.ActionCloudAccountGet,
		probo.ActionCloudAccountGenerateInstallAssets,
		probo.ActionCloudAccountCreate,
		probo.ActionCloudAccountVerify,
		probo.ActionCloudAccountRotateCredentials,
		probo.ActionCloudAccountDelete,
	}

	for _, a := range all {
		a := a
		t.Run("employee forbidden: "+a, func(t *testing.T) {
			t.Parallel()
			assert.False(t,
				statementGrants(probo.EmployeePolicy, a),
				"EMPLOYEE must NOT allow any cloud-account action; got %s",
				a,
			)
		})
	}
}

// --- adapter helper sanity checks (pure unit) ---

func TestNewMCPAzureInstallAssets_SkipsEmptyCode(t *testing.T) {
	t.Parallel()

	// newMCPAzureInstallAssets is the typed converter used by the
	// install-assets resolver path. The Azure step model allows Code
	// to be empty for narrative-only steps; the converter must NOT
	// surface an empty *string for those.
	in := mockAzureGuide()
	got := newMCPAzureInstallAssets(in)
	require.NotNil(t, got)
	require.Len(t, got.Steps, 2)
	assert.Nil(t, got.Steps[0].Code, "narrative step must have no Code pointer set")
	require.NotNil(t, got.Steps[1].Code)
	assert.Equal(t, "az role assignment create", *got.Steps[1].Code)
	assert.Equal(t, []string{"Reader"}, got.RequiredRbacRoles)
	assert.Equal(t, []string{"Directory.Read.All"}, got.RequiredGraphPermissions)
}
