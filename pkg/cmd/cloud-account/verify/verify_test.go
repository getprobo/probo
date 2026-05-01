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

package verify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cloud-account/verify"
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

func TestVerify_RequiresIDArg(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := verify.NewCmdVerify(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

const verifyResp = `{
  "data": {
    "verifyCloudAccount": {
      "cloudAccount": {"id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "status": "VERIFIED"},
      "status": "VERIFIED",
      "lastProbeError": null
    }
  }
}`

func TestVerify_BuildsMutationInput(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, verifyResp)

	f, ios := newFactory(srv.URL)
	cmd := verify.NewCmdVerify(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "verifyCloudAccount")
	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", input["cloudAccountId"])

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "Verified cloud account")
	assert.Contains(t, out, "VERIFIED")
}

func TestVerify_RendersProbeError(t *testing.T) {
	t.Parallel()

	resp := `{
  "data": {
    "verifyCloudAccount": {
      "cloudAccount": {"id": "X", "status": "PENDING_VERIFICATION"},
      "status": "PENDING_VERIFICATION",
      "lastProbeError": "trust policy denied"
    }
  }
}`

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, resp)

	f, ios := newFactory(srv.URL)
	cmd := verify.NewCmdVerify(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"X"})

	err := cmd.Execute()
	require.NoError(t, err)
	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "trust policy denied")
}
