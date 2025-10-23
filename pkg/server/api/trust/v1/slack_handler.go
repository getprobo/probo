// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package trust_v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/slack"
	"github.com/getprobo/probo/pkg/trust"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

type (
	SlackInteractivePayload struct {
		ResponseURL string `json:"response_url"`
		Actions     []struct {
			ActionID string `json:"action_id"`
			Value    string `json:"value"`
		} `json:"actions"`
		Container struct {
			MessageTS string `json:"message_ts"`
			ChannelID string `json:"channel_id"`
		} `json:"container"`
	}

	SlackInteractiveResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
	}
)

func slackHandler(trustSvc *trust.Service, slackSigningSecret string, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "cannot read request body"})
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		timestamp := r.Header.Get("X-Slack-Request-Timestamp")
		signature := r.Header.Get("X-Slack-Signature")
		if timestamp == "" || signature == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "missing Slack signature headers"})
			return
		}

		if err := slack.VerifySignature(slackSigningSecret, timestamp, signature, bodyBytes); err != nil {
			logger.ErrorCtx(ctx, "invalid Slack signature", log.Error(err))
			httpserver.RenderJSON(w, http.StatusUnauthorized, SlackInteractiveResponse{Success: false, Message: "invalid Slack signature"})
			return
		}

		var slackPayload SlackInteractivePayload
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "unsupported content type"})
			return
		}

		if err := r.ParseForm(); err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "cannot parse form"})
			return
		}

		raw := r.FormValue("payload")
		if raw == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "empty payload field"})
			return
		}

		if err := json.NewDecoder(strings.NewReader(raw)).Decode(&slackPayload); err != nil {
			logger.ErrorCtx(ctx, "cannot parse Slack payload", log.Error(err))
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "cannot parse Slack payload"})
			return
		}

		// Slack sends empty action for url button clicks
		if len(slackPayload.Actions) == 0 {
			httpserver.RenderJSON(w, http.StatusOK, SlackInteractiveResponse{Success: true, Message: "no action required"})
			return
		}
		action := slackPayload.Actions[0]
		if action.Value == "" {
			httpserver.RenderJSON(w, http.StatusOK, SlackInteractiveResponse{Success: true, Message: "no action required"})
			return
		}

		if slackPayload.Container.MessageTS == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "missing message_ts"})
			return
		}

		if slackPayload.Container.ChannelID == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "missing channel_id"})
			return
		}

		if slackPayload.ResponseURL == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "missing response_url"})
			return
		}

		initialSlackMessage, err := trustSvc.GetInitialSlackMessageByChannelAndTS(ctx, slackPayload.Container.ChannelID, slackPayload.Container.MessageTS)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot load slack message", log.Error(err))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}

		//TODO: Update the message when it is too old to be updated
		fourteenDaysAgo := time.Now().Add(-14 * 24 * time.Hour)
		if initialSlackMessage.CreatedAt.Before(fourteenDaysAgo) {
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "this message is too old to be updated (older than 14 days)"})
			return
		}

		if initialSlackMessage.RequesterEmail == nil || *initialSlackMessage.RequesterEmail == "" {
			logger.ErrorCtx(ctx, "missing requester email", log.String("slack_message_id", initialSlackMessage.ID.String()))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}
		requesterEmail := *initialSlackMessage.RequesterEmail

		tenantSvc := trustSvc.WithTenant(initialSlackMessage.OrganizationID.TenantID())

		var documentIDs []gid.GID
		var reportIDs []gid.GID

		switch action.ActionID {
		case "accept_all":
			currentMessageId, err := gid.ParseGID(action.Value)
			if err != nil {
				httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid message ID"})
				return
			}

			documentIDs, reportIDs, err = tenantSvc.SlackMessages.GetSlackMessageMetadataByID(ctx, currentMessageId)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot load slack message metadata by ID", log.Error(err))
				httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
				return
			}

		case "accept_document":
			docID, err := gid.ParseGID(action.Value)
			if err != nil {
				httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid document ID"})
				return
			}
			documentIDs = []gid.GID{docID}

		case "accept_report":
			repID, err := gid.ParseGID(action.Value)
			if err != nil {
				httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid report ID"})
				return
			}
			reportIDs = []gid.GID{repID}

		default:
			httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: fmt.Sprintf("unknown action: %s", action.ActionID)})
			return
		}

		if err := tenantSvc.TrustCenterAccesses.AcceptByIDs(
			ctx,
			initialSlackMessage.OrganizationID,
			requesterEmail,
			documentIDs,
			reportIDs,
		); err != nil {
			logger.ErrorCtx(ctx, "failed to grant access", log.Error(err))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}

		if err := tenantSvc.SlackMessages.UpdateSlackAccessMessage(
			ctx,
			initialSlackMessage.ID,
			slackPayload.ResponseURL,
			requesterEmail,
		); err != nil {
			logger.ErrorCtx(ctx, "failed to update Slack message", log.Error(err))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, SlackInteractiveResponse{Success: true, Message: "Access granted"})
	}
}
