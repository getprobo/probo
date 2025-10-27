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
	"errors"
	"fmt"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/slack"
	"go.gearno.de/kit/pg"
)

const (
	slackMessageDeduplicationWindow = 7 * 24 * time.Hour
)

type (
	SlackMessageService struct {
		svc         *TenantService
		slackClient *slack.Client
	}

	SlackMessageDocument struct {
		ID      string
		Title   string
		Granted bool
	}

	SlackMessageReport struct {
		ID      string
		Title   string
		AuditID string
		Granted bool
	}

	SlackMessageFile struct {
		ID       string
		Name     string
		Category string
		Granted  bool
	}

	SlackMessageMetadata struct {
		Documents []SlackMessageDocument
		Reports   []SlackMessageReport
		Files     []SlackMessageFile
	}
)

func (m SlackMessageMetadata) toMap() map[string]any {
	return map[string]any{
		"documents": m.Documents,
		"reports":   m.Reports,
		"files":     m.Files,
	}
}

func (s *Service) GetInitialSlackMessageByChannelAndTS(
	ctx context.Context,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	var slackMessage coredata.SlackMessage

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := slackMessage.LoadInitialByChannelAndTS(ctx, conn, coredata.NewNoScope(), channelID, messageTS); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &slackMessage, nil
}

func (s *SlackMessageService) GetSlackMessageDocumentIDs(
	ctx context.Context,
	slackMessageID gid.GID,
) (documentIDs []gid.GID, reportIDs []gid.GID, fileIDs []gid.GID, err error) {
	var slackMessage coredata.SlackMessage

	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := slackMessage.LoadById(ctx, conn, s.svc.scope, slackMessageID); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, err
	}

	documentIDs = extractIDsFromMetadata(slackMessage.Metadata, "documents")
	reportIDs = extractIDsFromMetadata(slackMessage.Metadata, "reports")
	fileIDs = extractIDsFromMetadata(slackMessage.Metadata, "files")

	return documentIDs, reportIDs, fileIDs, nil
}

func (s *SlackMessageService) UpdateSlackAccessMessage(
	ctx context.Context,
	slackMessageID gid.GID,
	responseURL string,
	requesterEmail string,
) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		var slackMessage coredata.SlackMessage
		if err := slackMessage.LoadById(ctx, tx, s.svc.scope, slackMessageID); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		var trustCenter coredata.TrustCenter
		if err := trustCenter.LoadByOrganizationID(ctx, tx, s.svc.scope, slackMessage.OrganizationID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		var trustCenterAccess coredata.TrustCenterAccess
		if err := trustCenterAccess.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, trustCenter.ID, requesterEmail); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		documents, reports, files, err := s.loadDocumentsReportsAndFilesFromAccesses(ctx, tx, trustCenterAccess.ID)
		if err != nil {
			return err
		}

		newSlackMessageID := gid.New(s.svc.scope.GetTenantID(), coredata.SlackMessageEntityType)

		updatedBody, err := s.buildAccessRequestMessage(
			newSlackMessageID,
			trustCenterAccess.Name,
			requesterEmail,
			trustCenter.OrganizationID,
			documents,
			reports,
			files,
		)
		if err != nil {
			return err
		}

		metadata := SlackMessageMetadata{
			Documents: documents,
			Reports:   reports,
			Files:     files,
		}

		now := time.Now()
		newSlackMessage := &coredata.SlackMessage{
			ID:                    newSlackMessageID,
			OrganizationID:        slackMessage.OrganizationID,
			Type:                  slackMessage.Type,
			Body:                  updatedBody,
			MessageTS:             slackMessage.MessageTS,
			ChannelID:             slackMessage.ChannelID,
			RequesterEmail:        slackMessage.RequesterEmail,
			Metadata:              metadata.toMap(),
			InitialSlackMessageID: slackMessage.InitialSlackMessageID,
			CreatedAt:             now,
			UpdatedAt:             now,
			SentAt:                &now,
		}

		if err := newSlackMessage.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert slack message: %w", err)
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

		var connectors coredata.Connectors
		if err := connectors.LoadAllByOrganizationIDWithoutDecryptedConnection(
			ctx,
			tx,
			s.svc.scope,
			trustCenter.OrganizationID,
		); err != nil {
			return fmt.Errorf("cannot load connectors: %w", err)
		}

		hasSlackConnector := false
		for _, connector := range connectors {
			if connector.Provider == coredata.ConnectorProviderSlack {
				hasSlackConnector = true
				break
			}
		}

		if !hasSlackConnector {
			return fmt.Errorf("no slack connector found for organization")
		}

		documents, reports, files, err := s.loadDocumentsReportsAndFilesFromAccesses(ctx, tx, trustCenterAccess.ID)
		if err != nil {
			return fmt.Errorf("cannot load documents, reports and files: %w", err)
		}

		slackMessageID := gid.New(s.svc.scope.GetTenantID(), coredata.SlackMessageEntityType)

		body, err := s.buildAccessRequestMessage(
			slackMessageID,
			trustCenterAccess.Name,
			requesterEmail,
			trustCenter.OrganizationID,
			documents,
			reports,
			files,
		)
		if err != nil {
			return fmt.Errorf("cannot build access request message: %w", err)
		}

		metadata := SlackMessageMetadata{
			Documents: documents,
			Reports:   reports,
			Files:     files,
		}

		now := time.Now()
		slackMessage := &coredata.SlackMessage{
			ID:             slackMessageID,
			OrganizationID: trustCenter.OrganizationID,
			Type:           coredata.SlackMessageTypeTrustCenterAccessRequest,
			Body:           body,
			RequesterEmail: &requesterEmail,
			Metadata:       metadata.toMap(),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		sevenDaysAgo := now.Add(-slackMessageDeduplicationWindow)

		var existingMessage coredata.SlackMessage
		err = existingMessage.LoadLatestByRequesterEmailAndType(
			ctx,
			tx,
			s.svc.scope,
			trustCenter.OrganizationID,
			requesterEmail,
			coredata.SlackMessageTypeTrustCenterAccessRequest,
			sevenDaysAgo,
		)
		if err == nil {
			slackMessage.MessageTS = existingMessage.MessageTS
			slackMessage.ChannelID = existingMessage.ChannelID
			slackMessage.InitialSlackMessageID = existingMessage.InitialSlackMessageID

			if err := slackMessage.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert slack message: %w", err)
			}

			return nil
		}
		var notFoundErr coredata.ErrSlackMessageNotFound
		if !errors.Is(err, notFoundErr) {
			return fmt.Errorf("cannot load existing slack message: %w", err)
		}

		slackMessage.InitialSlackMessageID = slackMessageID
		if err := slackMessage.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert slack message: %w", err)
		}

		return nil
	})
}

func (s *SlackMessageService) loadDocumentsReportsAndFilesFromAccesses(
	ctx context.Context,
	conn pg.Conn,
	trustCenterAccessID gid.GID,
) (
	documents []SlackMessageDocument,
	reports []SlackMessageReport,
	files []SlackMessageFile,
	err error,
) {
	documents = []SlackMessageDocument{}
	reports = []SlackMessageReport{}
	files = []SlackMessageFile{}

	var accesses coredata.TrustCenterDocumentAccesses
	if err := accesses.LoadAllByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID); err != nil {
		return nil, nil, nil, fmt.Errorf("cannot load trust center document accesses: %w", err)
	}

	for _, access := range accesses {
		if access.DocumentID != nil {
			doc := &coredata.Document{}
			if err := doc.LoadByID(ctx, conn, s.svc.scope, *access.DocumentID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load document: %w", err)
			}
			documents = append(documents, SlackMessageDocument{
				ID:      access.DocumentID.String(),
				Title:   doc.Title,
				Granted: access.Active,
			})
		}

		if access.ReportID != nil {
			rep := &coredata.Report{}
			if err := rep.LoadByID(ctx, conn, s.svc.scope, *access.ReportID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load report: %w", err)
			}

			audit := &coredata.Audit{}
			if err := audit.LoadByReportID(ctx, conn, s.svc.scope, *access.ReportID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load audit: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, audit.FrameworkID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load framework: %w", err)
			}

			label := framework.Name
			if audit.Name != nil && *audit.Name != "" {
				label = label + " - " + *audit.Name
			}
			reports = append(reports, SlackMessageReport{
				ID:      access.ReportID.String(),
				Title:   label,
				AuditID: audit.ID.String(),
				Granted: access.Active,
			})
		}

		if access.TrustCenterFileID != nil {
			file := &coredata.TrustCenterFile{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, *access.TrustCenterFileID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load trust center file: %w", err)
			}
			files = append(files, SlackMessageFile{
				ID:       access.TrustCenterFileID.String(),
				Name:     file.Name,
				Category: file.Category,
				Granted:  access.Active,
			})
		}
	}

	return documents, reports, files, nil
}

func (s *SlackMessageService) buildAccessRequestMessage(
	slackMessageID gid.GID,
	requesterName string,
	requesterEmail string,
	organizationID gid.GID,
	documents []SlackMessageDocument,
	reports []SlackMessageReport,
	files []SlackMessageFile,
) (map[string]any, error) {
	var documentIDs []string
	var reportIDs []string
	var fileIDs []string

	for _, doc := range documents {
		documentIDs = append(documentIDs, doc.ID)
	}
	for _, rep := range reports {
		reportIDs = append(reportIDs, rep.ID)
	}
	for _, file := range files {
		fileIDs = append(fileIDs, file.ID)
	}

	templateData := struct {
		RequesterName  string
		RequesterEmail string
		OrganizationID string
		Domain         string
		SlackMessageID string
		DocumentIDs    []string
		ReportIDs      []string
		FileIDs        []string
		Documents      []SlackMessageDocument
		Reports        []SlackMessageReport
		Files          []SlackMessageFile
	}{
		RequesterName:  requesterName,
		RequesterEmail: requesterEmail,
		OrganizationID: organizationID.String(),
		Domain:         s.svc.hostname,
		SlackMessageID: slackMessageID.String(),
		DocumentIDs:    documentIDs,
		ReportIDs:      reportIDs,
		FileIDs:        fileIDs,
		Documents:      documents,
		Reports:        reports,
		Files:          files,
	}

	var buf bytes.Buffer
	if err := accessRequestTemplate.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	var body map[string]any
	if err := json.NewDecoder(&buf).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	return body, nil
}

func extractIDsFromMetadata(metadata map[string]any, fieldName string) []gid.GID {
	ids := []gid.GID{}

	items, ok := metadata[fieldName].([]any)
	if !ok || items == nil {
		return ids
	}

	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		if !ok {
			continue
		}
		idStr, ok := item["ID"].(string)
		if !ok {
			continue
		}
		id, err := gid.ParseGID(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids
}
