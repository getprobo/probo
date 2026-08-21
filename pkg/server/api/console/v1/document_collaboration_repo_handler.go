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
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/jsonx"
)

// emptyRichEditorDocument is the ProseMirror document a version with no stored
// content seeds from, matching the frontend's default empty editor.
const emptyRichEditorDocument = `{"type":"doc","content":[{"type":"paragraph"}]}`

const documentCollaborationRepoRefreshInterval = 500 * time.Millisecond

// repoCollaborationRefresher pulls changes another server instance persisted
// into the shared in-memory document, so a NOTIFY-driven wake or the periodic
// tick can re-sync connected peers. *probo.DocumentService satisfies it.
type repoCollaborationRefresher interface {
	RefreshCollaboration(
		ctx context.Context,
		scope coredata.Scoper,
		documentVersionID gid.GID,
		document *automerge.Document,
		knownRevision int64,
	) (int64, bool, error)
}

// repoCollaborationConfig carries the per-connection dependencies of the repo
// collaboration loop. refresher may be nil, which disables cross-instance
// refresh (used by the in-process convergence test that has no database).
type repoCollaborationConfig struct {
	serverPeerID      string
	refresher         repoCollaborationRefresher
	scope             coredata.Scoper
	documentVersionID gid.GID
	logger            *log.Logger
	shutdown          context.Context
	// publishEphemeral relays a gossip frame to other server instances. It may be
	// nil, which limits ephemeral fan-out to this instance.
	publishEphemeral func(ctx context.Context, frame []byte) error
}

// handleRepo serves one document version over the automerge-repo protocol. It
// mirrors handle (parse, authorize, acquire a room lease), then drives the
// protocol with the transport-agnostic ServerConn and fans presence/sync out
// through the shared hub. It is mounted alongside the custom /sync route so the
// two protocols coexist during migration; see
// pkg/automerge/collaboration/GATEWAY_CONTRACT.md.
func (h *documentCollaborationHandler) handleRepo(w http.ResponseWriter, r *http.Request) {
	documentVersionIDString := chi.URLParam(r, "documentVersionID")

	documentVersionID, err := gid.ParseGID(documentVersionIDString)
	if err != nil || documentVersionID.EntityType() != coredata.DocumentVersionEntityType {
		jsonx.RenderNotFound(w, fmt.Errorf("document version not found"))
		return
	}

	scope, err := h.authorize(r.Context(), documentVersionID)
	if err != nil {
		h.renderAuthorizationError(w, err)
		return
	}

	connectionID, err := newDocumentCollaborationConnectionID()
	if err != nil {
		jsonx.RenderInternalServerError(w)
		return
	}

	lease, err := h.hub.acquire(r.Context(), scope, documentVersionID, connectionID)
	if err != nil {
		h.renderServiceError(w, r, documentVersionIDString, err)
		return
	}

	defer lease.Close()

	// The repo protocol has no seed handshake, so the server is authoritative for
	// seeding: the connection that claimed the seed converts the version's stored
	// ProseMirror content into the CRDT before serving it. The persist that
	// follows marks the state seeded, so later connections skip this.
	if lease.SeedOwner() {
		if err := seedRepoCollaboration(r.Context(), lease); err != nil {
			h.renderServiceError(w, r, documentVersionIDString, err)
			return
		}
	}

	connection, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{
			OriginPatterns:  h.allowedOrigins,
			CompressionMode: websocket.CompressionDisabled,
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			r.Context(),
			"cannot accept document collaboration repo connection",
			log.Error(err),
			log.String("document_version_id", documentVersionIDString),
		)

		return
	}

	defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

	connection.SetReadLimit(documentCollaborationMessageMaxBytes)

	syncState, err := lease.Collaboration().Document.NewSyncState(r.Context())
	if err != nil {
		h.closeWithError(r.Context(), connection, documentVersionIDString, err)
		return
	}

	defer func() { _ = syncState.Close(context.Background()) }()

	config := repoCollaborationConfig{
		serverPeerID:      "probo-gateway-" + connectionID,
		refresher:         h.probo.Documents,
		scope:             scope,
		documentVersionID: documentVersionID,
		logger:            h.logger,
		shutdown:          h.shutdown,
		publishEphemeral: func(ctx context.Context, frame []byte) error {
			return h.probo.Documents.NotifyCollaborationEphemeral(
				ctx,
				documentVersionID,
				h.hub.instanceID,
				frame,
			)
		},
	}

	if err := serveRepoCollaboration(r.Context(), connection, lease, syncState, config); err != nil {
		h.closeWithError(r.Context(), connection, documentVersionIDString, err)
	}
}

// seedRepoCollaboration converts the version's stored ProseMirror content into
// Automerge rich-text spans and writes them into the shared document, then
// schedules a persist. It is called once per document, by the connection that
// claimed the seed, and is safe if the body already exists (it reuses it rather
// than recreating it).
func seedRepoCollaboration(ctx context.Context, lease *documentCollaborationRoomLease) error {
	content := lease.Collaboration().SeedContent
	if content == "" {
		content = emptyRichEditorDocument
	}

	spans, err := automergeprosemirror.ToSpans(content)
	if err != nil {
		return fmt.Errorf("cannot convert seed content to spans: %w", err)
	}

	document := lease.Collaboration().Document

	text, err := document.Text(ctx, "body")
	if err != nil {
		text, err = document.CreateText(ctx, "body")
		if err != nil {
			return fmt.Errorf("cannot create seed text object: %w", err)
		}
	}

	if err := text.UpdateSpans(ctx, spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return fmt.Errorf("cannot write seed spans: %w", err)
	}

	if _, err := document.Commit(ctx, "Seed collaboration document", time.Now()); err != nil {
		return fmt.Errorf("cannot commit seed: %w", err)
	}

	lease.SchedulePersist()

	return nil
}

// serveRepoCollaboration runs the automerge-repo protocol for one connection. It
// is transport-focused glue over the tested ServerConn driver and the hub's
// fan-out primitives, kept as a standalone function so it can be driven by a
// real ClientConn in tests without a database. It returns nil on a clean close
// and an error only on an unexpected failure.
func serveRepoCollaboration(
	ctx context.Context,
	connection *websocket.Conn,
	lease *documentCollaborationRoomLease,
	syncState *automerge.SyncState,
	config repoCollaborationConfig,
) error {
	conn, err := collaboration.NewAdoptingServerConn(
		collaboration.ServerConfig{ServerPeerID: config.serverPeerID},
		syncState,
	)
	if err != nil {
		return fmt.Errorf("cannot create repo server connection: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	incoming := make(chan documentCollaborationIncoming, 1)
	go readCollaborationMessages(ctx, connection, incoming)

	// The first frame is the client's join.
	var join documentCollaborationIncoming
	select {
	case join = <-incoming:
	case <-ctx.Done():
		return nil
	}

	if join.Err != nil {
		return repoReadResult(join.Err)
	}

	if join.MessageType != websocket.MessageBinary {
		_ = connection.Close(websocket.StatusUnsupportedData, "expected a binary join frame")
		return nil
	}

	out, accepted, err := conn.Start(ctx, join.Data)
	if err != nil {
		return fmt.Errorf("cannot start repo server connection: %w", err)
	}

	if err := writeRepoFrames(ctx, connection, out); err != nil {
		return err
	}

	if !accepted {
		_ = connection.Close(websocket.StatusPolicyViolation, "unsupported protocol version")
		return nil
	}

	revision := lease.Revision()

	var tick <-chan time.Time
	if config.refresher != nil {
		ticker := time.NewTicker(documentCollaborationRepoRefreshInterval)
		defer ticker.Stop()

		tick = ticker.C
	}

	for {
		select {
		case <-config.shutdown.Done():
			_ = connection.CloseNow()
			return nil
		case <-ctx.Done():
			return nil
		case message := <-incoming:
			if message.Err != nil {
				return repoReadResult(message.Err)
			}

			// automerge-repo frames are always binary; ignore anything else.
			if message.MessageType != websocket.MessageBinary {
				continue
			}

			reply, fanout, err := conn.Receive(ctx, message.Data)
			if err != nil {
				return fmt.Errorf("cannot process repo frame: %w", err)
			}

			if err := writeRepoFrames(ctx, connection, reply); err != nil {
				return err
			}

			if fanout != nil {
				lease.BroadcastEphemeral(fanout)

				// Relay to peers on other instances. A failure here (including an
				// oversized frame) must not drop the connection: local peers
				// already have the frame and the sender re-emits its state.
				if config.publishEphemeral != nil {
					if err := config.publishEphemeral(ctx, fanout); err != nil && config.logger != nil {
						config.logger.WarnCtx(
							ctx,
							"cannot relay collaboration ephemeral across instances",
							log.Error(err),
							log.String("document_version_id", config.documentVersionID.String()),
						)
					}
				}

				continue
			}

			// A sync frame may have advanced the shared document: wake the other
			// peers so they re-sync, and schedule a debounced persist.
			lease.NotifyPeers()

			if err := lease.PersistError(); err != nil {
				return fmt.Errorf("cannot persist repo collaboration: %w", err)
			}

			lease.SchedulePersist()
		case frame := <-lease.Ephemeral:
			if err := writeCollaborationMessage(ctx, connection, websocket.MessageBinary, frame); err != nil {
				return err
			}
		case wake := <-lease.Wake:
			if err := lease.PersistError(); err != nil {
				return fmt.Errorf("cannot persist repo collaboration: %w", err)
			}

			if wake.refresh && config.refresher != nil {
				var changed bool

				revision, changed, err = config.refresher.RefreshCollaboration(
					ctx,
					config.scope,
					config.documentVersionID,
					lease.Collaboration().Document,
					revision,
				)
				if err != nil {
					return fmt.Errorf("cannot refresh repo collaboration: %w", err)
				}

				lease.SetRevision(revision)

				if !changed {
					continue
				}
			}

			frames, err := conn.SyncChanged(ctx)
			if err != nil {
				return fmt.Errorf("cannot generate repo sync frames: %w", err)
			}

			if err := writeRepoFrames(ctx, connection, frames); err != nil {
				return err
			}
		case <-tick:
			if err := lease.PersistError(); err != nil {
				return fmt.Errorf("cannot persist repo collaboration: %w", err)
			}

			var changed bool

			revision, changed, err = config.refresher.RefreshCollaboration(
				ctx,
				config.scope,
				config.documentVersionID,
				lease.Collaboration().Document,
				revision,
			)
			if err != nil {
				return fmt.Errorf("cannot refresh repo collaboration: %w", err)
			}

			if !changed {
				continue
			}

			lease.SetRevision(revision)

			frames, err := conn.SyncChanged(ctx)
			if err != nil {
				return fmt.Errorf("cannot generate repo sync frames: %w", err)
			}

			if err := writeRepoFrames(ctx, connection, frames); err != nil {
				return err
			}
		}
	}
}

// repoReadResult maps a websocket read error to a loop result: a normal or
// going-away close is a clean disconnect (nil), anything else is an error.
func repoReadResult(err error) error {
	switch websocket.CloseStatus(err) {
	case websocket.StatusNormalClosure, websocket.StatusGoingAway:
		return nil
	default:
		return fmt.Errorf("repo collaboration read failed: %w", err)
	}
}

func writeRepoFrames(ctx context.Context, connection *websocket.Conn, frames [][]byte) error {
	for _, frame := range frames {
		if err := writeCollaborationMessage(ctx, connection, websocket.MessageBinary, frame); err != nil {
			return err
		}
	}

	return nil
}
