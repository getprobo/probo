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

package console_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
)

func TestDocumentCollaborationRoom_NotifiesOtherPeers(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{1})
	require.NoError(t, err)

	room := &documentCollaborationRoom{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers: make(map[uint64]chan struct{}),
	}
	room.revision.Store(1)
	hub := &documentCollaborationHub{
		rooms: map[gid.GID]*documentCollaborationRoom{
			versionID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(versionID, room)
	second := hub.addPeerLocked(versionID, room)
	hub.mu.Unlock()

	first.SetRevision(2)
	first.NotifyPeers()

	assert.Equal(t, int64(2), second.Revision())

	select {
	case <-second.Wake:
	default:
		require.Fail(t, "second peer was not notified")
	}

	select {
	case <-first.Wake:
		require.Fail(t, "originating peer must not notify itself")
	default:
	}

	first.Close()
	assert.Contains(t, hub.rooms, versionID)
	second.Close()
	assert.NotContains(t, hub.rooms, versionID)
}
