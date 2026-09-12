// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package serviceaccount_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	serviceaccount "go.probo.inc/probo/pkg/cmd/service-account"
)

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func newFactory(t *testing.T, response string) (*cmdutil.Factory, <-chan graphQLRequest, *strings.Builder) {
	t.Helper()

	requests := make(chan graphQLRequest, 1)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var request graphQLRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				requests <- request
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(response))
			},
		),
	)
	t.Cleanup(server.Close)

	streams, _, _ := iostreams.Test()
	out := new(strings.Builder)
	streams.Out = out
	streams.ForceNonInteractive = true

	cfg := &config.Config{
		ActiveHost: server.URL,
		Hosts: map[string]*config.HostConfig{
			server.URL: {
				Token:        "test-token",
				Organization: "organization-id",
			},
		},
	}

	return &cmdutil.Factory{
		IOStreams: streams,
		Config: func() (*config.Config, error) {
			return cfg, nil
		},
	}, requests, out
}

func TestNewCmdServiceAccount_RegistersLifecycle(t *testing.T) {
	t.Parallel()

	streams, _, _ := iostreams.Test()
	cmd := serviceaccount.NewCmdServiceAccount(
		&cmdutil.Factory{IOStreams: streams},
	)

	children := make(map[string]bool)
	for _, child := range cmd.Commands() {
		children[child.Name()] = true
	}

	assert.True(t, children["list"])
	assert.True(t, children["view"])
	assert.True(t, children["create"])
	assert.True(t, children["update"])
	assert.True(t, children["disable"])
	assert.True(t, children["delete"])
	assert.True(t, children["credential"])

	credential, _, err := cmd.Find([]string{"credential"})
	require.NoError(t, err)

	credentialChildren := make(map[string]bool)
	for _, child := range credential.Commands() {
		credentialChildren[child.Name()] = true
	}

	assert.True(t, credentialChildren["list"])
	assert.True(t, credentialChildren["create"])
	assert.True(t, credentialChildren["revoke"])
}

func TestNewCmdServiceAccount_CreateResolvesOrgAndScopes(t *testing.T) {
	t.Parallel()

	factory, requests, out := newFactory(
		t,
		`{"data":{"createServiceAccount":{"serviceAccountEdge":{"node":{"id":"account-id","name":"Automation"}}}}}`,
	)
	cmd := serviceaccount.NewCmdServiceAccount(factory)
	cmd.SetArgs([]string{
		"create",
		"--name", "Automation",
		"--scope", "v1:asset,v1:control",
		"--scope", "v1:risk",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	request := <-requests
	input := request.Variables["input"].(map[string]any)
	assert.Equal(t, "organization-id", input["organizationId"])
	assert.Equal(t, "Automation", input["name"])
	assert.Equal(t, []any{"v1:asset", "v1:control", "v1:risk"}, input["scopes"])
	assert.Contains(t, out.String(), "Created service account Automation (account-id)")
}

func TestNewCmdServiceAccount_CredentialCreatePrintsOneTimeToken(t *testing.T) {
	t.Parallel()

	factory, requests, out := newFactory(
		t,
		`{"data":{"createServiceAccountCredential":{"serviceAccountCredentialEdge":{"node":{"id":"credential-id","name":"Deploy"}},"token":"secret-once"}}}`,
	)
	cmd := serviceaccount.NewCmdServiceAccount(factory)
	cmd.SetArgs([]string{
		"credential", "create", "account-id",
		"--name", "Deploy",
		"--scope", "v1:asset,v1:risk",
		"--expires-at", "2030-01-02T03:04:05+02:00",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	request := <-requests
	input := request.Variables["input"].(map[string]any)
	assert.Equal(t, "account-id", input["serviceAccountId"])
	assert.Equal(t, []any{"v1:asset", "v1:risk"}, input["scopes"])
	assert.Equal(t, "2030-01-02T03:04:05+02:00", input["expiresAt"])
	assert.Contains(t, out.String(), "it will not be shown again")
	assert.Contains(t, out.String(), "secret-once")
}

func TestNewCmdServiceAccount_CredentialCreateRejectsInvalidExpiration(t *testing.T) {
	t.Parallel()

	streams, _, _ := iostreams.Test()
	cmd := serviceaccount.NewCmdServiceAccount(
		&cmdutil.Factory{
			IOStreams: streams,
			Config: func() (*config.Config, error) {
				t.Fatal("config must not be loaded for invalid expiration")
				return nil, nil
			},
		},
	)
	cmd.SetArgs([]string{
		"credential", "create", "account-id",
		"--name", "Deploy",
		"--scope", "v1:asset",
		"--expires-at", "tomorrow",
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--expires-at must be an RFC3339 timestamp")
}

func TestNewCmdServiceAccount_CredentialListNeverPrintsToken(t *testing.T) {
	t.Parallel()

	factory, _, out := newFactory(
		t,
		`{"data":{"node":{"__typename":"ServiceAccount","credentials":{"totalCount":1,"edges":[{"node":{"id":"credential-id","serviceAccountId":"account-id","name":"Deploy","scopes":["v1:asset"],"expiresAt":"2030-01-02T03:04:05Z","createdAt":"2026-01-02T03:04:05Z","token":"must-not-leak"}}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}`,
	)
	cmd := serviceaccount.NewCmdServiceAccount(factory)
	cmd.SetArgs([]string{"credential", "list", "account-id", "--output", "json"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, out.String(), `"id": "credential-id"`)
	assert.NotContains(t, out.String(), "must-not-leak")
	assert.NotContains(t, out.String(), `"token"`)
}

func TestNewCmdServiceAccount_DestructiveCommandsRequireConfirmation(t *testing.T) {
	t.Parallel()

	testCases := map[string][]string{
		"disable": {"disable", "account-id"},
		"delete":  {"delete", "account-id"},
		"revoke":  {"credential", "revoke", "account-id", "credential-id"},
	}

	for name, args := range testCases {
		name := name
		args := args

		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				streams, _, _ := iostreams.Test()
				streams.ForceNonInteractive = true
				cmd := serviceaccount.NewCmdServiceAccount(
					&cmdutil.Factory{
						IOStreams: streams,
						Config: func() (*config.Config, error) {
							t.Fatal("config must not be loaded without confirmation")
							return nil, nil
						},
					},
				)
				cmd.SetArgs(args)

				err := cmd.Execute()
				require.Error(t, err)
				assert.Contains(t, err.Error(), "confirmation required")
			},
		)
	}
}
