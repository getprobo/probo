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

package collaboration_v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type fakeDocumentCollaborationDocuments struct {
	opened      *collaboration.DocumentCollaboration
	persisted   atomic.Int64
	persistedCh chan struct{}
}

func (f *fakeDocumentCollaborationDocuments) Open(
	context.Context,
	coredata.Scoper,
	gid.GID,
) (*collaboration.DocumentCollaboration, error) {
	return f.opened, nil
}

func (f *fakeDocumentCollaborationDocuments) Persist(
	context.Context,
	coredata.Scoper,
	gid.GID,
	*automerge.Document,
) (int64, error) {
	revision := f.persisted.Add(1) + 1

	select {
	case f.persistedCh <- struct{}{}:
	default:
	}

	return revision, nil
}

func newRepoTestServer(
	t *testing.T,
	hub *Hub,
	scope coredata.Scoper,
	documentID gid.GID,
	publishEphemeral func(ctx context.Context, frame []byte) error,
) *httptest.Server {
	t.Helper()

	return newRepoTestServerWithConfig(t, hub, scope, documentID, RepoConfig{
		Shutdown:         context.Background(),
		PublishEphemeral: publishEphemeral,
	})
}

func newRepoTestServerWithConfig(
	t *testing.T,
	hub *Hub,
	scope coredata.Scoper,
	documentID gid.GID,
	config RepoConfig,
) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionID, err := NewConnectionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lease, err := hub.Acquire(r.Context(), scope, documentID, connectionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer lease.Close()

		if lease.SeedOwner() {
			if err := SeedRepoCollaboration(r.Context(), lease); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}

		defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

		connection.SetReadLimit(1 << 20)

		syncState, err := lease.Collaboration().Document.NewSyncState()
		if err != nil {
			return
		}

		defer func() { _ = syncState.Close() }()

		loopConfig := config
		loopConfig.ServerPeerID = "probo-gateway-" + connectionID

		loopConfig.DocumentID = documentID
		if loopConfig.Shutdown == nil {
			loopConfig.Shutdown = context.Background()
		}

		_ = ServeRepoCollaboration(r.Context(), connection, lease, syncState, loopConfig)
	}))

	t.Cleanup(server.Close)

	return server
}

type repoTestClient struct {
	document  *automerge.Document
	conn      *websocket.Conn
	ephemeral chan []byte
	send      chan []byte
	errs      chan error
}

func dialRepoClient(
	t *testing.T,
	ctx context.Context,
	url string,
	peerID string,
	repoDocumentID string,
) *repoTestClient {
	t.Helper()

	document, err := automerge.New(actorFromByte(len(peerID)))
	require.NoError(t, err)

	syncState, err := document.NewSyncState()
	require.NoError(t, err)

	driver, err := collabpkg.NewClientConn(collabpkg.ClientConfig{
		ClientPeerID: peerID,
		DocumentID:   repoDocumentID,
		StartsEmpty:  true,
	}, syncState)
	require.NoError(t, err)

	conn, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)

	conn.SetReadLimit(1 << 20)

	client := &repoTestClient{
		document:  document,
		conn:      conn,
		ephemeral: make(chan []byte, 8),
		send:      make(chan []byte, 8),
		errs:      make(chan error, 1),
	}

	frames := make(chan []byte, 8)

	go func() {
		for {
			kind, data, readErr := conn.Read(ctx)
			if readErr != nil {
				close(frames)
				return
			}

			if kind != websocket.MessageBinary {
				continue
			}

			select {
			case frames <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	write := func(payload []byte, messageType websocket.MessageType) error {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return conn.Write(writeCtx, messageType, payload)
	}

	go func() {
		join, startErr := driver.Start()
		if startErr != nil {
			client.errs <- startErr
			return
		}

		if err := write(join, websocket.MessageBinary); err != nil {
			client.errs <- err
			return
		}

		var count uint64

		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-frames:
				if !ok {
					return
				}

				inbound, receiveErr := driver.Receive(frame)
				if receiveErr != nil {
					client.errs <- receiveErr
					return
				}

				for _, outgoing := range inbound.Outgoing {
					if err := write(outgoing, websocket.MessageBinary); err != nil {
						client.errs <- err
						return
					}
				}

				if inbound.Ephemeral != nil {
					select {
					case client.ephemeral <- inbound.Ephemeral.Data:
					default:
					}
				}
			case payload := <-client.send:
				count++

				frame, ephemeralErr := driver.Ephemeral("session-"+peerID, count, payload)
				if ephemeralErr != nil {
					client.errs <- ephemeralErr
					return
				}

				if err := write(frame, websocket.MessageBinary); err != nil {
					client.errs <- err
					return
				}
			}
		}
	}()

	t.Cleanup(func() { _ = conn.Close(websocket.StatusNormalClosure, "") })

	return client
}

func actorFromByte(value int) automerge.ActorID {
	var actorID automerge.ActorID

	actorID[0] = byte(value + 1)

	return actorID
}

func (c *repoTestClient) waitForBody(t *testing.T, ctx context.Context, want string) {
	t.Helper()

	deadline := time.After(10 * time.Second)

	for {
		select {
		case err := <-c.errs:
			require.NoError(t, err)
		case <-deadline:
			require.Failf(t, "client did not converge", "wanted body %q", want)
		default:
		}

		text, err := c.document.Text("body")
		if err == nil {
			if value, err := text.String(); err == nil && value == want {
				return
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestServeRepoCollaboration_ConvergesRealDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{200})
	require.NoError(t, err)

	defer func() { _ = serverDocument.Close() }()

	text, err := serverDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "hello repo"))

	_, err = serverDocument.Commit("seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		opened: &collaboration.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID, gid.Nil), documentID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", repoDocumentID)
	client.waitForBody(t, ctx, "hello repo")
}

func TestServeRepoCollaboration_SeedsUnseededDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{202})
	require.NoError(t, err)

	defer func() { _ = serverDocument.Close() }()

	const seedContent = `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Seeded"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"body text"}]}]}`

	documents := &fakeDocumentCollaborationDocuments{
		opened: &collaboration.DocumentCollaboration{
			Document:    serverDocument,
			Revision:    1,
			NeedsSeed:   true,
			SeedContent: seedContent,
		},
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID, gid.Nil), documentID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", repoDocumentID)
	client.waitForBody(t, ctx, "\uFFFCSeeded\uFFFCbody text")
}

func TestServeRepoCollaboration_FansOutEphemeral(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{201})
	require.NoError(t, err)

	defer func() { _ = serverDocument.Close() }()

	text, err := serverDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "shared"))

	_, err = serverDocument.Commit("seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		opened: &collaboration.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID, gid.Nil), documentID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())

	first := dialRepoClient(t, ctx, wsURL, "client-first", repoDocumentID)
	first.waitForBody(t, ctx, "shared")

	second := dialRepoClient(t, ctx, wsURL, "client-second", repoDocumentID)
	second.waitForBody(t, ctx, "shared")

	payload := []byte("cursor-at-3")
	first.send <- payload

	select {
	case got := <-second.ephemeral:
		assert.Equal(t, payload, got)
	case err := <-second.errs:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		require.Fail(t, "the second client did not receive the gossiped ephemeral")
	}
}

func TestServeRepoCollaboration_StampsSelectionDisplayName(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{205})
	require.NoError(t, err)

	defer func() { _ = serverDocument.Close() }()

	text, err := serverDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "shared"))

	_, err = serverDocument.Commit("seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		opened: &collaboration.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServerWithConfig(
		t,
		hub,
		coredata.NewScope(tenantID, gid.Nil),
		documentID,
		RepoConfig{
			Shutdown:    context.Background(),
			DisplayName: "Ada Lovelace",
			IdentityID:  "idnt_ada",
		},
	)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())

	first := dialRepoClient(t, ctx, wsURL, "client-first", repoDocumentID)
	first.waitForBody(t, ctx, "shared")

	second := dialRepoClient(t, ctx, wsURL, "client-second", repoDocumentID)
	second.waitForBody(t, ctx, "shared")

	payload, err := collabpkg.MarshalPresenceValue(
		map[string]any{
			"kind":        "selection",
			"displayName": "spoofed",
			"selection": map[string]any{
				"field":  "body",
				"anchor": "cursor-a",
				"head":   "cursor-a",
			},
		},
	)
	require.NoError(t, err)

	first.send <- payload

	select {
	case got := <-second.ephemeral:
		name, ok := collabpkg.SelectionDisplayName(got)
		require.True(t, ok)
		assert.Equal(t, "Ada Lovelace", name)

		wrapper := collabpkg.PresenceMessage{
			Type:  collabpkg.PresenceUpdate,
			Value: got,
		}

		var decoded struct {
			Kind       string `cbor:"kind"`
			IdentityID string `cbor:"identityID"`
		}
		require.NoError(t, wrapper.UnmarshalValue(&decoded))
		assert.Equal(t, "selection", decoded.Kind)
		assert.Equal(t, "idnt_ada", decoded.IdentityID)
	case err := <-second.errs:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		require.Fail(t, "the second client did not receive the stamped selection")
	}
}

func TestServeRepoCollaboration_PublishesEphemeralCrossInstance(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{203})
	require.NoError(t, err)

	defer func() { _ = serverDocument.Close() }()

	text, err := serverDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "shared"))

	_, err = serverDocument.Commit("seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		opened: &collaboration.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := NewHub(documents, nil)

	published := make(chan []byte, 4)
	publisher := func(_ context.Context, frame []byte) error {
		select {
		case published <- append([]byte(nil), frame...):
		default:
		}

		return nil
	}

	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID, gid.Nil), documentID, publisher)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", repoDocumentID)
	client.waitForBody(t, ctx, "shared")

	payload := []byte("cursor-payload")
	client.send <- payload

	select {
	case frame := <-published:
		message, err := collabpkg.DecodeMessage(frame)
		require.NoError(t, err)
		assert.Equal(t, collabpkg.MessageEphemeral, message.Type)
		assert.Equal(t, payload, message.Data)
	case <-time.After(10 * time.Second):
		require.Fail(t, "the loop did not publish the ephemeral across instances")
	}
}
