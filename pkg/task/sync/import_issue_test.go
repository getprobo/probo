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
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestIsLinearApplicationActor(t *testing.T) {
	t.Parallel()

	assert.True(t, isLinearApplicationActor(linear.WebhookActor{Type: "Application"}))
	assert.True(t, isLinearApplicationActor(linear.WebhookActor{Type: "application"}))
	assert.False(t, isLinearApplicationActor(linear.WebhookActor{Type: "user"}))
	assert.False(t, isLinearApplicationActor(linear.WebhookActor{}))
}

func TestImportedTaskName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ship the import", importedTaskName(&linear.IssueWebhookData{
		Title: "  Ship\nthe   import  ",
	}))
	assert.Equal(t, "ENG-1", importedTaskName(&linear.IssueWebhookData{
		Identifier: "ENG-1",
	}))
	assert.Empty(t, importedTaskName(&linear.IssueWebhookData{}))
}
