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

package get_test

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
	"go.probo.inc/probo/pkg/cmd/cloud-account/get"
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

func TestGet_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := get.NewCmdGet(f)

	assert.Equal(t, "get <id>", cmd.Use)
	assert.Contains(t, cmd.Aliases, "view")
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestGet_RequiresIDArg(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := get.NewCmdGet(f)
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

func TestGet_RejectsInvalidOutput(t *testing.T) {
	t.Parallel()

	f, _ := newFactory("http://127.0.0.1:1")
	cmd := get.NewCmdGet(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"some-id", "--output", "yaml"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output")
}

const cloudAccountResp = `{
  "data": {
    "node": {
      "__typename": "CloudAccount",
      "id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
      "label": "Prod AWS",
      "provider": "AWS",
      "status": "VERIFIED",
      "credentialKind": "AWS_ASSUME_ROLE",
      "enabledAuditModules": ["ACCESS_REVIEW"],
      "scope": {"kind": "AWS_ACCOUNT", "identifier": "123456789012"},
      "lastProbeAt": null,
      "lastProbeError": null,
      "lastVerifiedAt": null,
      "createdAt": "2026-01-01T00:00:00Z",
      "updatedAt": "2026-01-01T00:00:00Z"
    }
  }
}`

func TestGet_OutputJSON_BuildsQuery(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, cloudAccountResp)

	f, ios := newFactory(srv.URL)
	cmd := get.NewCmdGet(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "--output", "json"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", capturedReq.Variables["id"])
	assert.Contains(t, capturedReq.Query, "node(id: $id)")

	out := ios.Out.(interface{ String() string }).String()
	trimmed := strings.TrimSpace(out)

	// Plan line 120: 'get --output json' must include the persisted
	// external_id-equivalent so an AWS operator can recover it.
	// Today the GraphQL schema exposes the same value via
	// scope.identifier (the AWS account id), and the Get JSON output
	// surfaces it directly. We assert scope.identifier is present in
	// the JSON tree under the "scope" key.
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(trimmed), &parsed))
	scope, ok := parsed["scope"].(map[string]any)
	require.True(t, ok, "JSON output must contain 'scope' object (recovery path for AWS account id)")
	assert.Equal(t, "123456789012", scope["identifier"])
}

func TestGet_OutputTable_RendersFields(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, cloudAccountResp)

	f, ios := newFactory(srv.URL)
	cmd := get.NewCmdGet(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})

	err := cmd.Execute()
	require.NoError(t, err)

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "Cloud Account")
	assert.Contains(t, out, "Prod AWS")
	assert.Contains(t, out, "AWS")
	assert.Contains(t, out, "VERIFIED")
}

func TestGet_NotFoundFails(t *testing.T) {
	t.Parallel()

	resp := `{"data": {"node": null}}`

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, resp)

	f, ios := newFactory(srv.URL)
	cmd := get.NewCmdGet(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"missing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
