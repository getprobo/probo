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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

// Document collaboration is served over the automerge-repo protocol; see
// document_collaboration_repo_handler.go. This file holds the pieces shared by
// that handler: the handler type, connection authorization, and the small
// WebSocket read/write helpers.
const (
	documentCollaborationMessageMaxBytes = 1024 * 1024
	documentCollaborationWriteTimeout    = 10 * time.Second
)

type (
	documentCollaborationHandler struct {
		logger         *log.Logger
		probo          *probo.Service
		iam            *iam.Service
		baseURL        *baseurl.BaseURL
		allowedOrigins []string
		hub            *documentCollaborationHub
		shutdown       context.Context
	}

	documentCollaborationIncoming struct {
		MessageType websocket.MessageType
		Data        []byte
		Err         error
	}
)

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
