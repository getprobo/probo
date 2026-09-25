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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestInboundAssigneeEmail(t *testing.T) {
	t.Parallel()

	t.Run(
		"omitted assignee does not apply",
		func(t *testing.T) {
			t.Parallel()

			envelope, err := linear.ParseEnvelope([]byte(`{
				"action":"update",
				"type":"Issue",
				"data":{"title":"Renamed"}
			}`))
			require.NoError(t, err)

			data, err := envelope.IssueData()
			require.NoError(t, err)

			email, ok := inboundAssigneeEmail(data)
			assert.False(t, ok)
			assert.Empty(t, email)
		},
	)

	t.Run(
		"null assignee does not apply",
		func(t *testing.T) {
			t.Parallel()

			envelope, err := linear.ParseEnvelope([]byte(`{
				"action":"update",
				"type":"Issue",
				"data":{"assignee":null}
			}`))
			require.NoError(t, err)

			data, err := envelope.IssueData()
			require.NoError(t, err)
			require.True(t, data.Has("assignee"))

			email, ok := inboundAssigneeEmail(data)
			assert.False(t, ok)
			assert.Empty(t, email)
		},
	)

	t.Run(
		"assignee without email does not apply",
		func(t *testing.T) {
			t.Parallel()

			envelope, err := linear.ParseEnvelope([]byte(`{
				"action":"update",
				"type":"Issue",
				"data":{"assignee":{"id":"user-1","name":"Jane"}}
			}`))
			require.NoError(t, err)

			data, err := envelope.IssueData()
			require.NoError(t, err)

			email, ok := inboundAssigneeEmail(data)
			assert.False(t, ok)
			assert.Empty(t, email)
		},
	)

	t.Run(
		"assignee email is returned",
		func(t *testing.T) {
			t.Parallel()

			envelope, err := linear.ParseEnvelope([]byte(`{
				"action":"update",
				"type":"Issue",
				"data":{"assignee":{"id":"user-1","email":"jane@example.com"}}
			}`))
			require.NoError(t, err)

			data, err := envelope.IssueData()
			require.NoError(t, err)

			email, ok := inboundAssigneeEmail(data)
			assert.True(t, ok)
			assert.Equal(t, "jane@example.com", email)
		},
	)
}
