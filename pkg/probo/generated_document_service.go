// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
)

type GeneratedDocumentService struct {
	svc *TenantService
}

func (s *GeneratedDocumentService) PublishStatementOfApplicability(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
	approverIDs []gid.GID,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			soa := &coredata.StatementOfApplicability{}
			if err := soa.LoadByID(ctx, tx, s.svc.scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement of applicability: %w", err)
			}

			documentData, err := s.buildStatementOfApplicabilityDocumentData(ctx, tx, soa)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildStatementOfApplicabilityDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			var existingDoc *coredata.Document
			if soa.DocumentID != nil {
				doc := &coredata.Document{}
				err = doc.LoadByID(ctx, tx, s.svc.scope, *soa.DocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load statement of applicability document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					soa.DocumentID = nil
					soa.UpdatedAt = now
					if err := soa.Update(ctx, tx, s.svc.scope); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			hasApprovers := len(approverIDs) > 0

			if existingDoc == nil {
				documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        soa.OrganizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				soa.DocumentID = &documentID
				soa.UpdatedAt = now
				if err := soa.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update document reference: %w", err)
				}
			} else {
				document = existingDoc
			}

			var newMajor int
			if document.CurrentPublishedMajor != nil {
				newMajor = *document.CurrentPublishedMajor + 1
			} else {
				newMajor = 1
			}

			versionStatus := coredata.DocumentVersionStatusPublished
			var publishedAt *time.Time
			if hasApprovers {
				versionStatus = coredata.DocumentVersionStatusDraft
			} else {
				publishedAt = &now
			}

			documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: soa.OrganizationID,
				DocumentID:     document.ID,
				Title:          soa.Name,
				Major:          newMajor,
				Minor:          0,
				Content:        prosemirrorJSON,
				Status:         versionStatus,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeStatementOfApplicability,
				Orientation:    coredata.DocumentVersionOrientationLandscape,
				PublishedAt:    publishedAt,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := documentVersion.Insert(ctx, tx, s.svc.scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return fmt.Errorf("a version is pending approval, approve or reject it before publishing a new one: %w", err)
				}
				return fmt.Errorf("cannot insert document version: %w", err)
			}

			if hasApprovers {
				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, document.ID, soa.OrganizationID, approverIDs); err != nil {
					return fmt.Errorf("cannot save default approvers: %w", err)
				}

				_, err := s.svc.DocumentApprovals.RequestApprovalInTx(
					ctx,
					tx,
					document,
					documentVersion,
					approverIDs,
					nil,
				)
				if err != nil {
					return fmt.Errorf("cannot request approval: %w", err)
				}
			} else {
				document.CurrentPublishedMajor = &newMajor
				document.CurrentPublishedMinor = new(0)
				document.UpdatedAt = now

				if err := document.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update document: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) buildStatementOfApplicabilityDocumentData(
	ctx context.Context,
	conn pg.Querier,
	statementOfApplicability *coredata.StatementOfApplicability,
) (docgen.StatementOfApplicabilityData, error) {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, conn, s.svc.scope, statementOfApplicability.OrganizationID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load organization: %w", err)
	}

	var applicabilityStatements coredata.ApplicabilityStatements
	if err := applicabilityStatements.LoadAllByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicability.ID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load applicability statements: %w", err)
	}

	if len(applicabilityStatements) == 0 {
		return docgen.StatementOfApplicabilityData{
			Title:            statementOfApplicability.Name,
			OrganizationName: organization.Name,
			CreatedAt:        statementOfApplicability.CreatedAt,
			TotalControls:    0,
		}, nil
	}

	controlIDs := make([]gid.GID, len(applicabilityStatements))
	for i, stmt := range applicabilityStatements {
		controlIDs[i] = stmt.ControlID
	}

	var controls coredata.Controls
	if err := controls.LoadByIDs(ctx, conn, s.svc.scope, controlIDs); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load controls: %w", err)
	}

	controlMap := make(map[gid.GID]*coredata.Control, len(controls))
	frameworkIDSet := make(map[gid.GID]struct{})
	for _, c := range controls {
		controlMap[c.ID] = c
		frameworkIDSet[c.FrameworkID] = struct{}{}
	}

	frameworkIDs := make([]gid.GID, 0, len(frameworkIDSet))
	for id := range frameworkIDSet {
		frameworkIDs = append(frameworkIDs, id)
	}

	var frameworks coredata.Frameworks
	if err := frameworks.LoadByIDs(ctx, conn, s.svc.scope, frameworkIDs); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load frameworks: %w", err)
	}

	frameworkMap := make(map[gid.GID]*coredata.Framework, len(frameworks))
	for _, f := range frameworks {
		frameworkMap[f.ID] = f
	}

	controlOblTypes, err := coredata.LoadObligationTypesByControlIDs(ctx, conn, s.svc.scope, controlIDs)
	if err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load obligation types: %w", err)
	}

	type obligationKey struct {
		controlID gid.GID
		oblType   coredata.ObligationType
	}
	oblSet := make(map[obligationKey]struct{}, len(controlOblTypes))
	for _, co := range controlOblTypes {
		oblSet[obligationKey{co.ControlID, co.ObligationType}] = struct{}{}
	}

	var controlsWithRisk coredata.ControlsWithRisk
	if err := controlsWithRisk.LoadByControlIDs(ctx, conn, s.svc.scope, controlIDs); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load controls with risks: %w", err)
	}

	riskSet := make(map[gid.GID]struct{}, len(controlsWithRisk))
	for _, cwr := range controlsWithRisk {
		riskSet[cwr.ControlID] = struct{}{}
	}

	rows := make([]docgen.SOARow, 0, len(applicabilityStatements))

	for _, stmt := range applicabilityStatements {
		control := controlMap[stmt.ControlID]
		if control == nil {
			continue
		}
		framework := frameworkMap[control.FrameworkID]
		if framework == nil {
			continue
		}

		applicable := stmt.Applicability

		justification := "-"
		if !applicable && stmt.Justification != nil {
			justification = *stmt.Justification
		}

		notImplJustification := "-"
		if applicable && control.MaturityLevel == coredata.ControlMaturityLevelNone && control.NotImplementedJustification != nil {
			notImplJustification = *control.NotImplementedJustification
		}

		regulatory := "-"
		contractual := "-"
		bestPractice := "-"
		riskAssessment := "-"
		if applicable {
			_, hasLegal := oblSet[obligationKey{stmt.ControlID, coredata.ObligationTypeLegal}]
			regulatory = docgen.BoolLabel(hasLegal)
			_, hasContractual := oblSet[obligationKey{stmt.ControlID, coredata.ObligationTypeContractual}]
			contractual = docgen.BoolLabel(hasContractual)
			bestPractice = docgen.BoolLabel(control.BestPractice)
			_, hasRisk := riskSet[stmt.ControlID]
			riskAssessment = docgen.BoolLabel(hasRisk)
		}

		maturityLevel := "-"
		if applicable {
			maturityLevel = docgen.MaturityLabel(control.MaturityLevel)
		}

		rows = append(rows, docgen.SOARow{
			FrameworkName:        framework.Name,
			ControlSection:       control.SectionTitle,
			ControlName:          control.Name,
			Applicability:        docgen.BoolLabel(applicable),
			Justification:        justification,
			MaturityLevel:        maturityLevel,
			NotImplJustification: notImplJustification,
			Regulatory:           regulatory,
			Contractual:          contractual,
			BestPractice:         bestPractice,
			RiskAssessment:       riskAssessment,
		})
	}

	return docgen.StatementOfApplicabilityData{
		Title:            statementOfApplicability.Name,
		OrganizationName: organization.Name,
		CreatedAt:        statementOfApplicability.CreatedAt,
		TotalControls:    len(applicabilityStatements),
		Rows:             rows,
	}, nil
}

func (s *GeneratedDocumentService) PublishDataList(
	ctx context.Context,
	organizationID gid.GID,
	approverIDs []gid.GID,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildDataListDocumentData(ctx, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildDataListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			datum := coredata.Datum{}
			dataDocumentID, err := datum.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document
			if dataDocumentID != nil {
				doc := &coredata.Document{}
				err = doc.LoadByID(ctx, tx, s.svc.scope, *dataDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load data list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := datum.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*dataDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			hasApprovers := len(approverIDs) > 0

			if existingDoc == nil {
				documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := datum.UpsertGeneratedDocumentID(ctx, tx, organizationID, s.svc.scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			var newMajor int
			if document.CurrentPublishedMajor != nil {
				newMajor = *document.CurrentPublishedMajor + 1
			} else {
				newMajor = 1
			}

			versionStatus := coredata.DocumentVersionStatusPublished
			var publishedAt *time.Time
			if hasApprovers {
				versionStatus = coredata.DocumentVersionStatusDraft
			} else {
				publishedAt = &now
			}

			documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Data List",
				Major:          newMajor,
				Minor:          0,
				Content:        prosemirrorJSON,
				Status:         versionStatus,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				PublishedAt:    publishedAt,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := documentVersion.Insert(ctx, tx, s.svc.scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return fmt.Errorf("a version is pending approval, approve or reject it before publishing a new one: %w", err)
				}
				return fmt.Errorf("cannot insert document version: %w", err)
			}

			if hasApprovers {
				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, document.ID, organizationID, approverIDs); err != nil {
					return fmt.Errorf("cannot save default approvers: %w", err)
				}

				_, err := s.svc.DocumentApprovals.RequestApprovalInTx(
					ctx,
					tx,
					document,
					documentVersion,
					approverIDs,
					nil,
				)
				if err != nil {
					return fmt.Errorf("cannot request approval: %w", err)
				}
			} else {
				document.CurrentPublishedMajor = &newMajor
				document.CurrentPublishedMinor = new(0)
				document.UpdatedAt = now

				if err := document.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update document: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetDataListDocumentID(
	ctx context.Context,
	organizationID gid.GID,
) (*gid.GID, error) {
	var dataDocumentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		datum := coredata.Datum{}
		var err error
		dataDocumentID, err = datum.GetGeneratedDocumentID(ctx, conn, organizationID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get data list document ID: %w", err)
	}

	return dataDocumentID, nil
}

func (s *GeneratedDocumentService) buildDataListDocumentData(
	ctx context.Context,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.DataListData, error) {
	var data coredata.Data
	if err := data.LoadAllByOrganizationID(ctx, conn, s.svc.scope, organization.ID); err != nil {
		return docgen.DataListData{}, fmt.Errorf("cannot load data: %w", err)
	}

	if len(data) == 0 {
		return docgen.DataListData{
			Title:            "Data List",
			OrganizationName: organization.Name,
			CreatedAt:        time.Now(),
			TotalData:        0,
		}, nil
	}

	ownerIDs := make([]gid.GID, 0, len(data))
	ownerIDSet := make(map[gid.GID]struct{})
	for _, d := range data {
		if _, ok := ownerIDSet[d.OwnerID]; !ok {
			ownerIDs = append(ownerIDs, d.OwnerID)
			ownerIDSet[d.OwnerID] = struct{}{}
		}
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadByIDs(ctx, conn, s.svc.scope, ownerIDs); err != nil {
		return docgen.DataListData{}, fmt.Errorf("cannot load profiles: %w", err)
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	rows := make([]docgen.DataListRow, 0, len(data))
	for _, d := range data {
		ownerName := "-"
		if p, ok := profileMap[d.OwnerID]; ok {
			ownerName = p.FullName
		}

		var vendors coredata.Vendors
		if err := vendors.LoadAllByDatumID(ctx, conn, s.svc.scope, d.ID); err != nil {
			return docgen.DataListData{}, fmt.Errorf("cannot load vendors for datum %s: %w", d.ID, err)
		}

		vendorNames := make([]string, 0, len(vendors))
		for _, v := range vendors {
			vendorNames = append(vendorNames, v.Name)
		}

		vendorStr := "-"
		if len(vendorNames) > 0 {
			vendorStr = strings.Join(vendorNames, ", ")
		}

		rows = append(rows, docgen.DataListRow{
			Name:           d.Name,
			Classification: formatClassification(d.DataClassification),
			Owner:          ownerName,
			Vendors:        vendorStr,
		})
	}

	return docgen.DataListData{
		Title:            "Data List",
		OrganizationName: organization.Name,
		CreatedAt:        time.Now(),
		TotalData:        len(data),
		Rows:             rows,
	}, nil
}

func formatClassification(c coredata.DataClassification) string {
	switch c {
	case coredata.DataClassificationPublic:
		return "Public"
	case coredata.DataClassificationInternal:
		return "Internal"
	case coredata.DataClassificationConfidential:
		return "Confidential"
	case coredata.DataClassificationSecret:
		return "Secret"
	default:
		return string(c)
	}
}

var dataListTemplate = template.Must(
	template.New("data_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(b), nil
			},
		}).
		ParseFS(Templates, "templates/data_list.json.tmpl"),
)

func BuildDataListDocument(data docgen.DataListData) (string, error) {
	var buf bytes.Buffer
	if err := dataListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute data list template: %w", err)
	}
	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishAssetList(
	ctx context.Context,
	organizationID gid.GID,
	approverIDs []gid.GID,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildAssetListDocumentData(ctx, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildAssetListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			asset := coredata.Asset{}
			assetDocumentID, err := asset.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document
			if assetDocumentID != nil {
				doc := &coredata.Document{}
				err = doc.LoadByID(ctx, tx, s.svc.scope, *assetDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load asset list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := asset.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*assetDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			hasApprovers := len(approverIDs) > 0

			if existingDoc == nil {
				documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := asset.UpsertGeneratedDocumentID(ctx, tx, organizationID, s.svc.scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			var newMajor int
			if document.CurrentPublishedMajor != nil {
				newMajor = *document.CurrentPublishedMajor + 1
			} else {
				newMajor = 1
			}

			versionStatus := coredata.DocumentVersionStatusPublished
			var publishedAt *time.Time
			if hasApprovers {
				versionStatus = coredata.DocumentVersionStatusDraft
			} else {
				publishedAt = &now
			}

			documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Asset List",
				Major:          newMajor,
				Minor:          0,
				Content:        prosemirrorJSON,
				Status:         versionStatus,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				PublishedAt:    publishedAt,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := documentVersion.Insert(ctx, tx, s.svc.scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return fmt.Errorf("a version is pending approval, approve or reject it before publishing a new one: %w", err)
				}
				return fmt.Errorf("cannot insert document version: %w", err)
			}

			if hasApprovers {
				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, document.ID, organizationID, approverIDs); err != nil {
					return fmt.Errorf("cannot save default approvers: %w", err)
				}

				_, err := s.svc.DocumentApprovals.RequestApprovalInTx(
					ctx,
					tx,
					document,
					documentVersion,
					approverIDs,
					nil,
				)
				if err != nil {
					return fmt.Errorf("cannot request approval: %w", err)
				}
			} else {
				document.CurrentPublishedMajor = &newMajor
				document.CurrentPublishedMinor = new(0)
				document.UpdatedAt = now

				if err := document.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update document: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetAssetListDocumentID(
	ctx context.Context,
	organizationID gid.GID,
) (*gid.GID, error) {
	var assetDocumentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		asset := coredata.Asset{}
		var err error
		assetDocumentID, err = asset.GetGeneratedDocumentID(ctx, conn, organizationID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get asset list document ID: %w", err)
	}

	return assetDocumentID, nil
}

func (s *GeneratedDocumentService) buildAssetListDocumentData(
	ctx context.Context,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.AssetListData, error) {
	var assets coredata.Assets
	if err := assets.LoadAllByOrganizationID(ctx, conn, s.svc.scope, organization.ID); err != nil {
		return docgen.AssetListData{}, fmt.Errorf("cannot load assets: %w", err)
	}

	if len(assets) == 0 {
		return docgen.AssetListData{
			Title:            "Asset List",
			OrganizationName: organization.Name,
			CreatedAt:        time.Now(),
			TotalAssets:      0,
		}, nil
	}

	ownerIDs := make([]gid.GID, 0, len(assets))
	ownerIDSet := make(map[gid.GID]struct{})
	for _, a := range assets {
		if _, ok := ownerIDSet[a.OwnerID]; !ok {
			ownerIDs = append(ownerIDs, a.OwnerID)
			ownerIDSet[a.OwnerID] = struct{}{}
		}
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadByIDs(ctx, conn, s.svc.scope, ownerIDs); err != nil {
		return docgen.AssetListData{}, fmt.Errorf("cannot load profiles: %w", err)
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	rows := make([]docgen.AssetListRow, 0, len(assets))
	for _, a := range assets {
		ownerName := "-"
		if p, ok := profileMap[a.OwnerID]; ok {
			ownerName = p.FullName
		}

		var vendors coredata.Vendors
		if err := vendors.LoadAllByAssetID(ctx, conn, s.svc.scope, a.ID); err != nil {
			return docgen.AssetListData{}, fmt.Errorf("cannot load vendors for asset %s: %w", a.ID, err)
		}

		vendorNames := make([]string, 0, len(vendors))
		for _, v := range vendors {
			vendorNames = append(vendorNames, v.Name)
		}

		vendorStr := "-"
		if len(vendorNames) > 0 {
			vendorStr = strings.Join(vendorNames, ", ")
		}

		rows = append(rows, docgen.AssetListRow{
			Name:            a.Name,
			AssetType:       formatAssetType(a.AssetType),
			Amount:          a.Amount,
			DataTypesStored: a.DataTypesStored,
			Owner:           ownerName,
			Vendors:         vendorStr,
		})
	}

	return docgen.AssetListData{
		Title:            "Asset List",
		OrganizationName: organization.Name,
		CreatedAt:        time.Now(),
		TotalAssets:      len(assets),
		Rows:             rows,
	}, nil
}

func formatAssetType(t coredata.AssetType) string {
	switch t {
	case coredata.AssetTypePhysical:
		return "Physical"
	case coredata.AssetTypeVirtual:
		return "Virtual"
	default:
		return string(t)
	}
}

var assetListTemplate = template.Must(
	template.New("asset_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(b), nil
			},
			"printf": fmt.Sprintf,
		}).
		ParseFS(Templates, "templates/asset_list.json.tmpl"),
)

func BuildAssetListDocument(data docgen.AssetListData) (string, error) {
	var buf bytes.Buffer
	if err := assetListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute asset list template: %w", err)
	}
	return buf.String(), nil
}

var soaTemplate = template.Must(
	template.New("statement_of_applicability.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(b), nil
			},
		}).
		ParseFS(Templates, "templates/statement_of_applicability.json.tmpl"),
)

func BuildStatementOfApplicabilityDocument(data docgen.StatementOfApplicabilityData) (string, error) {
	var buf bytes.Buffer
	if err := soaTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute soa template: %w", err)
	}
	return buf.String(), nil
}
