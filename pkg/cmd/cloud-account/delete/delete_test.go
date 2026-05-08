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

package delete_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	cadelete "go.probo.inc/probo/pkg/cmd/cloud-account/delete"
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

func TestDelete_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := cadelete.NewCmdDelete(f)

	assert.NotNil(t, cmd.Flags().Lookup("yes"))
	short := cmd.Flags().ShorthandLookup("y")
	assert.NotNil(t, short, "--yes should have -y short flag")
}

func TestDelete_RequiresIDArg(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := cadelete.NewCmdDelete(f)
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

// Without --yes, in non-interactive mode delete must refuse rather
// than block on a TTY prompt.
func TestDelete_NonInteractive_RequiresYes(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	io.ForceNonInteractive = true
	cfg := &config.Config{
		ActiveHost: "http://127.0.0.1:1",
		Hosts: map[string]*config.HostConfig{
			"http://127.0.0.1:1": {Token: "t", Organization: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		},
	}
	f := &cmdutil.Factory{
		IOStreams: io,
		Config:    func() (*config.Config, error) { return cfg, nil },
	}
	cmd := cadelete.NewCmdDelete(f)
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)
	cmd.SetArgs([]string{"some-id"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirmation required")
}

const deleteResp = `{
  "data": {
    "deleteCloudAccount": {
      "deletedCloudAccountId": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    }
  }
}`

func TestDelete_WithYes_BuildsMutationInput(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, deleteResp)

	f, ios := newFactory(srv.URL)
	cmd := cadelete.NewCmdDelete(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "--yes"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "deleteCloudAccount")
	input, ok := capturedReq.Variables["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", input["cloudAccountId"])

	out := ios.Out.(interface{ String() string }).String()
	assert.Contains(t, out, "Deleted cloud account")
}
