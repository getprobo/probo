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

package create_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cloud-account/create"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

type captured struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func fakeServer(t *testing.T, capturedReq *captured, resp string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, capturedReq))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, resp)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newFactory(host string) (*cmdutil.Factory, *iostreams.IOStreams) {
	io, _, _ := iostreams.Test()
	io.ForceNonInteractive = true
	cfg := &config.Config{
		ActiveHost: host,
		Hosts: map[string]*config.HostConfig{
			host: {Token: "test-token", Organization: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		},
	}
	return &cmdutil.Factory{
		IOStreams: io,
		Config:    func() (*config.Config, error) { return cfg, nil },
	}, io
}

func TestCreate_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := create.NewCmdCreate(f)

	for _, name := range []string{
		"org",
		"label",
		"provider",
		"credential-kind",
		"scope-kind",
		"scope-identifier",
		"modules",
		"aws-role-arn",
		"aws-external-id",
		"gcp-project-id",
		"gcp-organization-id",
		"azure-tenant-id",
		"azure-client-id",
		"azure-subscription-id",
		"azure-management-group-id",
	} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "expected --%s flag", name)
	}
}

// Plan line 116: the create help text MUST point at install-assets
// so an operator never invents an --aws-external-id value.
func TestCreate_Help_ReferencesInstallAssets(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := create.NewCmdCreate(f)

	combined := cmd.Long + "\n" + cmd.Example
	assert.Contains(t, combined, "install-assets",
		"create help text must reference install-assets (plan line 116)")
	assert.Contains(t, combined, "external_id",
		"create help text must mention external_id so operators don't invent one")
}

func TestCreate_RequiresMandatoryFlags(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{}) // missing required flags

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestCreate_RejectsInvalidProvider(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "X",
		"--provider", "DigitalOcean",
		"--credential-kind", "AWS_ASSUME_ROLE",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123",
		"--modules", "ACCESS_REVIEW",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --provider")
}

func TestCreate_RejectsInvalidCredentialKind(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "X",
		"--provider", "AWS",
		"--credential-kind", "MAGIC",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123",
		"--modules", "ACCESS_REVIEW",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --credential-kind")
}

func TestCreate_RejectsInvalidModule(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "X",
		"--provider", "AWS",
		"--credential-kind", "AWS_ASSUME_ROLE",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123",
		"--modules", "BOGUS",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --modules")
}

const createResp = `{
  "data": {
    "createCloudAccount": {
      "cloudAccount": {
        "id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        "label": "Prod AWS",
        "provider": "AWS",
        "status": "PENDING_VERIFICATION"
      },
      "verifyStatus": "ERRORED",
      "lastProbeError": "no real cloud reachable"
    }
  }
}`

func TestCreate_BuildsAWSInputShape(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, createResp)

	f, ios := newFactory(srv.URL)
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "Prod AWS",
		"--provider", "AWS",
		"--credential-kind", "AWS_ASSUME_ROLE",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123456789012",
		"--modules", "ACCESS_REVIEW",
		"--aws-role-arn", "arn:aws:iam::123456789012:role/Probo",
		"--aws-external-id", "deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "createCloudAccount")
	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "Prod AWS", input["label"])
	assert.Equal(t, "AWS", input["provider"])
	assert.Equal(t, "AWS_ASSUME_ROLE", input["credentialKind"])
	assert.Equal(t, "AWS_ACCOUNT", input["scopeKind"])
	assert.Equal(t, "123456789012", input["scopeIdentifier"])
	assert.Equal(t, "arn:aws:iam::123456789012:role/Probo", input["awsRoleArn"])
	assert.Equal(t, "deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe", input["awsExternalId"])

	modules, ok := input["enabledAuditModules"].([]any)
	require.True(t, ok)
	require.Len(t, modules, 1)
	assert.Equal(t, "ACCESS_REVIEW", modules[0])
}

func TestCreate_BuildsGCPInputShape(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, createResp)

	f, ios := newFactory(srv.URL)
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "GCP-Prod",
		"--provider", "GCP",
		"--credential-kind", "GCP_SERVICE_ACCOUNT_KEY",
		"--scope-kind", "GCP_PROJECT",
		"--scope-identifier", "my-project-123",
		"--modules", "ACCESS_REVIEW",
		"--gcp-project-id", "my-project-123",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "GCP", input["provider"])
	assert.Equal(t, "GCP_SERVICE_ACCOUNT_KEY", input["credentialKind"])
	assert.Equal(t, "my-project-123", input["gcpProjectId"])
	// AWS-specific keys must NOT leak into a GCP request.
	_, awsRoleSet := input["awsRoleArn"]
	assert.False(t, awsRoleSet, "awsRoleArn must not appear in a GCP create request")
}

func TestCreate_BuildsAzureInputShape(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, createResp)

	f, ios := newFactory(srv.URL)
	cmd := create.NewCmdCreate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--label", "Azure-Prod",
		"--provider", "AZURE",
		"--credential-kind", "AZURE_CLIENT_SECRET",
		"--scope-kind", "AZURE_SUBSCRIPTION",
		"--scope-identifier", "00000000-0000-0000-0000-000000000001",
		"--modules", "ACCESS_REVIEW",
		"--azure-tenant-id", "00000000-0000-0000-0000-000000000002",
		"--azure-client-id", "00000000-0000-0000-0000-000000000003",
		"--azure-subscription-id", "00000000-0000-0000-0000-000000000001",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AZURE", input["provider"])
	assert.Equal(t, "AZURE_CLIENT_SECRET", input["credentialKind"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", input["azureTenantId"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", input["azureClientId"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", input["azureSubscriptionId"])
}
