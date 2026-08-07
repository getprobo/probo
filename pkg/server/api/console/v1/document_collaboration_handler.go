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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

const (
	documentCollaborationProtocol        = "automerge-sync-v1"
	documentCollaborationMessageMaxBytes = 1024 * 1024
	documentCollaborationRefreshInterval = 500 * time.Millisecond
	documentCollaborationWriteTimeout    = 10 * time.Second
)

type (
	documentCollaborationHandler struct {
		logger         *log.Logger
		probo          *probo.Service
		iam            *iam.Service
		baseURL        *baseurl.BaseURL
		allowedOrigins []string
	}

	documentCollaborationHandshake struct {
		Type         string `json:"type"`
		Version      int    `json:"version"`
		Revision     int64  `json:"revision"`
		NeedsSeed    bool   `json:"needsSeed"`
		SeedContent  string `json:"seedContent,omitempty"`
		ConnectionID string `json:"connectionId"`
	}

	documentCollaborationIncoming struct {
		MessageType websocket.MessageType
		Data        []byte
		Err         error
	}

	documentCollaborationPresenceInput struct {
		Type           string `json:"type"`
		AnchorPosition int    `json:"anchorPosition"`
		HeadPosition   int    `json:"headPosition"`
	}

	documentCollaborationPresence struct {
		ConnectionID   string `json:"connectionId"`
		IdentityID     string `json:"identityId"`
		AnchorPosition int    `json:"anchorPosition"`
		HeadPosition   int    `json:"headPosition"`
	}

	documentCollaborationPresenceSnapshot struct {
		Type      string                          `json:"type"`
		Presences []documentCollaborationPresence `json:"presences"`
	}
)

func (h *documentCollaborationHandler) handle(w http.ResponseWriter, r *http.Request) {
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
	identity := authn.IdentityFromContext(r.Context())
	connectionID, err := newDocumentCollaborationConnectionID()
	if err != nil {
		jsonx.RenderInternalServerError(w)
		return
	}

	collaboration, err := h.probo.Documents.OpenCollaboration(
		r.Context(),
		scope,
		documentVersionID,
	)
	if err != nil {
		h.renderServiceError(w, r, documentVersionIDString, err)
		return
	}
	defer func() { _ = collaboration.Document.Close(context.Background()) }()
	defer func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.probo.Documents.DeleteCollaborationPresence(
			deleteCtx,
			scope,
			connectionID,
		); err != nil {
			h.logger.WarnCtx(
				deleteCtx,
				"cannot delete document collaboration presence",
				log.Error(err),
				log.String("document_version_id", documentVersionIDString),
			)
		}
	}()
	if collaboration.NeedsSeed {
		defer func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.probo.Documents.ReleaseCollaborationSeed(
				releaseCtx,
				scope,
				documentVersionID,
			); err != nil {
				h.logger.WarnCtx(
					releaseCtx,
					"cannot release document collaboration seed",
					log.Error(err),
					log.String("document_version_id", documentVersionIDString),
				)
			}
		}()
	}

	connection, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{
			Subprotocols:       []string{documentCollaborationProtocol},
			OriginPatterns:     h.allowedOrigins,
			CompressionMode:    websocket.CompressionDisabled,
			InsecureSkipVerify: false,
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			r.Context(),
			"cannot accept document collaboration connection",
			log.Error(err),
			log.String("document_version_id", documentVersionIDString),
		)
		return
	}
	defer func() {
		_ = connection.Close(websocket.StatusNormalClosure, "")
	}()
	connection.SetReadLimit(documentCollaborationMessageMaxBytes)

	if connection.Subprotocol() != documentCollaborationProtocol {
		_ = connection.Close(websocket.StatusPolicyViolation, "required subprotocol not negotiated")
		return
	}

	syncState, err := collaboration.Document.NewSyncState(r.Context())
	if err != nil {
		h.closeWithError(r.Context(), connection, documentVersionIDString, err)
		return
	}
	defer func() { _ = syncState.Close(context.Background()) }()

	seedContent := ""
	if collaboration.NeedsSeed {
		seedContent = collaboration.SeedContent
	}
	handshake, err := json.Marshal(
		documentCollaborationHandshake{
			Type:         "ready",
			Version:      1,
			Revision:     collaboration.Revision,
			NeedsSeed:    collaboration.NeedsSeed,
			SeedContent:  seedContent,
			ConnectionID: connectionID,
		},
	)
	if err != nil {
		h.closeWithError(r.Context(), connection, documentVersionIDString, err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := writeCollaborationMessage(
		ctx,
		connection,
		websocket.MessageText,
		handshake,
	); err != nil {
		return
	}
	if err := sendAvailableSyncMessages(ctx, connection, syncState); err != nil {
		h.closeWithError(ctx, connection, documentVersionIDString, err)
		return
	}

	incoming := make(chan documentCollaborationIncoming, 1)
	go readCollaborationMessages(ctx, connection, incoming)

	revision := collaboration.Revision
	ticker := time.NewTicker(documentCollaborationRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-incoming:
			if message.Err != nil {
				if websocket.CloseStatus(message.Err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(message.Err) == websocket.StatusGoingAway {
					return
				}
				h.logger.WarnCtx(
					ctx,
					"document collaboration connection closed with read error",
					log.Error(message.Err),
					log.String("document_version_id", documentVersionIDString),
				)
				return
			}
			if message.MessageType != websocket.MessageBinary {
				if message.MessageType != websocket.MessageText {
					_ = connection.Close(websocket.StatusUnsupportedData, "unsupported message type")
					return
				}

				var presence documentCollaborationPresenceInput
				if err := json.Unmarshal(message.Data, &presence); err != nil ||
					presence.Type != "presence" {
					_ = connection.Close(websocket.StatusInvalidFramePayloadData, "invalid presence message")
					return
				}
				if err := h.probo.Documents.SaveCollaborationPresence(
					ctx,
					scope,
					documentVersionID,
					identity.ID,
					connectionID,
					presence.AnchorPosition,
					presence.HeadPosition,
				); err != nil {
					h.closeWithError(ctx, connection, documentVersionIDString, err)
					return
				}
				continue
			}
			if err := syncState.ReceiveMessage(ctx, message.Data); err != nil {
				h.closeWithError(ctx, connection, documentVersionIDString, err)
				return
			}
			revision, err = h.probo.Documents.PersistCollaboration(
				ctx,
				scope,
				documentVersionID,
				collaboration.Document,
			)
			if err != nil {
				h.closeWithError(ctx, connection, documentVersionIDString, err)
				return
			}
			if err := sendAvailableSyncMessages(ctx, connection, syncState); err != nil {
				h.closeWithError(ctx, connection, documentVersionIDString, err)
				return
			}
		case <-ticker.C:
			var changed bool
			revision, changed, err = h.probo.Documents.RefreshCollaboration(
				ctx,
				scope,
				documentVersionID,
				collaboration.Document,
				revision,
			)
			if err != nil {
				h.closeWithError(ctx, connection, documentVersionIDString, err)
				return
			}
			if changed {
				if err := sendAvailableSyncMessages(ctx, connection, syncState); err != nil {
					h.closeWithError(ctx, connection, documentVersionIDString, err)
					return
				}
			}
			if err := h.sendPresences(
				ctx,
				connection,
				scope,
				documentVersionID,
				connectionID,
			); err != nil {
				h.closeWithError(ctx, connection, documentVersionIDString, err)
				return
			}
		}
	}
}

func (h *documentCollaborationHandler) sendPresences(
	ctx context.Context,
	connection *websocket.Conn,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	connectionID string,
) error {
	stored, err := h.probo.Documents.ListCollaborationPresences(
		ctx,
		scope,
		documentVersionID,
		connectionID,
	)
	if err != nil {
		return fmt.Errorf("cannot list document collaboration presences: %w", err)
	}

	presences := make([]documentCollaborationPresence, len(stored))
	for i, presence := range stored {
		presences[i] = documentCollaborationPresence{
			ConnectionID:   presence.ConnectionID,
			IdentityID:     presence.IdentityID.String(),
			AnchorPosition: presence.AnchorPosition,
			HeadPosition:   presence.HeadPosition,
		}
	}
	data, err := json.Marshal(
		documentCollaborationPresenceSnapshot{
			Type:      "presence",
			Presences: presences,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal document collaboration presences: %w", err)
	}
	if err := writeCollaborationMessage(ctx, connection, websocket.MessageText, data); err != nil {
		return fmt.Errorf("cannot send document collaboration presences: %w", err)
	}

	return nil
}

func newDocumentCollaborationConnectionID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("cannot generate document collaboration connection ID: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func (h *documentCollaborationHandler) authorize(
	ctx context.Context,
	documentVersionID gid.GID,
) (*coredata.Scope, error) {
	versionScope, err := h.authorizeResource(
		ctx,
		documentVersionID,
		probo.ActionDocumentVersionGet,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot authorize document version: %w", err)
	}

	version, err := h.probo.Documents.GetVersion(ctx, versionScope, documentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot load authorized document version: %w", err)
	}

	scope, err := h.authorizeResource(ctx, version.DocumentID, probo.ActionDocumentUpdate)
	if err != nil {
		return nil, fmt.Errorf("cannot authorize document update: %w", err)
	}

	return scope, nil
}

func (h *documentCollaborationHandler) authorizeResource(
	ctx context.Context,
	resourceID gid.GID,
	action string,
) (*coredata.Scope, error) {
	identity := authn.IdentityFromContext(ctx)
	session := authn.SessionFromContext(ctx)
	params := iam.AuthorizeParams{
		Principal:          identity.ID,
		Resource:           resourceID,
		Action:             action,
		ResourceAttributes: make(map[string]string),
	}
	if session != nil {
		params.Session = &session.ID
	}

	scope, err := h.iam.Authorizer.Authorize(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cannot authorize resource: %w", err)
	}

	return scope, nil
}

func (h *documentCollaborationHandler) renderAuthorizationError(w http.ResponseWriter, err error) {
	if scopeErr, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		bearertoken.SetBearerInsufficientScope(w, h.baseURL, scopeErr.Scopes...)
		jsonx.RenderForbidden(w)
		return
	}
	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		jsonx.RenderForbidden(w)
		return
	}

	jsonx.RenderNotFound(w, fmt.Errorf("document version not found"))
}

func (h *documentCollaborationHandler) renderServiceError(
	w http.ResponseWriter,
	r *http.Request,
	documentVersionID string,
	err error,
) {
	if errors.Is(err, coredata.ErrResourceNotFound) {
		jsonx.RenderNotFound(w, fmt.Errorf("document version not found"))
		return
	}
	if _, ok := errors.AsType[*probo.ErrDocumentVersionNotDraft](err); ok {
		httpserver.RenderError(w, http.StatusConflict, err)
		return
	}
	if _, ok := errors.AsType[*probo.ErrDocumentArchived](err); ok {
		httpserver.RenderError(w, http.StatusConflict, err)
		return
	}
	if _, ok := errors.AsType[*probo.ErrDocumentVersionGenerated](err); ok {
		httpserver.RenderError(w, http.StatusConflict, err)
		return
	}

	h.logger.ErrorCtx(
		r.Context(),
		"cannot open document collaboration",
		log.Error(err),
		log.String("document_version_id", documentVersionID),
	)
	jsonx.RenderInternalServerError(w)
}

func (h *documentCollaborationHandler) closeWithError(
	ctx context.Context,
	connection *websocket.Conn,
	documentVersionID string,
	err error,
) {
	h.logger.ErrorCtx(
		ctx,
		"document collaboration failed",
		log.Error(err),
		log.String("document_version_id", documentVersionID),
	)
	_ = connection.Close(websocket.StatusInternalError, "collaboration failed")
}

func readCollaborationMessages(
	ctx context.Context,
	connection *websocket.Conn,
	incoming chan<- documentCollaborationIncoming,
) {
	for {
		messageType, data, err := connection.Read(ctx)
		message := documentCollaborationIncoming{
			MessageType: messageType,
			Data:        data,
			Err:         err,
		}
		select {
		case incoming <- message:
		case <-ctx.Done():
			return
		}
		if err != nil {
			return
		}
	}
}

func sendAvailableSyncMessages(
	ctx context.Context,
	connection *websocket.Conn,
	syncState *automerge.SyncState,
) error {
	for range 100 {
		message, ok, err := syncState.GenerateMessage(ctx)
		if err != nil {
			return fmt.Errorf("cannot generate sync message: %w", err)
		}
		if !ok {
			return nil
		}
		if err := writeCollaborationMessage(
			ctx,
			connection,
			websocket.MessageBinary,
			message,
		); err != nil {
			return fmt.Errorf("cannot write sync message: %w", err)
		}
	}

	return fmt.Errorf("cannot generate sync messages: protocol did not quiesce")
}

func writeCollaborationMessage(
	ctx context.Context,
	connection *websocket.Conn,
	messageType websocket.MessageType,
	data []byte,
) error {
	writeCtx, cancel := context.WithTimeout(ctx, documentCollaborationWriteTimeout)
	defer cancel()

	if err := connection.Write(writeCtx, messageType, data); err != nil {
		return fmt.Errorf("cannot write collaboration message: %w", err)
	}

	return nil
}
