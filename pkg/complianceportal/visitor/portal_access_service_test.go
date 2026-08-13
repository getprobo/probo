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

package visitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestRequestPortalAccessBotEnqueue(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	firstDocumentID := gid.New(tenantID, coredata.DocumentEntityType)
	secondDocumentID := gid.New(tenantID, coredata.DocumentEntityType)
	reportID := gid.New(tenantID, coredata.DocumentEntityType)

	t.Run(
		"skips enqueue when every requested target already exists",
		func(t *testing.T) {
			t.Parallel()

			eventKey, purpose, enqueue := requestPortalAccessBotEnqueue(
				[]gid.GID{firstDocumentID},
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			assert.False(t, enqueue)
			assert.Empty(t, eventKey)
			assert.Empty(t, purpose)
		},
	)

	t.Run(
		"posts a unique event key for the first request",
		func(t *testing.T) {
			t.Parallel()

			eventKey, purpose, enqueue := requestPortalAccessBotEnqueue(
				nil,
				nil,
				nil,
				[]gid.GID{firstDocumentID},
				nil,
				nil,
			)

			require.True(t, enqueue)
			assert.Equal(t, coredata.BotMessagePurposePost, purpose)
			assert.Equal(
				t,
				accessMutationEventKey("request", "", []gid.GID{firstDocumentID}, nil, nil),
				eventKey,
			)
		},
	)

	t.Run(
		"updates with a different event key when later targets are requested",
		func(t *testing.T) {
			t.Parallel()

			firstKey, firstPurpose, firstEnqueue := requestPortalAccessBotEnqueue(
				nil,
				nil,
				nil,
				[]gid.GID{firstDocumentID},
				nil,
				nil,
			)
			secondKey, secondPurpose, secondEnqueue := requestPortalAccessBotEnqueue(
				[]gid.GID{firstDocumentID},
				nil,
				nil,
				[]gid.GID{secondDocumentID},
				[]gid.GID{reportID},
				nil,
			)

			require.True(t, firstEnqueue)
			require.True(t, secondEnqueue)
			assert.Equal(t, coredata.BotMessagePurposePost, firstPurpose)
			assert.Equal(t, coredata.BotMessagePurposeUpdate, secondPurpose)
			assert.NotEqual(t, firstKey, secondKey)
		},
	)
}
