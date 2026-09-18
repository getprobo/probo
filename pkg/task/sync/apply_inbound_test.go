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

package tasksync

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestIsAppActor(t *testing.T) {
	t.Parallel()

	metadata, err := json.Marshal(map[string]string{"app_actor_id": "app-1"})
	require.NoError(t, err)

	assert.True(t, isAppActor(metadata, "app-1"))
	assert.False(t, isAppActor(metadata, "user-1"))
	assert.False(t, isAppActor(nil, "app-1"))
	assert.False(t, isAppActor(metadata, ""))
}

func TestInboundEventIsStale(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	newer := older.Add(time.Minute)

	assert.True(t, inboundEventIsStale(nil, &newer))
	assert.False(t, inboundEventIsStale(&newer, nil))
	assert.False(t, inboundEventIsStale(nil, nil))
	assert.False(t, inboundEventIsStale(&newer, &older))
	assert.True(t, inboundEventIsStale(&older, &newer))
	assert.True(t, inboundEventIsStale(&newer, &newer))
}

func TestLinearDateTime(t *testing.T) {
	t.Parallel()

	got := linearDateTime("2026-09-14T12:00:00.000Z")
	require.NotNil(t, got)
	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.September, got.Month())
	assert.Equal(t, 14, got.Day())

	assert.Nil(t, linearDateTime(""))
	assert.Nil(t, linearDateTime("not-a-date"))
}

func TestPickLinkForLinearOrganization(t *testing.T) {
	t.Parallel()

	linkA := &coredata.TaskExternalLink{
		Destination: json.RawMessage(`{"team_id":"team-1","linear_organization_id":"org-a"}`),
	}
	linkB := &coredata.TaskExternalLink{
		Destination: json.RawMessage(`{"team_id":"team-1","linear_organization_id":"org-b"}`),
	}

	got, ok := PickLinkForLinearOrganization(coredata.TaskExternalLinks{}, "org-a")
	assert.False(t, ok)
	assert.Nil(t, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA}, "org-a")
	require.True(t, ok)
	assert.Equal(t, linkA, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA}, "org-other")
	assert.False(t, ok)
	assert.Nil(t, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA, linkB}, "org-b")
	require.True(t, ok)
	assert.Equal(t, linkB, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA, linkB}, "org-missing")
	assert.False(t, ok)
	assert.Nil(t, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA, linkB}, "")
	assert.False(t, ok)
	assert.Nil(t, got)

	got, ok = PickLinkForLinearOrganization(coredata.TaskExternalLinks{linkA}, "")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestMapInboundIssue_PartialPayloadKeepsOmittedFields(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	content, err := MarkdownToContent("Keep this body")
	require.NoError(t, err)

	task := &coredata.Task{
		Name:     "Original",
		Content:  content,
		State:    coredata.TaskStateTodo,
		Priority: coredata.TaskPriorityHigh,
		Deadline: &deadline,
	}

	envelope, err := linear.ParseEnvelope([]byte(`{
		"action":"update",
		"type":"Issue",
		"data":{"title":"Renamed","dueDate":null}
	}`))
	require.NoError(t, err)

	data, err := envelope.IssueData()
	require.NoError(t, err)

	mapped, err := mapInboundIssue(task, data)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", mapped.Name)
	assert.Equal(t, content, mapped.Content)
	assert.Equal(t, coredata.TaskStateTodo, mapped.State)
	assert.Equal(t, coredata.TaskPriorityHigh, mapped.Priority)
	assert.Nil(t, mapped.Deadline)
}
