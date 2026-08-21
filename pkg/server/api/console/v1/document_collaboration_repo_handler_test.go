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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
)

// newRepoTestServer stands up an httptest WebSocket server that runs the repo
// collaboration loop against a shared in-memory document, with no database. Each
// connection acquires a lease from the hub, so sync and ephemeral fan-out go
// through the real room machinery.
func newRepoTestServer(
	t *testing.T,
	hub *documentCollaborationHub,
	scope coredata.Scoper,
	versionID gid.GID,
	publishEphemeral func(ctx context.Context, frame []byte) error,
) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionID, err := newDocumentCollaborationConnectionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lease, err := hub.acquire(r.Context(), scope, versionID, connectionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer lease.Close()

		if lease.SeedOwner() {
			if err := seedRepoCollaboration(r.Context(), lease); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // test-only: no Origin from the Go client
		})
		if err != nil {
			return
		}

		defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

		connection.SetReadLimit(1 << 20)

		syncState, err := lease.Collaboration().Document.NewSyncState(r.Context())
		if err != nil {
			return
		}

		defer func() { _ = syncState.Close(context.Background()) }()

		_ = serveRepoCollaboration(r.Context(), connection, lease, syncState, repoCollaborationConfig{
			serverPeerID:      "probo-gateway-" + connectionID,
			documentVersionID: versionID,
			shutdown:          context.Background(),
			publishEphemeral:  publishEphemeral,
		})
	}))

	t.Cleanup(server.Close)

	return server
}

// repoTestClient drives a real ClientConn over a WebSocket. A single reader
// goroutine owns the connection: it reads server frames, feeds them to the
// driver, and writes the driver's replies plus any ephemerals the test asks to
// send, so there is never a concurrent writer.
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
	documentID string,
) *repoTestClient {
	t.Helper()

	document, err := automerge.New(ctx, actorFromByte(len(peerID)))
	require.NoError(t, err)

	syncState, err := document.NewSyncState(ctx)
	require.NoError(t, err)

	driver, err := collaboration.NewClientConn(collaboration.ClientConfig{
		ClientPeerID: peerID,
		DocumentID:   documentID,
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

				inbound, receiveErr := driver.Receive(ctx, frame)
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

		text, err := c.document.Text(ctx, "body")
		if err == nil {
			if value, err := text.String(ctx); err == nil && value == want {
				return
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// TestServeRepoCollaboration_ConvergesRealDocument connects a real repo client to
// the gateway loop and confirms it materializes the server's seeded document.
// This exercises the whole server-side stack end to end without a database:
// handshake, document-id adoption, and the sync loop over the hub lease.
func TestServeRepoCollaboration_ConvergesRealDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)

	serverDocument, err := automerge.New(ctx, automerge.ActorID{200})
	require.NoError(t, err)
	defer func() { _ = serverDocument.Close(ctx) }()

	text, err := serverDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "hello repo"))
	_, err = serverDocument.Commit(ctx, "seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		collaboration: &probo.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := newDocumentCollaborationHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID), versionID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	documentID := collaboration.DeriveDocumentID(versionID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", documentID)
	client.waitForBody(t, ctx, "hello repo")
}

// TestServeRepoCollaboration_SeedsUnseededDocument confirms the server seeds an
// unseeded version from its stored ProseMirror content: the connection that
// claims the seed converts the content to spans, and a repo client then
// materializes it. This is the whole server-authoritative seeding path.
func TestServeRepoCollaboration_SeedsUnseededDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)

	// A fresh, empty document with no body: exactly what OpenCollaboration
	// returns for a version that has never been seeded.
	serverDocument, err := automerge.New(ctx, automerge.ActorID{202})
	require.NoError(t, err)
	defer func() { _ = serverDocument.Close(ctx) }()

	const seedContent = `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Seeded"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"body text"}]}]}`

	documents := &fakeDocumentCollaborationDocuments{
		collaboration: &probo.DocumentCollaboration{
			Document:    serverDocument,
			Revision:    1,
			NeedsSeed:   true,
			SeedContent: seedContent,
		},
	}
	hub := newDocumentCollaborationHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID), versionID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	documentID := collaboration.DeriveDocumentID(versionID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", documentID)
	// The flat text of the seeded document is the concatenation of its blocks.
	client.waitForBody(t, ctx, "Seededbody text")
}

// TestServeRepoCollaboration_FansOutEphemeral connects two repo clients and
// confirms an ephemeral one sends is gossiped to the other through the hub, which
// is how repo presence and cursors travel.
func TestServeRepoCollaboration_FansOutEphemeral(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)

	serverDocument, err := automerge.New(ctx, automerge.ActorID{201})
	require.NoError(t, err)
	defer func() { _ = serverDocument.Close(ctx) }()

	text, err := serverDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "shared"))
	_, err = serverDocument.Commit(ctx, "seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		collaboration: &probo.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := newDocumentCollaborationHub(documents, nil)
	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID), versionID, nil)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	documentID := collaboration.DeriveDocumentID(versionID.String())

	// Both clients converge first, which guarantees both handshakes completed.
	first := dialRepoClient(t, ctx, wsURL, "client-first", documentID)
	first.waitForBody(t, ctx, "shared")

	second := dialRepoClient(t, ctx, wsURL, "client-second", documentID)
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

// TestServeRepoCollaboration_PublishesEphemeralCrossInstance confirms the loop
// hands each gossiped frame to the cross-instance publisher, which is how
// presence and cursors reach peers on other server instances.
func TestServeRepoCollaboration_PublishesEphemeralCrossInstance(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)

	serverDocument, err := automerge.New(ctx, automerge.ActorID{203})
	require.NoError(t, err)
	defer func() { _ = serverDocument.Close(ctx) }()

	text, err := serverDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "shared"))
	_, err = serverDocument.Commit(ctx, "seed", time.Unix(1786147200, 0).UTC())
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		collaboration: &probo.DocumentCollaboration{Document: serverDocument, Revision: 1},
	}
	hub := newDocumentCollaborationHub(documents, nil)

	published := make(chan []byte, 4)
	publisher := func(_ context.Context, frame []byte) error {
		select {
		case published <- append([]byte(nil), frame...):
		default:
		}

		return nil
	}

	server := newRepoTestServer(t, hub, coredata.NewScope(tenantID), versionID, publisher)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	documentID := collaboration.DeriveDocumentID(versionID.String())

	client := dialRepoClient(t, ctx, wsURL, "client-a", documentID)
	client.waitForBody(t, ctx, "shared")

	payload := []byte("cursor-payload")
	client.send <- payload

	select {
	case frame := <-published:
		message, err := collaboration.DecodeMessage(frame)
		require.NoError(t, err)
		assert.Equal(t, collaboration.MessageEphemeral, message.Type)
		assert.Equal(t, payload, message.Data)
	case <-time.After(10 * time.Second):
		require.Fail(t, "the loop did not publish the ephemeral across instances")
	}
}
