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

package rotate_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cloud-account/rotate"
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

func TestRotate_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := rotate.NewCmdRotate(f)

	for _, name := range []string{
		"provider",
		"credential-kind",
		"aws-role-arn",
		"aws-external-id",
		"azure-tenant-id",
		"azure-client-id",
	} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "expected --%s flag", name)
	}
}

// Plan: rotate help/long text must reflect the per-provider flag groups
// and the secret-out-of-band rule.
func TestRotate_Help_DescribesProviderGroups(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := rotate.NewCmdRotate(f)

	combined := cmd.Long + "\n" + cmd.Short
	assert.Contains(t, combined, "AWS")
	assert.Contains(t, combined, "Azure")
	// Secret bodies must NOT be passed to rotate -- they go through
	// the credential-upload endpoint.
	assert.Contains(t, combined, "credentials/upload",
		"rotate help text must point operators at the dedicated upload endpoint for secret bodies")
}

func TestRotate_RequiresMandatoryFlags(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := rotate.NewCmdRotate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"some-id"}) // missing --provider/--credential-kind

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestRotate_RequiresIDArg(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := rotate.NewCmdRotate(f)
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)
	cmd.SetArgs([]string{
		"--provider", "AWS",
		"--credential-kind", "AWS_ASSUME_ROLE",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

func TestRotate_RejectsInvalidProvider(t *testing.T) {
	t.Parallel()

	f, ios := newFactory("http://127.0.0.1:1")
	cmd := rotate.NewCmdRotate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"some-id",
		"--provider", "DigitalOcean",
		"--credential-kind", "AWS_ASSUME_ROLE",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --provider")
}

const rotateResp = `{
  "data": {
    "rotateCloudAccountCredentials": {
      "cloudAccount": {"id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "status": "PENDING_VERIFICATION"},
      "verifyStatus": "ERRORED",
      "lastProbeError": "trust policy mismatch"
    }
  }
}`

func TestRotate_BuildsAWSInputShape(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, rotateResp)

	f, ios := newFactory(srv.URL)
	cmd := rotate.NewCmdRotate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"--provider", "AWS",
		"--credential-kind", "AWS_ASSUME_ROLE",
		"--aws-role-arn", "arn:aws:iam::123456789012:role/Rotated",
		"--aws-external-id", "deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "rotateCloudAccountCredentials")
	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", input["cloudAccountId"])
	assert.Equal(t, "AWS", input["provider"])
	assert.Equal(t, "AWS_ASSUME_ROLE", input["credentialKind"])
	assert.Equal(t, "arn:aws:iam::123456789012:role/Rotated", input["awsRoleArn"])
	assert.Equal(t, "deadbeefcafebabedeadbeefcafebabedeadbeefcafebabedeadbeefcafebabe", input["awsExternalId"])
}

func TestRotate_BuildsAzureInputShape(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, rotateResp)

	f, ios := newFactory(srv.URL)
	cmd := rotate.NewCmdRotate(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{
		"X",
		"--provider", "AZURE",
		"--credential-kind", "AZURE_CLIENT_SECRET",
		"--azure-tenant-id", "00000000-0000-0000-0000-000000000002",
		"--azure-client-id", "00000000-0000-0000-0000-000000000003",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AZURE", input["provider"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", input["azureTenantId"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", input["azureClientId"])
	_, hasAWS := input["awsRoleArn"]
	assert.False(t, hasAWS, "awsRoleArn must not appear in an Azure rotate request")
}
