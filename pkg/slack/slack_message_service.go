// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

const (
	slackMessageDeduplicationWindow = 7 * 24 * time.Hour
)

var (
	ErrNoSlackConnector                  = errors.New("no slack connector found for organization")
	ErrCompliancePortalMetadataAmbiguous = errors.New("compliance portal id is missing from slack message metadata and organization has multiple portals")
)

type (
	SlackMessageDocument struct {
		ID     string
		Title  string
		Status string
	}

	SlackMessageReport struct {
		ID      string
		Title   string
		AuditID string
		Status  string
	}

	SlackMessageFile struct {
		ID       string
		Name     string
		Category string
		Status   string
	}

	SlackMessageMetadata struct {
		CompliancePortalID gid.GID
		Documents          []SlackMessageDocument
		Reports            []SlackMessageReport
		Files              []SlackMessageFile
	}
)

func (m SlackMessageMetadata) toMap() map[string]any {
	return map[string]any{
		"compliance_portal_id": m.CompliancePortalID.String(),
		"documents":            m.Documents,
		"reports":              m.Reports,
		"files":                m.Files,
	}
}

// CompliancePortalIDFromMetadata reads the portal a message was raised for.
func CompliancePortalIDFromMetadata(metadata map[string]any) (gid.GID, bool) {
	raw, ok := metadata["compliance_portal_id"].(string)
	if !ok || raw == "" {
		return gid.Nil, false
	}

	portalID, err := gid.ParseGID(raw)
	if err != nil {
		return gid.Nil, false
	}

	return portalID, true
}

// ResolveCompliancePortalID returns the portal encoded in metadata, or the sole
// portal for the organization when legacy messages omit compliance_portal_id.
func (s *Service) ResolveCompliancePortalID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	metadata map[string]any,
) (gid.GID, error) {
	if portalID, ok := CompliancePortalIDFromMetadata(metadata); ok {
		compliancePortal := &coredata.CompliancePortal{}

		err := s.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := compliancePortal.LoadByID(ctx, conn, scope, portalID); err != nil {
					return fmt.Errorf("cannot load compliance portal from metadata: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			return gid.Nil, err
		}

		return portalID, nil
	}

	var (
		portalCount int
		portals     coredata.CompliancePortals
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			portalCount, err = portals.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count compliance portals: %w", err)
			}

			if portalCount != 1 {
				return nil
			}

			cursor := page.NewCursor(
				1,
				nil,
				page.Head,
				page.OrderBy[coredata.CompliancePortalOrderField]{
					Field:     coredata.CompliancePortalOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			)

			return portals.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	if portalCount != 1 {
		return gid.Nil, ErrCompliancePortalMetadataAmbiguous
	}

	if len(portals) != 1 {
		return gid.Nil, fmt.Errorf("cannot resolve compliance portal: expected one portal, found %d", len(portals))
	}

	return portals[0].ID, nil
}

func (s *Service) GetSlackMessageDocumentIDs(
	ctx context.Context,
	scope coredata.Scoper,
	slackMessageID gid.GID,
) (documentIDs []gid.GID, reportIDs []gid.GID, fileIDs []gid.GID, err error) {
	var slackMessage coredata.SlackMessage

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := slackMessage.LoadById(ctx, conn, scope, slackMessageID); err != nil {
				return fmt.Errorf("cannot load slack message: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	documentIDs = extractIDsFromMetadata(slackMessage.Metadata, "documents")
	reportIDs = extractIDsFromMetadata(slackMessage.Metadata, "reports")
	fileIDs = extractIDsFromMetadata(slackMessage.Metadata, "files")

	return documentIDs, reportIDs, fileIDs, nil
}

func (s *Service) UpdateSlackAccessMessage(
	ctx context.Context,
	scope coredata.Scoper,
	slackMessageID gid.GID,
	responseURL string,
	requesterEmail mail.Addr,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var slackMessage coredata.SlackMessage
			if err := slackMessage.LoadById(ctx, tx, scope, slackMessageID); err != nil {
				return fmt.Errorf("cannot load slack message: %w", err)
			}

			var compliancePortal coredata.CompliancePortal

			portalID, err := s.ResolveCompliancePortalID(
				ctx,
				scope,
				slackMessage.OrganizationID,
				slackMessage.Metadata,
			)
			if err != nil {
				return err
			}

			if err := compliancePortal.LoadByID(ctx, tx, scope, portalID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load compliance portal by id: %w", err)
			}

			if compliancePortal.OrganizationID != slackMessage.OrganizationID {
				return fmt.Errorf("compliance portal does not belong to slack message organization")
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByEmail(ctx, tx, requesterEmail); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			var compliancePortalAccess coredata.CompliancePortalAccess
			if err := compliancePortalAccess.LoadByCompliancePortalIDAndIdentityID(ctx, tx, scope, compliancePortal.ID, identity.ID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			documents, reports, files, err := s.loadDocumentsReportsAndFilesFromAccesses(ctx, tx, scope, compliancePortal.ID, compliancePortalAccess.ID)
			if err != nil {
				return err
			}

			newSlackMessageID := gid.New(scope.GetTenantID(), coredata.SlackMessageEntityType)

			updatedBody, err := s.buildAccessRequestMessage(
				newSlackMessageID,
				identity.FullName,
				requesterEmail,
				compliancePortal.OrganizationID,
				compliancePortal.ID,
				documents,
				reports,
				files,
			)
			if err != nil {
				return err
			}

			metadata := SlackMessageMetadata{
				CompliancePortalID: compliancePortal.ID,
				Documents:          documents,
				Reports:            reports,
				Files:              files,
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

			if err := newSlackMessage.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert slack message: %w", err)
			}

			if err := s.GetSlackClient().UpdateInteractiveMessage(ctx, responseURL, updatedBody); err != nil {
				return fmt.Errorf("cannot update Slack message: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) QueueSlackNotification(
	ctx context.Context,
	scope coredata.Scoper,
	identityID gid.GID,
	compliancePortalID gid.GID,
) error {
	return s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var (
			identity               = &coredata.Identity{}
			compliancePortalAccess *coredata.CompliancePortalAccess
		)

		if err := identity.LoadByID(ctx, tx, identityID); err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		compliancePortalAccess = &coredata.CompliancePortalAccess{}
		if err := compliancePortalAccess.LoadByCompliancePortalIDAndIdentityID(ctx, tx, scope, compliancePortalID, identityID); err != nil {
			return fmt.Errorf("cannot load compliance portal access: %w", err)
		}

		var compliancePortal coredata.CompliancePortal
		if err := compliancePortal.LoadByID(ctx, tx, scope, compliancePortalID); err != nil {
			return fmt.Errorf("cannot load compliance portal: %w", err)
		}

		var connectors coredata.Connectors
		if err := connectors.LoadAllByOrganizationIDWithoutDecryptedConnection(
			ctx,
			tx,
			scope,
			compliancePortal.OrganizationID,
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
			return ErrNoSlackConnector
		}

		documents, reports, files, err := s.loadDocumentsReportsAndFilesFromAccesses(ctx, tx, scope, compliancePortalID, compliancePortalAccess.ID)
		if err != nil {
			return fmt.Errorf("cannot load documents, reports and files: %w", err)
		}

		slackMessageID := gid.New(scope.GetTenantID(), coredata.SlackMessageEntityType)

		body, err := s.buildAccessRequestMessage(
			slackMessageID,
			identity.FullName,
			identity.EmailAddress,
			compliancePortal.OrganizationID,
			compliancePortalID,
			documents,
			reports,
			files,
		)
		if err != nil {
			return fmt.Errorf("cannot build access request message: %w", err)
		}

		metadata := SlackMessageMetadata{
			CompliancePortalID: compliancePortalID,
			Documents:          documents,
			Reports:            reports,
			Files:              files,
		}

		now := time.Now()
		slackMessage := &coredata.SlackMessage{
			ID:             slackMessageID,
			OrganizationID: compliancePortal.OrganizationID,
			Type:           coredata.SlackMessageTypeCompliancePortalAccessRequest,
			Body:           body,
			RequesterEmail: &identity.EmailAddress,
			Metadata:       metadata.toMap(),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		sevenDaysAgo := now.Add(-slackMessageDeduplicationWindow)

		var existingMessage coredata.SlackMessage

		err = existingMessage.LoadLatestByRequesterEmailAndType(
			ctx,
			tx,
			scope,
			compliancePortal.OrganizationID,
			identity.EmailAddress,
			coredata.SlackMessageTypeCompliancePortalAccessRequest,
			sevenDaysAgo,
		)
		if err == nil {
			slackMessage.MessageTS = existingMessage.MessageTS
			slackMessage.ChannelID = existingMessage.ChannelID
			slackMessage.InitialSlackMessageID = existingMessage.InitialSlackMessageID

			if err := slackMessage.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert slack message: %w", err)
			}

			return nil
		}

		var notFoundErr coredata.ErrSlackMessageNotFound
		if !errors.Is(err, notFoundErr) {
			return fmt.Errorf("cannot load existing slack message: %w", err)
		}

		slackMessage.InitialSlackMessageID = slackMessageID
		if err := slackMessage.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert slack message: %w", err)
		}

		return nil
	})
}

func (s *Service) loadDocumentsReportsAndFilesFromAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	compliancePortalAccessID gid.GID,
) (
	documents []SlackMessageDocument,
	reports []SlackMessageReport,
	files []SlackMessageFile,
	err error,
) {
	documents = []SlackMessageDocument{}
	reports = []SlackMessageReport{}
	files = []SlackMessageFile{}

	accesses, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.CompliancePortalDocumentAccessOrderField]{
			Field:     coredata.CompliancePortalDocumentAccessOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.CompliancePortalDocumentAccessOrderField]) ([]*coredata.CompliancePortalDocumentAccess, error) {
			var batch coredata.CompliancePortalDocumentAccesses
			if err := batch.LoadByCompliancePortalAccessID(ctx, conn, scope, compliancePortalAccessID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load compliance portal document accesses: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, access := range accesses {
		if access.DocumentID != nil {
			doc := &coredata.Document{}
			if err := doc.LoadByID(ctx, conn, scope, *access.DocumentID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load document: %w", err)
			}

			if doc.CurrentPublishedMajor == nil {
				continue
			}

			portalDocument := &coredata.CompliancePortalDocument{}

			err := portalDocument.LoadByCompliancePortalIDAndDocumentID(ctx, conn, scope, compliancePortalID, *access.DocumentID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					continue
				}

				return nil, nil, nil, fmt.Errorf("cannot load compliance portal document: %w", err)
			}

			if portalDocument.Visibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			documents = append(
				documents,
				SlackMessageDocument{
					ID:     access.DocumentID.String(),
					Title:  doc.Title,
					Status: access.Status.String(),
				},
			)
		}

		if access.ReportFileID != nil {
			audit := &coredata.Audit{}
			if err := audit.LoadByReportFileID(ctx, conn, scope, *access.ReportFileID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load audit: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, scope, audit.FrameworkID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load framework: %w", err)
			}

			label := framework.Name
			if audit.Name != nil && *audit.Name != "" {
				label = label + " - " + *audit.Name
			}

			reports = append(
				reports,
				SlackMessageReport{
					ID:      access.ReportFileID.String(),
					Title:   label,
					AuditID: audit.ID.String(),
					Status:  access.Status.String(),
				},
			)
		}

		if access.CompliancePortalFileID != nil {
			file := &coredata.CompliancePortalFile{}
			if err := file.LoadByID(ctx, conn, scope, *access.CompliancePortalFileID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load compliance portal file: %w", err)
			}

			files = append(
				files,
				SlackMessageFile{
					ID:       access.CompliancePortalFileID.String(),
					Name:     file.Name,
					Category: file.Category,
					Status:   access.Status.String(),
				},
			)
		}
	}

	return documents, reports, files, nil
}

func (s *Service) buildAccessRequestMessage(
	slackMessageID gid.GID,
	requesterName string,
	requesterEmail mail.Addr,
	organizationID gid.GID,
	compliancePortalID gid.GID,
	documents []SlackMessageDocument,
	reports []SlackMessageReport,
	files []SlackMessageFile,
) (map[string]any, error) {
	base, err := baseurl.Parse(s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse base URL: %w", err)
	}

	templateData := struct {
		RequesterName      string
		RequesterEmail     string
		OrganizationID     string
		CompliancePortalID string
		Domain             string
		SlackMessageID     string
		Documents          []SlackMessageDocument
		Reports            []SlackMessageReport
		Files              []SlackMessageFile
	}{
		RequesterName:      requesterName,
		RequesterEmail:     requesterEmail.String(),
		OrganizationID:     organizationID.String(),
		CompliancePortalID: compliancePortalID.String(),
		Domain:             base.Host(),
		SlackMessageID:     slackMessageID.String(),
		Documents:          documents,
		Reports:            reports,
		Files:              files,
	}

	var buf bytes.Buffer
	if err := accessRequestTemplate.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("cannot execute template: %w", err)
	}

	var body map[string]any
	if err := json.NewDecoder(&buf).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot parse template JSON: %w", err)
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
