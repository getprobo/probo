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
		CanUpdate:        true,
		CanComment:       true,
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

		connection, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

		connection.SetReadLimit(1 << 20)

		collaborationCtx, cancelCollaboration := context.WithCancel(r.Context())
		defer cancelCollaboration()

		incoming := make(chan Incoming, 1)
		go func() {
			_ = ReadMessages(collaborationCtx, connection, incoming)

			cancelCollaboration()
		}()

		serverPeerID := "probo-gateway-" + connectionID

		session, accepted, joinErr := acceptRepoJoin(
			collaborationCtx,
			connection,
			incoming,
			serverPeerID,
		)
		if joinErr != nil || !accepted {
			if joinErr != nil {
				_ = connection.Close(websocket.StatusInternalError, "collaboration failed")
			}

			return
		}

		lease, err := hub.Acquire(collaborationCtx, scope, documentID, connectionID)
		if err != nil {
			_ = connection.Close(websocket.StatusInternalError, "collaboration failed")
			return
		}

		defer lease.Close()

		syncState, err := lease.Collaboration().Document.NewSyncState()
		if err != nil {
			return
		}

		defer func() { _ = syncState.Close() }()

		loopConfig := config
		loopConfig.ServerPeerID = serverPeerID
		loopConfig.Session = session
		loopConfig.DocumentID = documentID
		loopConfig.Incoming = incoming

		if loopConfig.Shutdown == nil {
			loopConfig.Shutdown = context.Background()
		}

		_ = ServeRepoCollaboration(collaborationCtx, connection, lease, syncState, loopConfig)
	}))

	t.Cleanup(server.Close)

	return server
}

func TestRepoServer_CanceledWhenClientDisconnectsDuringOpen(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	documents := &cancelProbeDocuments{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServer(
		t,
		hub,
		coredata.NewScope(tenantID, gid.Nil),
		documentID,
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	connection, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.CloseNow() })

	writeRepoJoin(t, ctx, connection, "client-cancel")

	select {
	case <-documents.started:
	case <-ctx.Done():
		require.Fail(t, "collaboration open did not start")
	}

	require.NoError(t, connection.CloseNow())

	select {
	case <-documents.openCtx.Done():
		assert.ErrorIs(t, documents.openCtx.Err(), context.Canceled)
	case <-ctx.Done():
		require.Fail(t, "client disconnect did not cancel collaboration open")
	}
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

func writeRepoJoin(
	t *testing.T,
	ctx context.Context,
	connection *websocket.Conn,
	peerID string,
) {
	t.Helper()

	frame, err := collabpkg.EncodeJoinFrame(
		collabpkg.NewJoinFrame(peerID, collabpkg.PeerMetadata{}),
	)
	require.NoError(t, err)

	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	require.NoError(t, connection.Write(writeCtx, websocket.MessageBinary, frame))
}

func TestRepoServer_RepliesToJoinBeforeDocumentOpen(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	documents := &cancelProbeDocuments{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	t.Cleanup(func() {
		select {
		case <-documents.release:
		default:
			close(documents.release)
		}
	})

	hub := NewHub(documents, nil)
	server := newRepoTestServer(
		t,
		hub,
		coredata.NewScope(tenantID, gid.Nil),
		documentID,
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	connection, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.CloseNow() })

	writeRepoJoin(t, ctx, connection, "client-far")

	readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
	defer readCancel()

	_, data, err := connection.Read(readCtx)
	require.NoError(t, err)

	peer, err := collabpkg.DecodePeerFrame(data)
	require.NoError(t, err)
	assert.Equal(t, collabpkg.ProtocolV1, peer.SelectedProtocolVersion)
	assert.Equal(t, "client-far", peer.TargetID)

	select {
	case <-documents.started:
	case <-ctx.Done():
		require.Fail(t, "collaboration open did not start after the peer reply")
	}
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
			CanUpdate:   true,
			CanComment:  true,
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

func TestServeRepoCollaboration_ReadOnlyPeerCannotChangeDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	serverDocument, err := automerge.New(automerge.ActorID{204})
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverDocument.Close() })

	text, err := serverDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "kept"))

	_, err = serverDocument.Commit("seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	var recorded atomic.Int32

	persistedCh := make(chan struct{}, 32)
	documents := &fakeDocumentCollaborationDocuments{
		opened:      &collaboration.DocumentCollaboration{Document: serverDocument, Revision: 1},
		persistedCh: persistedCh,
	}
	hub := NewHub(documents, nil)
	server := newRepoTestServerWithConfig(
		t,
		hub,
		coredata.NewScope(tenantID, gid.Nil),
		documentID,
		RepoConfig{
			Shutdown: context.Background(),
			RecordActors: func(context.Context, [][]byte) error {
				recorded.Add(1)

				return nil
			},
		},
	)

	repoDocumentID := collabpkg.DeriveDocumentID(documentID.String())
	clientDocument, err := automerge.New(actorFromByte(7))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientDocument.Close() })

	syncState, err := clientDocument.NewSyncState()
	require.NoError(t, err)
	t.Cleanup(func() { _ = syncState.Close() })

	driver, err := collabpkg.NewClientConn(
		collabpkg.ClientConfig{
			ClientPeerID: "reader",
			DocumentID:   repoDocumentID,
			StartsEmpty:  true,
		},
		syncState,
	)
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.CloseNow() })
	conn.SetReadLimit(1 << 20)

	frames := make(chan []byte, 8)

	go func() {
		for {
			_, data, readErr := conn.Read(ctx)
			if readErr != nil {
				close(frames)
				return
			}

			select {
			case frames <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	write := func(payload []byte) {
		t.Helper()

		writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
		defer writeCancel()

		require.NoError(t, conn.Write(writeCtx, websocket.MessageBinary, payload))
	}
	nextFrame := func(timeout time.Duration) ([]byte, bool) {
		t.Helper()

		select {
		case frame, ok := <-frames:
			return frame, ok
		case <-time.After(timeout):
			return nil, false
		}
	}
	exchange := func(frame []byte) {
		t.Helper()

		inbound, receiveErr := driver.Receive(frame)
		require.NoError(t, receiveErr)

		for _, outgoing := range inbound.Outgoing {
			write(outgoing)
		}
	}

	join, err := driver.Start()
	require.NoError(t, err)
	write(join)

	deadline := time.Now().Add(10 * time.Second)

	for {
		body, bodyErr := clientDocument.Text("body")
		if bodyErr == nil {
			value, valueErr := body.String()
			if valueErr == nil && value == "kept" {
				break
			}
		}

		require.True(t, time.Now().Before(deadline), "reader did not receive the document")

		frame, ok := nextFrame(2 * time.Second)
		require.True(t, ok)
		exchange(frame)
	}

	for {
		frame, ok := nextFrame(300 * time.Millisecond)
		if !ok {
			break
		}

		exchange(frame)
	}

	body, err := clientDocument.Text("body")
	require.NoError(t, err)
	require.NoError(t, body.Splice(4, 0, "!"))
	require.NoError(t, clientDocument.PutString("note", "secret"))

	_, err = clientDocument.Commit("tamper", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	outgoing, err := driver.SyncChanged()
	require.NoError(t, err)
	require.NotEmpty(t, outgoing)

	drain := true
	for drain {
		select {
		case <-persistedCh:
		default:
			drain = false
		}
	}

	for _, frame := range outgoing {
		write(frame)
	}

	select {
	case <-persistedCh:
	case <-time.After(5 * time.Second):
		require.Fail(t, "server did not handle the rejected edit")
	}

	serverText, err := serverDocument.Text("body")
	require.NoError(t, err)

	serverBody, err := serverText.String()
	require.NoError(t, err)
	assert.Equal(t, "kept", serverBody)

	_, err = serverDocument.String("note")
	assert.Error(t, err)
	assert.Equal(t, int32(0), recorded.Load())
}
