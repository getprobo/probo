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

package list_test

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
	"go.probo.inc/probo/pkg/cmd/cloud-account/list"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

// captured holds the request body the fake GraphQL server received,
// for assertions on mutation/query construction.
type captured struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
	Raw       map[string]any `json:"-"`
}

// fakeServer spins up an httptest.Server that captures the GraphQL
// request body and returns the supplied response.
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

// newFactory builds a Factory pointing at host (an httptest base URL)
// with an organization id wired in the host config.
func newFactory(host, org string) (*cmdutil.Factory, *iostreams.IOStreams) {
	io, _, _ := iostreams.Test()
	io.ForceNonInteractive = true
	cfg := &config.Config{
		ActiveHost: host,
		Hosts: map[string]*config.HostConfig{
			host: {
				Token:        "test-token",
				Organization: org,
			},
		},
	}
	return &cmdutil.Factory{
		IOStreams: io,
		Config:    func() (*config.Config, error) { return cfg, nil },
	}, io
}

func TestList_Flags_Registered(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	cmd := list.NewCmdList(f)

	for _, name := range []string{
		"org",
		"limit",
		"order-by",
		"order-direction",
		"provider",
		"status",
		"scope-kind",
		"output",
	} {
		flag := cmd.Flags().Lookup(name)
		assert.NotNil(t, flag, "expected --%s flag", name)
	}

	assert.Equal(t, []string{"ls"}, cmd.Aliases)
}

func TestList_RejectsInvalidOutput(t *testing.T) {
	t.Parallel()

	f, _ := newFactory("http://127.0.0.1:1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--output", "yaml"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output")
}

func TestList_RejectsInvalidProvider(t *testing.T) {
	t.Parallel()

	f, _ := newFactory("http://127.0.0.1:1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--provider", "DigitalOcean"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --provider")
}

func TestList_RejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	f, _ := newFactory("http://127.0.0.1:1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--order-by", "STATUS", "--status", "FROZEN"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --status")
}

func TestList_RejectsInvalidOrderDirection(t *testing.T) {
	t.Parallel()

	f, _ := newFactory("http://127.0.0.1:1", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--order-direction", "SIDEWAYS"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --order-direction")
}

func TestList_NoOrgFails(t *testing.T) {
	t.Parallel()

	// Empty org in host config -- list must reject before contacting the server.
	io, _, _ := iostreams.Test()
	io.ForceNonInteractive = true
	cfg := &config.Config{
		ActiveHost: "http://127.0.0.1:1",
		Hosts: map[string]*config.HostConfig{
			"http://127.0.0.1:1": {Token: "t"},
		},
	}
	f := &cmdutil.Factory{
		IOStreams: io,
		Config:    func() (*config.Config, error) { return cfg, nil },
	}

	cmd := list.NewCmdList(f)
	cmd.SetOut(io.Out)
	cmd.SetErr(io.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization")
}

const listEmptyResponse = `{
  "data": {
    "node": {
      "__typename": "Organization",
      "cloudAccounts": {
        "totalCount": 0,
        "edges": [],
        "pageInfo": {"hasNextPage": false, "endCursor": null}
      }
    }
  }
}`

func TestList_OutputJSON_BuildsQuery(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, listEmptyResponse)

	f, ios := newFactory(srv.URL, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"--output", "json", "--provider", "AWS"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, capturedReq.Query, "cloudAccounts")
	require.NotNil(t, capturedReq.Variables)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", capturedReq.Variables["id"])

	filter, ok := capturedReq.Variables["filter"].(map[string]any)
	require.True(t, ok, "filter variable must be a map")
	assert.Equal(t, "AWS", filter["provider"])

	// Empty list with --output json produces a valid JSON array.
	out, ok := ios.Out.(interface{ String() string })
	require.True(t, ok)
	trimmed := strings.TrimSpace(out.String())
	require.NotEmpty(t, trimmed)

	var parsed []any
	require.NoError(t, json.Unmarshal([]byte(trimmed), &parsed), "JSON output must parse")
	assert.Empty(t, parsed)
}

func TestList_OutputTable_Empty(t *testing.T) {
	t.Parallel()

	var capturedReq captured
	srv := fakeServer(t, &capturedReq, listEmptyResponse)

	f, ios := newFactory(srv.URL, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	cmd := list.NewCmdList(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	out, ok := ios.Out.(interface{ String() string })
	require.True(t, ok)
	assert.Contains(t, out.String(), "No cloud accounts found")
}
