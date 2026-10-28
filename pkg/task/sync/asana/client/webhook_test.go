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

package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"events":[]}`)
	secret := "hook-secret"
	signature := Sign(secret, body)

	assert.True(t, VerifySignature(secret, signature, body))
	assert.False(t, VerifySignature(secret, signature, []byte(`{"events":[{"action":"changed"}]}`)))
	assert.False(t, VerifySignature("", signature, body))
}

func TestParseEventsReadsTaskIdentity(t *testing.T) {
	t.Parallel()

	batch, err := ParseEvents([]byte(`{
		"events": [{
			"action": "added",
			"resource": {"gid": "task-1", "resource_type": "task"},
			"parent": {"gid": "project-1", "resource_type": "project"}
		}]
	}`))
	require.NoError(t, err)
	require.Len(t, batch.Events, 1)
	assert.Equal(t, "task-1", batch.Events[0].TaskGID())
	assert.True(t, batch.Events[0].AddedTask("project-1"))
	assert.False(t, batch.Events[0].DeletedTask())
	_, removed := batch.Events[0].ProjectRemoval()
	assert.False(t, removed)
}

func TestDeletedTaskIgnoresSectionRemoval(t *testing.T) {
	t.Parallel()

	batch, err := ParseEvents([]byte(`{
		"events": [{
			"action": "removed",
			"resource": {"gid": "task-1", "resource_type": "task"},
			"parent": {"gid": "section-1", "resource_type": "section"}
		}, {
			"action": "deleted",
			"resource": {"gid": "task-2", "resource_type": "task"}
		}, {
			"action": "removed",
			"resource": {"gid": "task-3", "resource_type": "task"},
			"parent": {"gid": "project-1", "resource_type": "project"}
		}]
	}`))
	require.NoError(t, err)
	require.Len(t, batch.Events, 3)

	assert.False(t, batch.Events[0].DeletedTask())
	_, removed := batch.Events[0].ProjectRemoval()
	assert.False(t, removed)
	assert.Equal(t, "task-1", batch.Events[0].TaskGID())

	assert.True(t, batch.Events[1].DeletedTask())

	projectGID, removed := batch.Events[2].ProjectRemoval()
	assert.True(t, removed)
	assert.Equal(t, "project-1", projectGID)
	assert.False(t, batch.Events[2].DeletedTask())
}

func TestCreateWebhook_AttachmentFilters(t *testing.T) {
	t.Parallel()

	var (
		mu      sync.Mutex
		raw     []byte
		readErr error
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		mu.Lock()

		raw = body
		readErr = err

		mu.Unlock()

		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/webhooks") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"gid":"wh-1"}}`))
	}))
	t.Cleanup(server.Close)

	webhookGID, err := NewClient(server.Client(), server.URL).CreateWebhook(
		t.Context(),
		"project-1",
		"https://example.com/hook",
	)
	require.NoError(t, err)
	assert.Equal(t, "wh-1", webhookGID)

	mu.Lock()

	body := append([]byte(nil), raw...)
	err = readErr

	mu.Unlock()

	require.NoError(t, err)

	var request struct {
		Data struct {
			Filters []struct {
				ResourceType string `json:"resource_type"`
				Action       string `json:"action"`
			} `json:"filters"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &request))

	var (
		attachmentActions []string
		taskAdded         bool
	)

	for _, filter := range request.Data.Filters {
		if filter.ResourceType == "task" && filter.Action == "added" {
			taskAdded = true
		}

		if filter.ResourceType != "attachment" {
			continue
		}

		attachmentActions = append(attachmentActions, filter.Action)
	}

	assert.True(t, taskAdded)
	assert.ElementsMatch(t, []string{"deleted", "undeleted"}, attachmentActions)
}

func TestDownloadAllowedHosts(t *testing.T) {
	t.Parallel()

	client := &Client{}

	assert.True(t, client.downloadAllowed("https://app.asana.com/app/asana/-/get_asset?id=1"))
	assert.True(t, client.downloadAllowed("https://asana-user-private-us-east-1.s3.us-east-1.amazonaws.com/file"))
	assert.False(t, client.downloadAllowed("https://notasana.s3.amazonaws.com/file"))
	assert.False(t, client.downloadAllowed("http://app.asana.com/file"))
}
