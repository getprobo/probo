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

package console_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTask_AsanaSettingsWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					asanaSettings {
						defaultProject { id name }
						taskImport
					}
					asanaProjects(first: 20) {
						edges { node { id } }
					}
				}
			}
		}
	`

	var result struct {
		Node *struct {
			AsanaSettings struct {
				DefaultProject *struct {
					ID string `json:"id"`
				} `json:"defaultProject"`
				TaskImport string `json:"taskImport"`
			} `json:"asanaSettings"`
			AsanaProjects struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"asanaProjects"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node)
	assert.Nil(t, result.Node.AsanaSettings.DefaultProject)
	assert.Equal(t, "INACTIVE", result.Node.AsanaSettings.TaskImport)
	assert.Empty(t, result.Node.AsanaProjects.Edges)
}

func TestTask_PublishToAsanaWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutInternalControl(owner).Create()

	const mutation = `
		mutation($input: PublishTaskToAsanaInput!) {
			publishTaskToAsana(input: $input) {
				task { id }
			}
		}
	`

	err := owner.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{
			"taskId":    taskID,
			"projectId": "1200000000000001",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")
}

func TestTask_UpdateAsanaSettingsWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	const mutation = `
		mutation($input: UpdateAsanaSettingsInput!) {
			updateAsanaSettings(input: $input) {
				organization { id }
			}
		}
	`

	ownerErr := owner.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{
			"organizationId":   owner.GetOrganizationID().String(),
			"defaultProjectId": "1200000000000001",
			"taskImport":       "ACTIVE",
		},
	})
	testutil.RequireErrorCode(t, ownerErr, "INVALID")

	viewerErr := viewer.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"taskImport":     "ACTIVE",
		},
	})
	testutil.RequireForbiddenError(t, viewerErr)
}

func TestTask_AsanaWebhookRejectsBadConnector(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/asana/v1/webhooks/not-a-gid/nonce",
		bytes.NewReader([]byte(`{}`)),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hook-Secret", "handshake")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTask_LinkToAsanaWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutInternalControl(owner).Create()

	const mutation = `
		mutation($input: LinkTaskToAsanaInput!) {
			linkTaskToAsana(input: $input) {
				task { id }
			}
		}
	`

	err := owner.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{
			"taskId":    taskID,
			"projectId": "1200000000000001",
			"taskGid":   "1200000000000002",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")
}

func TestTask_UpdateAsanaProjectMappingWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const mutation = `
		mutation($input: UpdateAsanaProjectMappingInput!) {
			updateAsanaProjectMapping(input: $input) {
				setup { project { id } }
			}
		}
	`

	err := owner.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"projectId":      "1200000000000001",
			"clear":          true,
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")
}

func TestTask_AsanaListsWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					asanaTeams(first: 20) {
						edges { node { id } }
					}
					asanaProjects(first: 20) {
						edges { node { id } }
					}
					asanaTasks(projectId: "1200000000000001", first: 20) {
						edges { node { id } }
					}
				}
			}
		}
	`

	var result struct {
		Node *struct {
			AsanaTeams struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"asanaTeams"`
			AsanaProjects struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"asanaProjects"`
			AsanaTasks struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"asanaTasks"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node)
	assert.Empty(t, result.Node.AsanaTeams.Edges)
	assert.Empty(t, result.Node.AsanaProjects.Edges)
	assert.Empty(t, result.Node.AsanaTasks.Edges)
}

func TestTask_AsanaWebhookRejectsUnknownNonce(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/asana/v1/webhooks/"+owner.GetOrganizationID().String()+"/unknown-nonce",
		bytes.NewReader([]byte(`{"events":[]}`)),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hook-Secret", "handshake")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestTask_AsanaWebhookRejectsBadSignature(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/asana/v1/webhooks/"+owner.GetOrganizationID().String()+"/unknown-nonce",
		bytes.NewReader([]byte(`{"events":[]}`)),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hook-Signature", "deadbeef")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
