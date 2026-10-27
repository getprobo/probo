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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

const (
	MessageMaxBytes = 1024 * 1024
	WriteTimeout    = 10 * time.Second
)

const (
	RepoRefreshInterval    = 500 * time.Millisecond
	DepartureNotifyTimeout = 2 * time.Second
)

type (
	Handler struct {
		Logger         *log.Logger
		Probo          *probo.Service
		IAM            *iam.Service
		BaseURL        *baseurl.BaseURL
		AllowedOrigins []string
		Hub            *Hub
		Shutdown       context.Context
	}

	Incoming struct {
		MessageType websocket.MessageType
		Data        []byte
		Err         error
	}

	// RepoRefresher pulls changes another server instance persisted into
	// the shared in-memory document.
	RepoRefresher interface {
		Refresh(
			ctx context.Context,
			scope coredata.Scoper,
			documentID gid.GID,
			document *automerge.Document,
			knownRevision int64,
		) (int64, bool, error)
	}

	RepoConfig struct {
		ServerPeerID     string
		Refresher        RepoRefresher
		Scope            coredata.Scoper
		DocumentID       gid.GID
		Logger           *log.Logger
		Shutdown         context.Context
		PublishEphemeral func(ctx context.Context, frame []byte) error
		DisplayName      string
		IdentityID       string
		AvatarURL        string
		RecordActors     func(ctx context.Context, actors [][]byte) error
		Incoming         <-chan Incoming
	}
)

func (h *Handler) HandleRepo(w http.ResponseWriter, r *http.Request) {
	documentIDString := chi.URLParam(r, "documentID")

	documentID, err := gid.ParseGID(documentIDString)
	if err != nil || documentID.EntityType() != coredata.DocumentEntityType {
		jsonx.RenderNotFound(w, fmt.Errorf("document not found"))
		return
	}

	scope, document, err := h.authorize(r.Context(), documentID)
	if err != nil {
		h.renderAuthorizationError(w, err)
		return
	}

	identity := authn.IdentityFromContext(r.Context())
	displayName, identityID, avatarURL, profileID := collaborationPresenceIdentity(
		r.Context(),
		h,
		document,
		identity,
	)

	connectionID, err := NewConnectionID()
	if err != nil {
		jsonx.RenderInternalServerError(w)
		return
	}

	// Upgrade before loading the collaboration row. Until hijack, net/http
	// cancels the request when the client drops, and that aborts the SELECT.
	connection, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{
			OriginPatterns:  h.AllowedOrigins,
			CompressionMode: websocket.CompressionDisabled,
		},
	)
	if err != nil {
		h.Logger.WarnCtx(
			r.Context(),
			"cannot accept collaboration repo connection",
			log.Error(err),
			log.String("document_id", documentIDString),
		)

		return
	}

	defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

	connection.SetReadLimit(MessageMaxBytes)

	collaborationCtx, cancelCollaboration := context.WithCancel(r.Context())
	defer cancelCollaboration()

	incoming := make(chan Incoming, 1)
	go func() {
		_ = ReadMessages(collaborationCtx, connection, incoming)

		cancelCollaboration()
	}()

	lease, err := h.Hub.Acquire(collaborationCtx, scope, documentID, connectionID)
	if err != nil {
		h.closeOpenError(collaborationCtx, connection, documentIDString, err)
		return
	}

	defer lease.Close()

	if lease.SeedOwner() {
		if err := SeedRepoCollaboration(collaborationCtx, lease); err != nil {
			h.releaseSeedClaim(collaborationCtx, scope, documentID, lease)
			h.closeOpenError(collaborationCtx, connection, documentIDString, err)

			return
		}
	}

	syncState, err := lease.Collaboration().Document.NewSyncState()
	if err != nil {
		h.closeWithError(r.Context(), connection, documentIDString, err)
		return
	}

	defer func() { _ = syncState.Close() }()

	config := RepoConfig{
		ServerPeerID: "probo-gateway-" + connectionID,
		Refresher:    h.Probo.Collaboration,
		Scope:        scope,
		DocumentID:   documentID,
		Logger:       h.Logger,
		Shutdown:     h.Shutdown,
		DisplayName:  displayName,
		IdentityID:   identityID,
		AvatarURL:    avatarURL,
		Incoming:     incoming,
		PublishEphemeral: func(ctx context.Context, frame []byte) error {
			return h.Probo.Collaboration.NotifyEphemeral(
				ctx,
				documentID,
				h.Hub.InstanceID(),
				frame,
			)
		},
	}

	if identity != nil {
		organizationID := document.OrganizationID
		authenticatedIdentityID := identity.ID
		authenticatedProfileID := profileID

		config.RecordActors = func(ctx context.Context, actors [][]byte) error {
			return h.Probo.Collaboration.RecordActors(
				ctx,
				scope,
				documentID,
				organizationID,
				authenticatedIdentityID,
				authenticatedProfileID,
				actors,
			)
		}
	}

	if err := ServeRepoCollaboration(collaborationCtx, connection, lease, syncState, config); err != nil {
		h.closeWithError(collaborationCtx, connection, documentIDString, err)
	}
}

func (h *Handler) authorize(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Scope, *coredata.Document, error) {
	identity := authn.IdentityFromContext(ctx)
	session := authn.SessionFromContext(ctx)

	scope := coredata.NewScopeFromObjectID(documentID)

	document, err := h.Probo.Documents.LoadDocument(ctx, scope, documentID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load document: %w", err)
	}

	params := iam.AuthorizeParams{
		Principal:          identity.ID,
		Resource:           document.ID,
		Action:             probo.ActionDocumentUpdate,
		ResourceAttributes: make(map[string]string),
	}
	if session != nil {
		params.Session = &session.ID
	}

	authScope, err := h.IAM.Authorizer.Authorize(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot authorize document update: %w", err)
	}

	return authScope, document, nil
}

func (h *Handler) releaseSeedClaim(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	lease *RoomLease,
) {
	releaseCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		PersistTimeout,
	)
	defer cancel()

	err := h.Probo.Collaboration.ReleaseSeed(
		releaseCtx,
		scope,
		documentID,
		lease.SeedClaimedAt(),
	)
	if err != nil {
		if h.Logger != nil {
			h.Logger.ErrorCtx(
				releaseCtx,
				"cannot release collaboration seed",
				log.Error(err),
				log.String("document_id", documentID.String()),
			)
		}

		return
	}

	lease.markSeedReleased()
}

func (h *Handler) renderAuthorizationError(w http.ResponseWriter, err error) {
	if scopeErr, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		bearertoken.SetBearerInsufficientScope(w, h.BaseURL, scopeErr.Scopes...)
		jsonx.RenderForbidden(w)

		return
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		jsonx.RenderForbidden(w)
		return
	}

	jsonx.RenderNotFound(w, fmt.Errorf("document not found"))
}

func (h *Handler) closeOpenError(
	ctx context.Context,
	connection *websocket.Conn,
	documentID string,
	err error,
) {
	if requestCanceled(err) {
		_ = connection.Close(websocket.StatusGoingAway, "")
		return
	}

	if collaborationOpenUnavailable(err) {
		_ = connection.Close(websocket.StatusPolicyViolation, "collaboration unavailable")
		return
	}

	h.Logger.ErrorCtx(
		ctx,
		"cannot open document collaboration",
		log.Error(err),
		log.String("document_id", documentID),
	)

	_ = connection.Close(websocket.StatusInternalError, "collaboration failed")
}

func collaborationOpenUnavailable(err error) bool {
	if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, collaboration.ErrSeedInProgress) {
		return true
	}

	if _, ok := errors.AsType[*collaboration.ErrDocumentArchived](err); ok {
		return true
	}

	if _, ok := errors.AsType[*collaboration.ErrDocumentVersionGenerated](err); ok {
		return true
	}

	if _, ok := errors.AsType[*collaboration.ErrCollaborationStateTooLarge](err); ok {
		return true
	}

	return false
}

func (h *Handler) closeWithError(
	ctx context.Context,
	connection *websocket.Conn,
	documentID string,
	err error,
) {
	h.Logger.ErrorCtx(
		ctx,
		"document collaboration failed",
		log.Error(err),
		log.String("document_id", documentID),
	)

	_ = connection.Close(websocket.StatusInternalError, "collaboration failed")
}

func collaborationPresenceIdentity(
	ctx context.Context,
	h *Handler,
	document *coredata.Document,
	identity *coredata.Identity,
) (displayName string, identityID string, avatarURL string, profileID *gid.GID) {
	if identity == nil {
		return "", "", "", nil
	}

	identityID = identity.ID.String()
	avatarURL = collaborationPresenceAvatarURL(h, identity)

	profile, err := h.IAM.OrganizationService.GetProfileForIdentityAndOrganization(
		ctx,
		identity.ID,
		document.OrganizationID,
	)
	if err == nil {
		name := strings.TrimSpace(profile.FullName)
		if name == "" {
			name = identity.EmailAddress.String()
		}

		return name, identityID, avatarURL, &profile.ID
	}

	if _, ok := errors.AsType[*iam.ErrProfileNotFound](err); !ok && h.Logger != nil {
		h.Logger.WarnCtx(
			ctx,
			"cannot load collaboration presence profile",
			log.Error(err),
			log.String("document_id", document.ID.String()),
			log.String("identity_id", identityID),
		)
	}

	name := strings.TrimSpace(identity.FullName)
	if name == "" {
		name = identity.EmailAddress.String()
	}

	return name, identityID, avatarURL, nil
}

func collaborationPresenceAvatarURL(
	h *Handler,
	identity *coredata.Identity,
) string {
	if identity.AvatarFileID == nil || h.BaseURL == nil {
		return ""
	}

	avatarURL, err := h.BaseURL.WithPath(
		"/api/files/v1/public/" + identity.AvatarFileID.String(),
	).String()
	if err != nil {
		return ""
	}

	return avatarURL
}

func SeedRepoCollaboration(ctx context.Context, lease *RoomLease) error {
	content := lease.Collaboration().SeedContent
	if content == "" {
		content = collaboration.EmptyProseMirrorDocument
	}

	spans, err := automergeprosemirror.ToSpans(content)
	if err != nil {
		return fmt.Errorf("cannot convert seed content to spans: %w", err)
	}

	document := lease.Collaboration().Document

	text, err := document.Text("body")
	if err != nil {
		text, err = document.CreateText("body")
		if err != nil {
			return fmt.Errorf("cannot create seed text object: %w", err)
		}
	}

	if err := text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return fmt.Errorf("cannot write seed spans: %w", err)
	}

	if _, err := document.Commit("Seed collaboration document", time.Now()); err != nil {
		return fmt.Errorf("cannot commit seed: %w", err)
	}

	lease.SchedulePersist()
	lease.completeSeed()

	return nil
}

func announceCollaborationDeparture(
	ctx context.Context,
	conn *collabpkg.ServerConn,
	lease *RoomLease,
	config RepoConfig,
) {
	senderID := conn.RemotePeerID()

	documentID := conn.DocumentID()
	if senderID == "" || documentID == "" {
		return
	}

	sessionID, err := NewConnectionID()
	if err != nil {
		return
	}

	frame, err := collabpkg.EncodeDepartureEphemeral(
		senderID,
		config.ServerPeerID,
		documentID,
		sessionID,
	)
	if err != nil {
		if config.Logger != nil {
			config.Logger.WarnCtx(
				ctx,
				"cannot encode collaboration departure",
				log.Error(err),
				log.String("document_id", config.DocumentID.String()),
			)
		}

		return
	}

	lease.BroadcastEphemeral(frame)

	if config.PublishEphemeral == nil {
		return
	}

	notifyCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		DepartureNotifyTimeout,
	)
	defer cancel()

	if err := config.PublishEphemeral(notifyCtx, frame); err != nil && config.Logger != nil {
		config.Logger.WarnCtx(
			notifyCtx,
			"cannot relay collaboration departure across instances",
			log.Error(err),
			log.String("document_id", config.DocumentID.String()),
		)
	}
}

func ServeRepoCollaboration(
	ctx context.Context,
	connection *websocket.Conn,
	lease *RoomLease,
	syncState *automerge.SyncState,
	config RepoConfig,
) error {
	conn, err := collabpkg.NewAdoptingServerConn(
		collabpkg.ServerConfig{ServerPeerID: config.ServerPeerID},
		syncState,
	)
	if err != nil {
		return fmt.Errorf("cannot create repo server connection: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	incoming := config.Incoming
	if incoming == nil {
		messages := make(chan Incoming, 1)

		incoming = messages

		go func() {
			_ = ReadMessages(ctx, connection, messages)
		}()
	}

	var join Incoming
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

	out, accepted, err := conn.Start(join.Data)
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

	lease.SetRemotePeerID(conn.RemotePeerID())
	lease.ReplayPresence()

	defer announceCollaborationDeparture(ctx, conn, lease, config)

	revision := lease.Revision()

	var tick <-chan time.Time

	if config.Refresher != nil {
		ticker := time.NewTicker(RepoRefreshInterval)
		defer ticker.Stop()

		tick = ticker.C
	}

	for {
		select {
		case <-config.Shutdown.Done():
			_ = connection.CloseNow()
			return nil
		case <-ctx.Done():
			return nil
		case message := <-incoming:
			if message.Err != nil {
				return repoReadResult(message.Err)
			}

			if message.MessageType != websocket.MessageBinary {
				continue
			}

			reply, fanout, err := conn.Receive(message.Data)
			if err != nil {
				return fmt.Errorf("cannot process repo frame: %w", err)
			}

			if err := writeRepoFrames(ctx, connection, reply); err != nil {
				return err
			}

			if fanout != nil {
				fanout = collabpkg.StampSelectionIdentity(
					fanout,
					config.DisplayName,
					config.IdentityID,
					config.AvatarURL,
				)
				if len(fanout) == 0 {
					continue
				}

				lease.BroadcastEphemeral(fanout)

				if config.PublishEphemeral != nil {
					if err := config.PublishEphemeral(ctx, fanout); err != nil && config.Logger != nil {
						config.Logger.WarnCtx(
							ctx,
							"cannot relay collaboration ephemeral across instances",
							log.Error(err),
							log.String("document_id", config.DocumentID.String()),
						)
					}
				}

				continue
			}

			kind, err := collabpkg.FrameKind(message.Data)
			if err != nil {
				return fmt.Errorf("cannot read repo frame type: %w", err)
			}

			frameType := collabpkg.MessageType(kind)
			if frameType != collabpkg.MessageSync && frameType != collabpkg.MessageRequest {
				continue
			}

			if config.RecordActors != nil {
				recordActorsFromFrame(
					ctx,
					message.Data,
					config,
				)
			}

			lease.NotifyPeers()

			if err := lease.PersistError(); err != nil {
				return fmt.Errorf("cannot persist repo collaboration: %w", err)
			}

			lease.SchedulePersist()
		case frame := <-lease.EphemeralCh:
			if err := WriteMessage(ctx, connection, websocket.MessageBinary, frame); err != nil {
				return err
			}
		case wake := <-lease.WakeCh:
			if err := lease.PersistError(); err != nil {
				return fmt.Errorf("cannot persist repo collaboration: %w", err)
			}

			if wake.Refresh && config.Refresher != nil {
				var changed bool

				revision, changed, err = config.Refresher.Refresh(
					ctx,
					config.Scope,
					config.DocumentID,
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

			frames, err := conn.SyncChanged()
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

			revision, changed, err = config.Refresher.Refresh(
				ctx,
				config.Scope,
				config.DocumentID,
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

			frames, err := conn.SyncChanged()
			if err != nil {
				return fmt.Errorf("cannot generate repo sync frames: %w", err)
			}

			if err := writeRepoFrames(ctx, connection, frames); err != nil {
				return err
			}
		}
	}
}

func repoReadResult(err error) error {
	switch websocket.CloseStatus(err) {
	case websocket.StatusNormalClosure,
		websocket.StatusGoingAway,
		websocket.StatusNoStatusRcvd:
		return nil
	default:
		return fmt.Errorf("repo collaboration read failed: %w", err)
	}
}

// A joined rollback is ignored when the operation itself was canceled.
func requestCanceled(err error) bool {
	for err != nil {
		type joined interface{ Unwrap() []error }
		if group, ok := err.(joined); ok {
			parts := group.Unwrap()
			if len(parts) == 0 {
				return false
			}

			return requestCanceled(parts[0])
		}

		if err == context.Canceled {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}

func writeRepoFrames(ctx context.Context, connection *websocket.Conn, frames [][]byte) error {
	for _, frame := range frames {
		if err := WriteMessage(ctx, connection, websocket.MessageBinary, frame); err != nil {
			return err
		}
	}

	return nil
}

func recordActorsFromFrame(
	ctx context.Context,
	frame []byte,
	config RepoConfig,
) {
	message, err := collabpkg.DecodeMessage(frame)
	if err != nil {
		if config.Logger != nil {
			config.Logger.WarnCtx(
				ctx,
				"cannot decode collaboration sync frame for actor attribution",
				log.Error(err),
				log.String("document_id", config.DocumentID.String()),
			)
		}

		return
	}

	actors, err := automerge.ActorsFromSyncMessage(message.Data)
	if err != nil {
		if config.Logger != nil {
			config.Logger.WarnCtx(
				ctx,
				"cannot read collaboration actors from sync message",
				log.Error(err),
				log.String("document_id", config.DocumentID.String()),
			)
		}

		return
	}

	if len(actors) == 0 {
		return
	}

	if err := config.RecordActors(ctx, actors); err != nil && config.Logger != nil {
		config.Logger.WarnCtx(
			ctx,
			"cannot record collaboration actors",
			log.Error(err),
			log.String("document_id", config.DocumentID.String()),
		)
	}
}
