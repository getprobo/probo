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

package slack_v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
)

type (
	SlackInteractivePayload struct {
		ResponseURL string `json:"response_url"`
		Actions     []struct {
			ActionID       string `json:"action_id"`
			Value          string `json:"value"`
			SelectedOption struct {
				Value string `json:"value"`
			} `json:"selected_option"`
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

const (
	StatusAccept = "accept"
	StatusReject = "reject"
)

func SlackHandler(slackSvc *slack.Service, slackSigningSecret string, logger *log.Logger, trustSvc *trust.Service) http.HandlerFunc {
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
		if action.Value == "" && action.SelectedOption.Value == "" {
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

		initialSlackMessage, err := slackSvc.GetInitialSlackMessageByChannelAndTS(ctx, slackPayload.Container.ChannelID, slackPayload.Container.MessageTS)
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

		if initialSlackMessage.RequesterEmail == nil {
			logger.ErrorCtx(ctx, "missing requester email", log.String("slack_message_id", initialSlackMessage.ID.String()))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}
		requesterEmail := *initialSlackMessage.RequesterEmail

		tenantSvc := trustSvc.WithTenant(initialSlackMessage.OrganizationID.TenantID())

		var documentIDs []gid.GID
		var reportIDs []gid.GID
		var fileIDs []gid.GID
		var statusAction string
		tenantSlackSvc := slackSvc.WithTenant(initialSlackMessage.OrganizationID.TenantID())

		// accept_all, reject_all
		if strings.HasSuffix(action.ActionID, "_all") {
			currentMessageId, err := gid.ParseGID(action.Value)
			if err != nil {
				httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid message ID"})
				return
			}

			documentIDs, reportIDs, fileIDs, err = tenantSlackSvc.SlackMessages.GetSlackMessageDocumentIDs(ctx, currentMessageId)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot load slack message document ids", log.Error(err))
				httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
				return
			}

			if strings.HasPrefix(action.ActionID, "accept_") {
				statusAction = StatusAccept
			} else {
				statusAction = StatusReject
			}
		} else {
			var gID gid.GID

			// handle_<document|report|file> is used in an overflow menu (a select) to choose between grant and reject on requested accesses
			if strings.HasPrefix(action.ActionID, "handle_") {
				// action value is the select option value. <accept|reject>-<ID>
				params := strings.Split(action.SelectedOption.Value, "/")

				if len(params) < 2 {
					httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid selected option format"})
					return
				}

				statusAction = params[0]
				gID, err = gid.ParseGID(params[1])
				if err != nil {
					httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid ID"})
					return
				}
			} else {
				// accept_<document|report|file>, reject_<document|report|file>, revoke_<document|report|file>
				gID, err = gid.ParseGID(action.Value)
				if err != nil {
					httpserver.RenderJSON(w, http.StatusBadRequest, SlackInteractiveResponse{Success: false, Message: "invalid ID"})
					return
				}

				if strings.HasPrefix(action.ActionID, "accept_") {
					statusAction = StatusAccept
				} else {
					statusAction = StatusReject
				}
			}

			switch gID.EntityType() {
			case coredata.DocumentEntityType:
				documentIDs = []gid.GID{gID}
			case coredata.ReportEntityType:
				reportIDs = []gid.GID{gID}
			case coredata.TrustCenterFileEntityType:
				fileIDs = []gid.GID{gID}
			default:
				logger.ErrorCtx(ctx, "unknown entity type", log.Error(err))
				httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
				return
			}
		}

		switch statusAction {
		case StatusAccept:
			if err := tenantSvc.TrustCenterAccesses.GrantByIDs(
				ctx,
				initialSlackMessage.OrganizationID,
				requesterEmail,
				documentIDs,
				reportIDs,
				fileIDs,
			); err != nil {
				logger.ErrorCtx(ctx, "cannot grant access", log.Error(err))
				httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
				return
			}
		case StatusReject:
			if err := tenantSvc.TrustCenterAccesses.RejectOrRevokeByIDs(
				ctx,
				initialSlackMessage.OrganizationID,
				requesterEmail,
				documentIDs,
				reportIDs,
				fileIDs,
			); err != nil {
				logger.ErrorCtx(ctx, "cannot reject access", log.Error(err))
				httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
				return
			}
		default:
			logger.ErrorCtx(ctx, "unknown status action", log.String("status_action", statusAction))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}

		if err := tenantSlackSvc.SlackMessages.UpdateSlackAccessMessage(
			ctx,
			initialSlackMessage.ID,
			slackPayload.ResponseURL,
			requesterEmail,
		); err != nil {
			logger.ErrorCtx(ctx, "cannot update Slack message", log.Error(err))
			httpserver.RenderJSON(w, http.StatusInternalServerError, SlackInteractiveResponse{Success: false, Message: "internal server error"})
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, SlackInteractiveResponse{Success: true, Message: "Access granted"})
	}
}
