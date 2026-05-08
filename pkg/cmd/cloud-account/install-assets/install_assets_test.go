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

package installassets_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	installassets "go.probo.inc/probo/pkg/cmd/cloud-account/install-assets"
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

func TestInstallAssets_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := installassets.NewCmdInstallAssets(f)

	for _, name := range []string{
		"org",
		"provider",
		"scope-kind",
		"scope-identifier",
		"modules",
		"aws-region",
		"output",
	} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "expected --%s flag", name)
	}
}

func TestInstallAssets_RequiresMandatoryFlags(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestInstallAssets_RejectsInvalidOutput(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "AWS",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123",
		"--modules", "ACCESS_REVIEW",
		"--output", "yaml",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output")
}

const awsAssetsResp = `{
  "data": {
    "generateCloudAccountInstallAssets": {
      "assets": {
        "__typename": "AWSInstallAssets",
        "quickCreateURL": "https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/quickcreate?templateUrl=https%3A%2F%2Fexample.com%2Ftpl",
        "externalId": "deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe",
        "principalArn": "arn:aws:iam::000000000000:root",
        "requiredActions": ["iam:GenerateCredentialReport", "iam:ListUsers"]
      }
    }
  }
}`

// Plan line 120: install-assets --output json must include external_id
// so an operator can pipe it via jq into 'create --aws-external-id ...'.
func TestInstallAssets_OutputJSON_IncludesExternalID(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, awsAssetsResp)

	f, ios := newFactory(srv.URL)
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "AWS",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123456789012",
		"--modules", "ACCESS_REVIEW",
		"--aws-region", "us-east-1",
		"--output", "json",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "generateCloudAccountInstallAssets")
	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AWS", input["provider"])
	assert.Equal(t, "AWS_ACCOUNT", input["scopeKind"])
	assert.Equal(t, "us-east-1", input["awsRegion"])

	out := ios.Out.(interface{ String() string }).String()
	trimmed := strings.TrimSpace(out)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(trimmed), &parsed),
		"--output json must produce parseable JSON (plan line 120: pipeline via jq)")
	assert.Equal(t,
		"deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe",
		parsed["externalId"],
		"externalId must appear in JSON output for operator pipeline (plan line 120)")
}

func TestInstallAssets_AWS_TableOutput(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, awsAssetsResp)

	f, ios := newFactory(srv.URL)
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "AWS",
		"--scope-kind", "AWS_ACCOUNT",
		"--scope-identifier", "123456789012",
		"--modules", "ACCESS_REVIEW",
		"--aws-region", "us-east-1",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "AWS Install Assets")
	assert.Contains(t, out, "Quick-Create URL")
	assert.Contains(t, out, "External ID")
	// Plan line 120: human-readable output should ALSO point at create.
	assert.Contains(t, out, "prb cloud-account create")
}

const gcpAssetsResp = `{
  "data": {
    "generateCloudAccountInstallAssets": {
      "assets": {
        "__typename": "GCPInstallAssets",
        "setupScript": "#!/bin/bash\necho gcloud projects add-iam-policy-binding ...",
        "requiredRoles": ["roles/iam.securityReviewer"],
        "requiredApis": ["cloudresourcemanager.googleapis.com"]
      }
    }
  }
}`

func TestInstallAssets_GCP_TableOutput(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, gcpAssetsResp)

	f, ios := newFactory(srv.URL)
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "GCP",
		"--scope-kind", "GCP_PROJECT",
		"--scope-identifier", "my-project-123",
		"--modules", "ACCESS_REVIEW",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "GCP", input["provider"])
	_, hasAWSRegion := input["awsRegion"]
	assert.False(t, hasAWSRegion, "awsRegion must not appear in a GCP install-assets request")

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "GCP Install Assets")
	assert.Contains(t, out, "Setup script")
	assert.Contains(t, out, "Required roles")
}

const azureAssetsResp = `{
  "data": {
    "generateCloudAccountInstallAssets": {
      "assets": {
        "__typename": "AzureInstallAssets",
        "steps": [
          {"title": "Step 1", "body": "Open Azure portal", "code": null},
          {"title": "Step 2", "body": "Run", "code": "az role assignment create ..."}
        ],
        "requiredRbacRoles": ["Reader"],
        "requiredGraphPermissions": ["User.Read.All"]
      }
    }
  }
}`

func TestInstallAssets_Azure_TableOutput(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, azureAssetsResp)

	f, ios := newFactory(srv.URL)
	cmd := installassets.NewCmdInstallAssets(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "AZURE",
		"--scope-kind", "AZURE_MANAGEMENT_GROUP",
		"--scope-identifier", "00000000-0000-0000-0000-000000000003",
		"--modules", "ACCESS_REVIEW",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AZURE", input["provider"])

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "Azure Install Assets")
	assert.Contains(t, out, "Step 1")
	assert.Contains(t, out, "az role assignment create")
}
