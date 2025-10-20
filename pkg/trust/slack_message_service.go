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

package trust

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/slack"
	"go.gearno.de/kit/pg"
)

const (
	slackMessageDeduplicationWindow = 7 * 24 * time.Hour
	trustCenterAccessURLFormat      = "https://%s/organizations/%s/trust-center/access"
)

type SlackMessageService struct {
	svc         *TenantService
	slackClient *slack.Client
}

func (s *SlackMessageService) LoadSlackMessageUnscoped(
	ctx context.Context,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	var slackMessage coredata.SlackMessage

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := slackMessage.LoadByChannelAndTSUnscoped(ctx, conn, channelID, messageTS); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &slackMessage, nil
}

func (s *SlackMessageService) UpdateSlackAccessMessage(
	ctx context.Context,
	slackMessageID gid.GID,
	actionID string,
	value string,
	responseURL string,
) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		var slackMessage coredata.SlackMessage
		if err := slackMessage.LoadById(ctx, tx, s.svc.scope, slackMessageID); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		baseBody := slackMessage.Body
		var latestUpdate coredata.SlackMessageUpdate
		if err := latestUpdate.LoadLatestBySlackMessageID(ctx, tx, slackMessage.ID); err == nil {
			baseBody = latestUpdate.Body
		}

		accessTabURL := fmt.Sprintf(trustCenterAccessURLFormat, s.svc.hostname, slackMessage.OrganizationID)
		updatedBody := s.changeButton(baseBody, actionID, value, accessTabURL)

		slackMessageUpdate := coredata.NewSlackMessageUpdate(s.svc.scope, slackMessage.ID, updatedBody)
		now := time.Now()
		slackMessageUpdate.SentAt = &now
		if err := slackMessageUpdate.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert slack message update: %w", err)
		}

		if err := s.slackClient.UpdateInteractiveMessage(ctx, responseURL, updatedBody); err != nil {
			return fmt.Errorf("failed to update Slack message: %w", err)
		}

		return nil
	})
}

func (s *SlackMessageService) QueueSlackNotification(
	ctx context.Context,
	requesterEmail string,
	trustCenterID gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		var trustCenterAccess coredata.TrustCenterAccess
		if err := trustCenterAccess.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, trustCenterID, requesterEmail); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		var trustCenter coredata.TrustCenter
		if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, trustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		var accesses coredata.TrustCenterDocumentAccesses
		if err := accesses.LoadAllByTrustCenterAccessID(ctx, tx, s.svc.scope, trustCenterAccess.ID); err != nil {
			return fmt.Errorf("cannot load trust center document accesses: %w", err)
		}

		var documentIDs []string
		var reportIDs []string
		var documents []struct {
			ID      string
			Title   string
			Granted bool
		}
		var reports []struct {
			ID      string
			Title   string
			AuditID string
			Granted bool
		}

		for _, access := range accesses {
			if access.DocumentID != nil {
				doc := &coredata.Document{}
				if err := doc.LoadByID(ctx, tx, s.svc.scope, *access.DocumentID); err != nil {
					return fmt.Errorf("cannot load document: %w", err)
				}
				documentIDs = append(documentIDs, access.DocumentID.String())
				documents = append(documents, struct {
					ID      string
					Title   string
					Granted bool
				}{
					ID:      access.DocumentID.String(),
					Title:   doc.Title,
					Granted: access.Active,
				})
			}

			if access.ReportID != nil {
				rep := &coredata.Report{}
				if err := rep.LoadByID(ctx, tx, s.svc.scope, *access.ReportID); err != nil {
					return fmt.Errorf("cannot load report: %w", err)
				}

				audit := &coredata.Audit{}
				if err := audit.LoadByReportID(ctx, tx, s.svc.scope, *access.ReportID); err != nil {
					return fmt.Errorf("cannot load audit: %w", err)
				}

				framework := &coredata.Framework{}
				if err := framework.LoadByID(ctx, tx, s.svc.scope, audit.FrameworkID); err != nil {
					return fmt.Errorf("cannot load framework: %w", err)
				}

				label := framework.Name
				if audit.Name != nil && *audit.Name != "" {
					label = label + " - " + *audit.Name
				}
				reportIDs = append(reportIDs, access.ReportID.String())
				reports = append(reports, struct {
					ID      string
					Title   string
					AuditID string
					Granted bool
				}{
					ID:      access.ReportID.String(),
					Title:   label,
					AuditID: audit.ID.String(),
					Granted: access.Active,
				})
			}
		}

		templateData := struct {
			RequesterName  string
			RequesterEmail string
			OrganizationID string
			Domain         string
			DocumentIDs    []string
			ReportIDs      []string
			Documents      []struct {
				ID      string
				Title   string
				Granted bool
			}
			Reports []struct {
				ID      string
				Title   string
				AuditID string
				Granted bool
			}
		}{
			RequesterName:  trustCenterAccess.Name,
			RequesterEmail: requesterEmail,
			OrganizationID: trustCenter.OrganizationID.String(),
			Domain:         s.svc.hostname,
			DocumentIDs:    documentIDs,
			ReportIDs:      reportIDs,
			Documents:      documents,
			Reports:        reports,
		}

		var buf bytes.Buffer
		if err := accessRequestTemplate.Execute(&buf, templateData); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		var body map[string]any
		if err := json.NewDecoder(&buf).Decode(&body); err != nil {
			return fmt.Errorf("failed to parse template JSON: %w", err)
		}

		sevenDaysAgo := time.Now().Add(-slackMessageDeduplicationWindow)
		var existingMessage coredata.SlackMessage
		err := existingMessage.LoadLatestByRequesterEmailAndType(
			ctx,
			tx,
			s.svc.scope,
			trustCenter.OrganizationID,
			requesterEmail,
			coredata.SlackMessageTypeTrustCenterAccessRequest,
			sevenDaysAgo,
		)

		if err == nil && existingMessage.MessageTS != nil && existingMessage.ChannelID != nil {
			slackMessageUpdate := coredata.NewSlackMessageUpdate(s.svc.scope, existingMessage.ID, body)
			if err := slackMessageUpdate.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert slack message update: %w", err)
			}

			return nil
		}

		slackMessage := coredata.NewSlackMessage(s.svc.scope, trustCenter.OrganizationID, coredata.SlackMessageTypeTrustCenterAccessRequest, body, &requesterEmail)
		if err := slackMessage.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert slack message: %w", err)
		}

		return nil
	})
}

func (s *SlackMessageService) changeButton(body map[string]any, actionID string, value string, accessTabURL string) map[string]any {
	blocks, ok := body["blocks"].([]any)
	if !ok {
		return body
	}

	isAcceptAll := actionID == "accept_all"

	updatedBlocks := make([]any, len(blocks))
	for i, blockAny := range blocks {
		block, ok := blockAny.(map[string]any)
		if !ok {
			updatedBlocks[i] = blockAny
			continue
		}

		blockCopy := make(map[string]any)
		maps.Copy(blockCopy, block)

		if blockType, ok := block["type"].(string); ok && blockType == "section" {
			if acc, ok := block["accessory"].(map[string]any); ok {
				if s.shouldChangeButton(acc, actionID, value, isAcceptAll) {
					blockCopy["accessory"] = s.makeStaticButton(accessTabURL)
				}
			}
		}

		if blockType, ok := block["type"].(string); ok && blockType == "actions" {
			if elements, ok := block["elements"].([]any); ok {
				updatedElements := make([]any, len(elements))
				for j, elemAny := range elements {
					elem, ok := elemAny.(map[string]any)
					if !ok {
						updatedElements[j] = elemAny
						continue
					}

					if s.shouldChangeButton(elem, actionID, value, isAcceptAll) {
						updatedElements[j] = s.makeStaticButton(accessTabURL)
					} else {
						updatedElements[j] = elem
					}
				}
				blockCopy["elements"] = updatedElements
			}
		}

		updatedBlocks[i] = blockCopy
	}

	updatedBody := make(map[string]any)
	maps.Copy(updatedBody, body)
	updatedBody["blocks"] = updatedBlocks

	return updatedBody
}

func (s *SlackMessageService) shouldChangeButton(button map[string]any, actionID string, value string, isAcceptAll bool) bool {
	if button["type"] != "button" {
		return false
	}

	btnActionID, _ := button["action_id"].(string)
	btnValue, _ := button["value"].(string)

	isExactMatch := btnActionID == actionID && btnValue == value
	isAcceptAllMatch := isAcceptAll && (btnActionID == "accept_document" || btnActionID == "accept_report")

	return isExactMatch || isAcceptAllMatch
}

func (s *SlackMessageService) makeStaticButton(accessTabURL string) map[string]any {
	return map[string]any{
		"type": "button",
		"text": map[string]any{
			"type": "plain_text",
			"text": "✓ Granted",
		},
		"url": accessTabURL,
	}
}
