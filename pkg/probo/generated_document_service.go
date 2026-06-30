// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/page"
)

type GeneratedDocumentService struct {
	svc *Service
}

func (s *GeneratedDocumentService) PublishStatementOfApplicability(
	ctx context.Context, scope coredata.Scoper,
	statementOfApplicabilityID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			soa := &coredata.StatementOfApplicability{}
			if err := soa.LoadByID(ctx, tx, scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement of applicability: %w", err)
			}

			documentData, err := s.buildStatementOfApplicabilityDocumentData(ctx, scope, tx, soa)
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

				err = doc.LoadByID(ctx, tx, scope, *soa.DocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load statement of applicability document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					soa.DocumentID = nil

					soa.UpdatedAt = now
					if err := soa.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        soa.OrganizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				soa.DocumentID = &documentID

				soa.UpdatedAt = now
				if err := soa.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update document reference: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: soa.OrganizationID,
				DocumentID:     document.ID,
				Title:          soa.Name,
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeStatementOfApplicability,
				Orientation:    coredata.DocumentVersionOrientationLandscape,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, soa.OrganizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) buildStatementOfApplicabilityDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	statementOfApplicability *coredata.StatementOfApplicability,
) (docgen.StatementOfApplicabilityData, error) {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, conn, scope, statementOfApplicability.OrganizationID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load organization: %w", err)
	}

	applicabilityStatements, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ApplicabilityStatementOrderField]{
			Field:     coredata.ApplicabilityStatementOrderFieldControlSectionTitle,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ApplicabilityStatementOrderField]) ([]*coredata.ApplicabilityStatement, error) {
			var batch coredata.ApplicabilityStatements
			if err := batch.LoadByStatementOfApplicabilityID(ctx, conn, scope, statementOfApplicability.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load applicability statements: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.StatementOfApplicabilityData{}, err
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
	if err := controls.LoadByIDs(ctx, conn, scope, controlIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
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
	if err := frameworks.LoadByIDs(ctx, conn, scope, frameworkIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load frameworks: %w", err)
	}

	frameworkMap := make(map[gid.GID]*coredata.Framework, len(frameworks))
	for _, f := range frameworks {
		frameworkMap[f.ID] = f
	}

	controlOblTypes, err := coredata.LoadObligationTypesByControlIDs(ctx, conn, scope, controlIDs)
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
	if err := controlsWithRisk.LoadByControlIDs(ctx, conn, scope, controlIDs); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildDataListDocumentData(ctx, scope, tx, organization)
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

				err = doc.LoadByID(ctx, tx, scope, *dataDocumentID)
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

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := datum.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Data",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetDataListDocumentID(
	ctx context.Context, scope coredata.Scoper,
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
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.DataListData, error) {
	data, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.DatumOrderField]{
			Field:     coredata.DatumOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DatumOrderField]) ([]*coredata.Datum, error) {
			var batch coredata.Data
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load data: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.DataListData{}, err
	}

	if len(data) == 0 {
		return docgen.DataListData{
			Title:            "Data",
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
	if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return docgen.DataListData{}, fmt.Errorf("cannot load profiles: %w", err)
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	rows := make([]docgen.DataListRow, 0, len(data))
	for _, d := range data {
		ownerName := "-"
		if p, ok := profileMap[d.OwnerID]; ok && p.FullName != "" {
			ownerName = p.FullName
		}

		thirdParties, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.ThirdPartyOrderField]{
				Field:     coredata.ThirdPartyOrderFieldName,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.ThirdPartyOrderField]) ([]*coredata.ThirdParty, error) {
				var batch coredata.ThirdParties
				if err := batch.LoadByDatumID(ctx, conn, scope, d.ID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load thirdParties for datum %s: %w", d.ID, err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return docgen.DataListData{}, err
		}

		thirdPartyNames := make([]string, 0, len(thirdParties))
		for _, v := range thirdParties {
			thirdPartyNames = append(thirdPartyNames, v.Name)
		}

		thirdPartyStr := "-"
		if len(thirdPartyNames) > 0 {
			thirdPartyStr = strings.Join(thirdPartyNames, ", ")
		}

		rows = append(rows, docgen.DataListRow{
			Name:           d.Name,
			Classification: formatClassification(d.DataClassification),
			Owner:          ownerName,
			ThirdParties:   thirdPartyStr,
		})
	}

	return docgen.DataListData{
		Title:            "Data",
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
		return stringOrNotSpecified(string(c))
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildAssetListDocumentData(ctx, scope, tx, organization)
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

				err = doc.LoadByID(ctx, tx, scope, *assetDocumentID)
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

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := asset.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Assets",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetAssetListDocumentID(
	ctx context.Context, scope coredata.Scoper,
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
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.AssetListData, error) {
	assets, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.AssetOrderField]{
			Field:     coredata.AssetOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.AssetOrderField]) ([]*coredata.Asset, error) {
			var batch coredata.Assets
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load assets: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.AssetListData{}, err
	}

	if len(assets) == 0 {
		return docgen.AssetListData{
			Title:            "Assets",
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
	if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return docgen.AssetListData{}, fmt.Errorf("cannot load profiles: %w", err)
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	rows := make([]docgen.AssetListRow, 0, len(assets))
	for _, a := range assets {
		ownerName := "-"
		if p, ok := profileMap[a.OwnerID]; ok && p.FullName != "" {
			ownerName = p.FullName
		}

		thirdParties, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.ThirdPartyOrderField]{
				Field:     coredata.ThirdPartyOrderFieldName,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.ThirdPartyOrderField]) ([]*coredata.ThirdParty, error) {
				var batch coredata.ThirdParties
				if err := batch.LoadByAssetID(ctx, conn, scope, a.ID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load thirdParties for asset %s: %w", a.ID, err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return docgen.AssetListData{}, err
		}

		thirdPartyNames := make([]string, 0, len(thirdParties))
		for _, v := range thirdParties {
			thirdPartyNames = append(thirdPartyNames, v.Name)
		}

		thirdPartyStr := "-"
		if len(thirdPartyNames) > 0 {
			thirdPartyStr = strings.Join(thirdPartyNames, ", ")
		}

		rows = append(rows, docgen.AssetListRow{
			Name:            a.Name,
			AssetType:       formatAssetType(a.AssetType),
			Amount:          a.Amount,
			DataTypesStored: stringOrNotSpecified(a.DataTypesStored),
			Owner:           ownerName,
			ThirdParties:    thirdPartyStr,
		})
	}

	return docgen.AssetListData{
		Title:            "Assets",
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
		return stringOrNotSpecified(string(t))
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

func (s *GeneratedDocumentService) PublishFindingList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildFindingListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildFindingListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			finding := coredata.Finding{}

			findingDocumentID, err := finding.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if findingDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *findingDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load finding list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := finding.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*findingDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := finding.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Findings",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationLandscape,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetFindingsDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var findingDocumentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		finding := coredata.Finding{}

		var err error

		findingDocumentID, err = finding.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get finding list document ID: %w", err)
	}

	return findingDocumentID, nil
}

func (s *GeneratedDocumentService) buildFindingListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.FindingListData, error) {
	findings, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.FindingOrderField]{
			Field:     coredata.FindingOrderFieldReferenceId,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.FindingOrderField]) ([]*coredata.Finding, error) {
			var batch coredata.Findings
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor, coredata.NewFindingFilter(nil, nil, nil, nil)); err != nil {
				return nil, fmt.Errorf("cannot load findings: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.FindingListData{}, err
	}

	if len(findings) == 0 {
		return docgen.FindingListData{
			Title:            "Findings",
			OrganizationName: organization.Name,
			CreatedAt:        time.Now(),
			TotalFindings:    0,
		}, nil
	}

	ownerIDs := make([]gid.GID, 0, len(findings))
	ownerIDSet := make(map[gid.GID]struct{})

	for _, f := range findings {
		if f.OwnerID != nil {
			if _, ok := ownerIDSet[*f.OwnerID]; !ok {
				ownerIDs = append(ownerIDs, *f.OwnerID)
				ownerIDSet[*f.OwnerID] = struct{}{}
			}
		}
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile)

	if len(ownerIDs) > 0 {
		var profiles coredata.MembershipProfiles
		if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return docgen.FindingListData{}, fmt.Errorf("cannot load profiles: %w", err)
		}

		for _, p := range profiles {
			profileMap[p.ID] = p
		}
	}

	rows := make([]docgen.FindingListRow, 0, len(findings))
	for _, f := range findings {
		ownerName := "-"

		if f.OwnerID != nil {
			if p, ok := profileMap[*f.OwnerID]; ok && p.FullName != "" {
				ownerName = p.FullName
			}
		}

		description := "-"
		if f.Description != nil && *f.Description != "" {
			description = *f.Description
		}

		source := "-"
		if f.Source != nil && *f.Source != "" {
			source = *f.Source
		}

		identifiedOn := "-"
		if f.IdentifiedOn != nil {
			identifiedOn = f.IdentifiedOn.Format("2006-01-02")
		}

		rootCause := "-"
		if f.RootCause != nil && *f.RootCause != "" {
			rootCause = *f.RootCause
		}

		correctiveAction := "-"
		if f.CorrectiveAction != nil && *f.CorrectiveAction != "" {
			correctiveAction = *f.CorrectiveAction
		}

		effectivenessCheck := "-"
		if f.EffectivenessCheck != nil && *f.EffectivenessCheck != "" {
			effectivenessCheck = *f.EffectivenessCheck
		}

		dueDate := "-"
		if f.DueDate != nil {
			dueDate = f.DueDate.Format("2006-01-02")
		}

		rows = append(rows, docgen.FindingListRow{
			ReferenceID:        f.ReferenceID,
			Kind:               formatFindingKind(f.Kind),
			Description:        description,
			Source:             source,
			IdentifiedOn:       identifiedOn,
			RootCause:          rootCause,
			CorrectiveAction:   correctiveAction,
			EffectivenessCheck: effectivenessCheck,
			Status:             formatFindingStatus(f.Status),
			Priority:           formatFindingPriority(f.Priority),
			Owner:              ownerName,
			DueDate:            dueDate,
		})
	}

	return docgen.FindingListData{
		Title:            "Findings",
		OrganizationName: organization.Name,
		CreatedAt:        time.Now(),
		TotalFindings:    len(findings),
		Rows:             rows,
	}, nil
}

func formatFindingKind(k coredata.FindingKind) string {
	switch k {
	case coredata.FindingKindMinorNonconformity:
		return "Minor Nonconformity"
	case coredata.FindingKindMajorNonconformity:
		return "Major Nonconformity"
	case coredata.FindingKindObservation:
		return "Observation"
	case coredata.FindingKindException:
		return "Exception"
	default:
		return stringOrNotSpecified(string(k))
	}
}

func formatFindingStatus(s coredata.FindingStatus) string {
	switch s {
	case coredata.FindingStatusOpen:
		return "Open"
	case coredata.FindingStatusInProgress:
		return "In Progress"
	case coredata.FindingStatusClosed:
		return "Closed"
	case coredata.FindingStatusRiskAccepted:
		return "Risk Accepted"
	case coredata.FindingStatusMitigated:
		return "Mitigated"
	case coredata.FindingStatusFalsePositive:
		return "False Positive"
	default:
		return stringOrNotSpecified(string(s))
	}
}

func formatFindingPriority(p coredata.FindingPriority) string {
	switch p {
	case coredata.FindingPriorityLow:
		return "Low"
	case coredata.FindingPriorityMedium:
		return "Medium"
	case coredata.FindingPriorityHigh:
		return "High"
	default:
		return stringOrNotSpecified(string(p))
	}
}

var findingListTemplate = template.Must(
	template.New("finding_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
		}).
		ParseFS(Templates, "templates/finding_list.json.tmpl"),
)

func BuildFindingListDocument(data docgen.FindingListData) (string, error) {
	var buf bytes.Buffer
	if err := findingListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute finding list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishObligationList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildObligationListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildObligationListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			obligation := coredata.Obligation{}

			obligationDocumentID, err := obligation.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if obligationDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *obligationDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load obligation list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := obligation.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*obligationDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := obligation.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Obligations",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationLandscape,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetObligationsDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var obligationDocumentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		obligation := coredata.Obligation{}

		var err error

		obligationDocumentID, err = obligation.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get obligation list document ID: %w", err)
	}

	return obligationDocumentID, nil
}

func (s *GeneratedDocumentService) buildObligationListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.ObligationListData, error) {
	obligations, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ObligationOrderField]{
			Field:     coredata.ObligationOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ObligationOrderField]) ([]*coredata.Obligation, error) {
			var batch coredata.Obligations
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load obligations: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.ObligationListData{}, err
	}

	if len(obligations) == 0 {
		return docgen.ObligationListData{
			Title:            "Obligations",
			OrganizationName: organization.Name,
			CreatedAt:        time.Now(),
			TotalObligations: 0,
		}, nil
	}

	ownerIDs := make([]gid.GID, 0, len(obligations))
	ownerIDSet := make(map[gid.GID]struct{})

	for _, o := range obligations {
		if o.OwnerID == gid.Nil {
			continue
		}

		if _, ok := ownerIDSet[o.OwnerID]; !ok {
			ownerIDs = append(ownerIDs, o.OwnerID)
			ownerIDSet[o.OwnerID] = struct{}{}
		}
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile)

	if len(ownerIDs) > 0 {
		var profiles coredata.MembershipProfiles
		if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return docgen.ObligationListData{}, fmt.Errorf("cannot load profiles: %w", err)
		}

		for _, p := range profiles {
			profileMap[p.ID] = p
		}
	}

	rows := make([]docgen.ObligationListRow, 0, len(obligations))
	for _, o := range obligations {
		ownerName := "-"
		if p, ok := profileMap[o.OwnerID]; ok && p.FullName != "" {
			ownerName = p.FullName
		}

		area := "-"
		if o.Area != nil && *o.Area != "" {
			area = *o.Area
		}

		source := "-"
		if o.Source != nil && *o.Source != "" {
			source = *o.Source
		}

		requirement := "-"
		if o.Requirement != nil && *o.Requirement != "" {
			requirement = *o.Requirement
		}

		actionsToBeImplemented := "-"
		if o.ActionsToBeImplemented != nil && *o.ActionsToBeImplemented != "" {
			actionsToBeImplemented = *o.ActionsToBeImplemented
		}

		regulator := "-"
		if o.Regulator != nil && *o.Regulator != "" {
			regulator = *o.Regulator
		}

		dueDate := "-"
		if o.DueDate != nil {
			dueDate = o.DueDate.Format("2006-01-02")
		}

		rows = append(rows, docgen.ObligationListRow{
			Area:                   area,
			Source:                 source,
			Requirement:            requirement,
			ActionsToBeImplemented: actionsToBeImplemented,
			Status:                 formatObligationStatus(o.Status),
			Type:                   formatObligationType(o.Type),
			Regulator:              regulator,
			Owner:                  ownerName,
			DueDate:                dueDate,
		})
	}

	return docgen.ObligationListData{
		Title:            "Obligations",
		OrganizationName: organization.Name,
		CreatedAt:        time.Now(),
		TotalObligations: len(obligations),
		Rows:             rows,
	}, nil
}

func formatObligationStatus(s coredata.ObligationStatus) string {
	switch s {
	case coredata.ObligationStatusNonCompliant:
		return "Non Compliant"
	case coredata.ObligationStatusPartiallyCompliant:
		return "Partially Compliant"
	case coredata.ObligationStatusCompliant:
		return "Compliant"
	default:
		return stringOrNotSpecified(string(s))
	}
}

func formatObligationType(t coredata.ObligationType) string {
	switch t {
	case coredata.ObligationTypeLegal:
		return "Legal"
	case coredata.ObligationTypeContractual:
		return "Contractual"
	default:
		return stringOrNotSpecified(string(t))
	}
}

var obligationListTemplate = template.Must(
	template.New("obligation_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
		}).
		ParseFS(Templates, "templates/obligation_list.json.tmpl"),
)

func BuildObligationListDocument(data docgen.ObligationListData) (string, error) {
	var buf bytes.Buffer
	if err := obligationListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute obligation list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishProcessingActivityList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildProcessingActivityListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildProcessingActivityListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			processingActivity := coredata.ProcessingActivity{}

			processingActivityDocumentID, err := processingActivity.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if processingActivityDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *processingActivityDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load processing activity list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := processingActivity.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*processingActivityDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := processingActivity.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Processing Activities",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetProcessingActivitiesDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		processingActivity := coredata.ProcessingActivity{}

		var err error

		documentID, err = processingActivity.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get processing activity list document ID: %w", err)
	}

	return documentID, nil
}

func (s *GeneratedDocumentService) buildProcessingActivityListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.ProcessingActivityListData, error) {
	processingActivities, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ProcessingActivityOrderField]{
			Field:     coredata.ProcessingActivityOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ProcessingActivityOrderField]) ([]*coredata.ProcessingActivity, error) {
			var batch coredata.ProcessingActivities
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load processing activities: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.ProcessingActivityListData{}, err
	}

	if len(processingActivities) == 0 {
		return docgen.ProcessingActivityListData{
			Title:                     "Processing Activities",
			OrganizationName:          organization.Name,
			CreatedAt:                 time.Now(),
			TotalProcessingActivities: 0,
		}, nil
	}

	var thirdParties coredata.ThirdParties

	thirdPartyMap, err := thirdParties.LoadAllByProcessingActivities(ctx, conn, scope, organization.ID)
	if err != nil {
		return docgen.ProcessingActivityListData{}, fmt.Errorf("cannot load thirdParties: %w", err)
	}

	dpoIDs := make([]gid.GID, 0, len(processingActivities))
	dpoIDSet := make(map[gid.GID]struct{})

	for _, pa := range processingActivities {
		if pa.DataProtectionOfficerID != nil {
			if _, ok := dpoIDSet[*pa.DataProtectionOfficerID]; !ok {
				dpoIDs = append(dpoIDs, *pa.DataProtectionOfficerID)
				dpoIDSet[*pa.DataProtectionOfficerID] = struct{}{}
			}
		}
	}

	dpoMap := make(map[gid.GID]*coredata.MembershipProfile)

	if len(dpoIDs) > 0 {
		var profiles coredata.MembershipProfiles
		if err := profiles.LoadByIDs(ctx, conn, scope, dpoIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return docgen.ProcessingActivityListData{}, fmt.Errorf("cannot load DPO profiles: %w", err)
		}

		for _, p := range profiles {
			dpoMap[p.ID] = p
		}
	}

	rows := make([]docgen.ProcessingActivityListRow, 0, len(processingActivities))
	for _, pa := range processingActivities {
		dpoName := "Not assigned"

		if pa.DataProtectionOfficerID != nil {
			if p, ok := dpoMap[*pa.DataProtectionOfficerID]; ok && p.FullName != "" {
				dpoName = p.FullName
			}
		}

		thirdPartyStr := "None"
		if thirdPartyNames, ok := thirdPartyMap[pa.ID]; ok && len(thirdPartyNames) > 0 {
			thirdPartyStr = strings.Join(thirdPartyNames, ", ")
		}

		rows = append(rows, docgen.ProcessingActivityListRow{
			Name:                                 pa.Name,
			Purpose:                              derefStringOrNotSpecified(pa.Purpose),
			Role:                                 formatProcessingActivityRole(pa.Role),
			DataSubjectCategory:                  derefStringOrNotSpecified(pa.DataSubjectCategory),
			PersonalDataCategory:                 derefStringOrNotSpecified(pa.PersonalDataCategory),
			SpecialOrCriminalData:                formatSpecialOrCriminalData(pa.SpecialOrCriminalData),
			LawfulBasis:                          formatLawfulBasis(pa.LawfulBasis),
			ConsentEvidenceLink:                  derefStringOrNotSpecified(pa.ConsentEvidenceLink),
			Recipients:                           derefStringOrNotSpecified(pa.Recipients),
			Location:                             derefStringOrNotSpecified(pa.Location),
			InternationalTransfers:               yesNoLabel(pa.InternationalTransfers),
			TransferSafeguards:                   formatTransferSafeguard(pa.TransferSafeguard),
			RetentionPeriod:                      derefStringOrNotSpecified(pa.RetentionPeriod),
			SecurityMeasures:                     derefStringOrNotSpecified(pa.SecurityMeasures),
			DataProtectionImpactAssessmentNeeded: formatDPIANeeded(pa.DataProtectionImpactAssessmentNeeded),
			TransferImpactAssessmentNeeded:       formatTIANeeded(pa.TransferImpactAssessmentNeeded),
			LastReviewDate:                       formatDateOrNotSpecified(pa.LastReviewDate),
			NextReviewDate:                       formatDateOrNotSpecified(pa.NextReviewDate),
			DataProtectionOfficer:                dpoName,
			ThirdParties:                         thirdPartyStr,
		})
	}

	return docgen.ProcessingActivityListData{
		Title:                     "Processing Activities",
		OrganizationName:          organization.Name,
		CreatedAt:                 time.Now(),
		TotalProcessingActivities: len(processingActivities),
		Rows:                      rows,
	}, nil
}

func derefStringOrNotSpecified(s *string) string {
	if s == nil || *s == "" {
		return "Not specified"
	}

	return *s
}

func formatDateOrNotSpecified(t *time.Time) string {
	if t == nil {
		return "Not specified"
	}

	return t.Format("January 2, 2006")
}

func yesNoLabel(b bool) string {
	if b {
		return "Yes"
	}

	return "No"
}

func formatProcessingActivityRole(role coredata.ProcessingActivityRole) string {
	switch role {
	case coredata.ProcessingActivityRoleController:
		return "Controller"
	case coredata.ProcessingActivityRoleProcessor:
		return "Processor"
	default:
		return stringOrNotSpecified(string(role))
	}
}

func formatLawfulBasis(basis coredata.ProcessingActivityLawfulBasis) string {
	switch basis {
	case coredata.ProcessingActivityLawfulBasisConsent:
		return "Consent"
	case coredata.ProcessingActivityLawfulBasisContractualNecessity:
		return "Contractual Necessity"
	case coredata.ProcessingActivityLawfulBasisLegalObligation:
		return "Legal Obligation"
	case coredata.ProcessingActivityLawfulBasisLegitimateInterest:
		return "Legitimate Interest"
	case coredata.ProcessingActivityLawfulBasisPublicTask:
		return "Public Task"
	case coredata.ProcessingActivityLawfulBasisVitalInterests:
		return "Vital Interests"
	default:
		return stringOrNotSpecified(string(basis))
	}
}

func formatSpecialOrCriminalData(data coredata.ProcessingActivitySpecialOrCriminalDatum) string {
	switch data {
	case coredata.ProcessingActivitySpecialOrCriminalDatumYes:
		return "Yes"
	case coredata.ProcessingActivitySpecialOrCriminalDatumNo:
		return "No"
	case coredata.ProcessingActivitySpecialOrCriminalDatumPossible:
		return "Possible"
	default:
		return stringOrNotSpecified(string(data))
	}
}

func formatTransferSafeguard(safeguard *coredata.ProcessingActivityTransferSafeguard) string {
	if safeguard == nil {
		return "Not specified"
	}

	switch *safeguard {
	case coredata.ProcessingActivityTransferSafeguardStandardContractualClauses:
		return "Standard Contractual Clauses"
	case coredata.ProcessingActivityTransferSafeguardBindingCorporateRules:
		return "Binding Corporate Rules"
	case coredata.ProcessingActivityTransferSafeguardAdequacyDecision:
		return "Adequacy Decision"
	case coredata.ProcessingActivityTransferSafeguardDerogations:
		return "Derogations"
	case coredata.ProcessingActivityTransferSafeguardCodesOfConduct:
		return "Codes of Conduct"
	case coredata.ProcessingActivityTransferSafeguardCertificationMechanisms:
		return "Certification Mechanisms"
	default:
		return stringOrNotSpecified(string(*safeguard))
	}
}

func formatDPIANeeded(needed coredata.ProcessingActivityDataProtectionImpactAssessment) string {
	switch needed {
	case coredata.ProcessingActivityDataProtectionImpactAssessmentNeeded:
		return "Yes"
	case coredata.ProcessingActivityDataProtectionImpactAssessmentNotNeeded:
		return "No"
	default:
		return stringOrNotSpecified(string(needed))
	}
}

func formatTIANeeded(needed coredata.ProcessingActivityTransferImpactAssessment) string {
	switch needed {
	case coredata.ProcessingActivityTransferImpactAssessmentNeeded:
		return "Yes"
	case coredata.ProcessingActivityTransferImpactAssessmentNotNeeded:
		return "No"
	default:
		return stringOrNotSpecified(string(needed))
	}
}

func formatResidualRisk(risk *coredata.DataProtectionImpactAssessmentResidualRisk) string {
	if risk == nil {
		return "Not specified"
	}

	switch *risk {
	case coredata.DataProtectionImpactAssessmentResidualRiskLow:
		return "Low"
	case coredata.DataProtectionImpactAssessmentResidualRiskMedium:
		return "Medium"
	case coredata.DataProtectionImpactAssessmentResidualRiskHigh:
		return "High"
	default:
		return stringOrNotSpecified(string(*risk))
	}
}

var processingActivityListTemplate = template.Must(
	template.New("processing_activity_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
			"printf": fmt.Sprintf,
			"add":    func(a, b int) int { return a + b },
		}).
		ParseFS(Templates, "templates/processing_activity_list.json.tmpl"),
)

func BuildProcessingActivityListDocument(data docgen.ProcessingActivityListData) (string, error) {
	var buf bytes.Buffer
	if err := processingActivityListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute processing activity list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishDataProtectionImpactAssessmentList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildDataProtectionImpactAssessmentListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildDataProtectionImpactAssessmentListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			dpia := coredata.DataProtectionImpactAssessment{}

			dpiaDocumentID, err := dpia.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if dpiaDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *dpiaDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load DPIA list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := dpia.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*dpiaDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := dpia.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Data Protection Impact Assessments",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetDataProtectionImpactAssessmentsDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		dpia := coredata.DataProtectionImpactAssessment{}

		var err error

		documentID, err = dpia.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get DPIA list document ID: %w", err)
	}

	return documentID, nil
}

func (s *GeneratedDocumentService) buildDataProtectionImpactAssessmentListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.DataProtectionImpactAssessmentListData, error) {
	assessments, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.DataProtectionImpactAssessmentOrderField]{
			Field:     coredata.DataProtectionImpactAssessmentOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DataProtectionImpactAssessmentOrderField]) ([]*coredata.DataProtectionImpactAssessment, error) {
			var batch coredata.DataProtectionImpactAssessments
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load DPIAs: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.DataProtectionImpactAssessmentListData{}, err
	}

	if len(assessments) == 0 {
		return docgen.DataProtectionImpactAssessmentListData{
			Title:                                "Data Protection Impact Assessments",
			OrganizationName:                     organization.Name,
			CreatedAt:                            time.Now(),
			TotalDataProtectionImpactAssessments: 0,
		}, nil
	}

	processingActivityIDs := make([]gid.GID, 0, len(assessments))

	processingActivityIDSet := make(map[gid.GID]struct{}, len(assessments))
	for _, a := range assessments {
		if _, ok := processingActivityIDSet[a.ProcessingActivityID]; !ok {
			processingActivityIDs = append(processingActivityIDs, a.ProcessingActivityID)
			processingActivityIDSet[a.ProcessingActivityID] = struct{}{}
		}
	}

	var processingActivities coredata.ProcessingActivities
	if err := processingActivities.LoadByIDs(ctx, conn, scope, processingActivityIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return docgen.DataProtectionImpactAssessmentListData{}, fmt.Errorf("cannot load processing activities: %w", err)
	}

	processingActivityMap := make(map[gid.GID]*coredata.ProcessingActivity, len(processingActivities))
	for _, pa := range processingActivities {
		processingActivityMap[pa.ID] = pa
	}

	rows := make([]docgen.DataProtectionImpactAssessmentListRow, 0, len(assessments))
	for _, a := range assessments {
		paName := "-"
		if pa, ok := processingActivityMap[a.ProcessingActivityID]; ok {
			paName = pa.Name
		}

		rows = append(rows, docgen.DataProtectionImpactAssessmentListRow{
			ProcessingActivityName:      paName,
			Description:                 derefStringOrNotSpecified(a.Description),
			NecessityAndProportionality: derefStringOrNotSpecified(a.NecessityAndProportionality),
			PotentialRisk:               derefStringOrNotSpecified(a.PotentialRisk),
			Mitigations:                 derefStringOrNotSpecified(a.Mitigations),
			ResidualRisk:                formatResidualRisk(a.ResidualRisk),
		})
	}

	return docgen.DataProtectionImpactAssessmentListData{
		Title:                                "Data Protection Impact Assessments",
		OrganizationName:                     organization.Name,
		CreatedAt:                            time.Now(),
		TotalDataProtectionImpactAssessments: len(assessments),
		Rows:                                 rows,
	}, nil
}

var dataProtectionImpactAssessmentListTemplate = template.Must(
	template.New("data_protection_impact_assessment_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
			"printf": fmt.Sprintf,
			"add":    func(a, b int) int { return a + b },
		}).
		ParseFS(Templates, "templates/data_protection_impact_assessment_list.json.tmpl"),
)

func BuildDataProtectionImpactAssessmentListDocument(data docgen.DataProtectionImpactAssessmentListData) (string, error) {
	var buf bytes.Buffer
	if err := dataProtectionImpactAssessmentListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute DPIA list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishTransferImpactAssessmentList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildTransferImpactAssessmentListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildTransferImpactAssessmentListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			tia := coredata.TransferImpactAssessment{}

			tiaDocumentID, err := tia.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if tiaDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *tiaDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load TIA list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := tia.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*tiaDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := tia.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Transfer Impact Assessments",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetTransferImpactAssessmentsDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		tia := coredata.TransferImpactAssessment{}

		var err error

		documentID, err = tia.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get TIA list document ID: %w", err)
	}

	return documentID, nil
}

func (s *GeneratedDocumentService) buildTransferImpactAssessmentListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.TransferImpactAssessmentListData, error) {
	assessments, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.TransferImpactAssessmentOrderField]{
			Field:     coredata.TransferImpactAssessmentOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.TransferImpactAssessmentOrderField]) ([]*coredata.TransferImpactAssessment, error) {
			var batch coredata.TransferImpactAssessments
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load TIAs: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.TransferImpactAssessmentListData{}, err
	}

	if len(assessments) == 0 {
		return docgen.TransferImpactAssessmentListData{
			Title:                          "Transfer Impact Assessments",
			OrganizationName:               organization.Name,
			CreatedAt:                      time.Now(),
			TotalTransferImpactAssessments: 0,
		}, nil
	}

	processingActivityIDs := make([]gid.GID, 0, len(assessments))

	processingActivityIDSet := make(map[gid.GID]struct{}, len(assessments))
	for _, a := range assessments {
		if _, ok := processingActivityIDSet[a.ProcessingActivityID]; !ok {
			processingActivityIDs = append(processingActivityIDs, a.ProcessingActivityID)
			processingActivityIDSet[a.ProcessingActivityID] = struct{}{}
		}
	}

	var processingActivities coredata.ProcessingActivities
	if err := processingActivities.LoadByIDs(ctx, conn, scope, processingActivityIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return docgen.TransferImpactAssessmentListData{}, fmt.Errorf("cannot load processing activities: %w", err)
	}

	processingActivityMap := make(map[gid.GID]*coredata.ProcessingActivity, len(processingActivities))
	for _, pa := range processingActivities {
		processingActivityMap[pa.ID] = pa
	}

	rows := make([]docgen.TransferImpactAssessmentListRow, 0, len(assessments))
	for _, a := range assessments {
		paName := "-"
		if pa, ok := processingActivityMap[a.ProcessingActivityID]; ok {
			paName = pa.Name
		}

		rows = append(rows, docgen.TransferImpactAssessmentListRow{
			ProcessingActivityName: paName,
			DataSubjects:           derefStringOrNotSpecified(a.DataSubjects),
			Transfer:               derefStringOrNotSpecified(a.Transfer),
			LegalMechanism:         derefStringOrNotSpecified(a.LegalMechanism),
			LocalLawRisk:           derefStringOrNotSpecified(a.LocalLawRisk),
			SupplementaryMeasures:  derefStringOrNotSpecified(a.SupplementaryMeasures),
		})
	}

	return docgen.TransferImpactAssessmentListData{
		Title:                          "Transfer Impact Assessments",
		OrganizationName:               organization.Name,
		CreatedAt:                      time.Now(),
		TotalTransferImpactAssessments: len(assessments),
		Rows:                           rows,
	}, nil
}

var transferImpactAssessmentListTemplate = template.Must(
	template.New("transfer_impact_assessment_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
			"printf": fmt.Sprintf,
			"add":    func(a, b int) int { return a + b },
		}).
		ParseFS(Templates, "templates/transfer_impact_assessment_list.json.tmpl"),
)

func BuildTransferImpactAssessmentListDocument(data docgen.TransferImpactAssessmentListData) (string, error) {
	var buf bytes.Buffer
	if err := transferImpactAssessmentListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute TIA list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishThirdPartyList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	// Phase 1: collect data and render the prosemirror document outside any
	// write transaction. Both the bulk reads of thirdParties + sub-entities and the
	// JSON template rendering are slow enough that holding write locks across
	// them would needlessly block other writers.
	var documentData docgen.ThirdPartyListData

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		var err error

		documentData, err = s.buildThirdPartyListDocumentData(ctx, scope, conn, organization)
		if err != nil {
			return fmt.Errorf("cannot build document data: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	prosemirrorJSON, err := BuildThirdPartyListDocument(documentData)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot build prosemirror document: %w", err)
	}

	// Phase 2: persist the document and version in a write transaction.
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			thirdParty := coredata.ThirdParty{}

			thirdPartyDocumentID, err := thirdParty.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if thirdPartyDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *thirdPartyDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load thirdParty list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := thirdParty.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*thirdPartyDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := thirdParty.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "ThirdParties",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetThirdPartiesDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		thirdParty := coredata.ThirdParty{}

		var err error

		documentID, err = thirdParty.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get thirdParty list document ID: %w", err)
	}

	return documentID, nil
}

func (s *GeneratedDocumentService) buildThirdPartyListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.ThirdPartyListData, error) {
	firstLevel := 1

	thirdParties, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ThirdPartyOrderField]{
			Field:     coredata.ThirdPartyOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ThirdPartyOrderField]) ([]*coredata.ThirdParty, error) {
			var batch coredata.ThirdParties
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor, coredata.NewThirdPartyFilter(nil, &firstLevel, nil, nil, nil)); err != nil {
				return nil, fmt.Errorf("cannot load thirdParties: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.ThirdPartyListData{}, err
	}

	if len(thirdParties) == 0 {
		return docgen.ThirdPartyListData{
			Title:             "ThirdParties",
			OrganizationName:  organization.Name,
			CreatedAt:         time.Now(),
			TotalThirdParties: 0,
		}, nil
	}

	ownerIDSet := make(map[gid.GID]struct{})
	ownerIDs := make([]gid.GID, 0)

	for _, v := range thirdParties {
		if v.BusinessOwnerID != nil {
			if _, ok := ownerIDSet[*v.BusinessOwnerID]; !ok {
				ownerIDs = append(ownerIDs, *v.BusinessOwnerID)
				ownerIDSet[*v.BusinessOwnerID] = struct{}{}
			}
		}

		if v.SecurityOwnerID != nil {
			if _, ok := ownerIDSet[*v.SecurityOwnerID]; !ok {
				ownerIDs = append(ownerIDs, *v.SecurityOwnerID)
				ownerIDSet[*v.SecurityOwnerID] = struct{}{}
			}
		}
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile)

	if len(ownerIDs) > 0 {
		var profiles coredata.MembershipProfiles
		if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load owner profiles: %w", err)
		}

		for _, p := range profiles {
			profileMap[p.ID] = p
		}
	}

	thirdPartyIDs := make([]gid.GID, len(thirdParties))
	for i, v := range thirdParties {
		thirdPartyIDs[i] = v.ID
	}

	var allServices coredata.ThirdPartyServices
	if err := allServices.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty services: %w", err)
	}

	servicesByThirdParty := make(map[gid.GID]coredata.ThirdPartyServices, len(thirdParties))
	for _, vs := range allServices {
		servicesByThirdParty[vs.ThirdPartyID] = append(servicesByThirdParty[vs.ThirdPartyID], vs)
	}

	var allContacts coredata.ThirdPartyContacts
	if err := allContacts.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty contacts: %w", err)
	}

	contactsByThirdParty := make(map[gid.GID]coredata.ThirdPartyContacts, len(thirdParties))
	for _, c := range allContacts {
		contactsByThirdParty[c.ThirdPartyID] = append(contactsByThirdParty[c.ThirdPartyID], c)
	}

	var allAssessments coredata.ThirdPartyRiskAssessments
	if err := allAssessments.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty risk assessments: %w", err)
	}

	assessmentsByThirdParty := make(map[gid.GID]coredata.ThirdPartyRiskAssessments, len(thirdParties))
	for _, ra := range allAssessments {
		assessmentsByThirdParty[ra.ThirdPartyID] = append(assessmentsByThirdParty[ra.ThirdPartyID], ra)
	}

	var allReports coredata.ThirdPartyComplianceReports
	if err := allReports.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty compliance reports: %w", err)
	}

	reportsByThirdParty := make(map[gid.GID]coredata.ThirdPartyComplianceReports, len(thirdParties))
	for _, r := range allReports {
		reportsByThirdParty[r.ThirdPartyID] = append(reportsByThirdParty[r.ThirdPartyID], r)
	}

	var allBAAs coredata.ThirdPartyBusinessAssociateAgreements
	if err := allBAAs.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty business associate agreements: %w", err)
	}

	baaByThirdParty := make(map[gid.GID]*coredata.ThirdPartyBusinessAssociateAgreement, len(allBAAs))
	for _, b := range allBAAs {
		baaByThirdParty[b.ThirdPartyID] = b
	}

	var allDPAs coredata.ThirdPartyDataPrivacyAgreements
	if err := allDPAs.LoadByThirdPartyIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return docgen.ThirdPartyListData{}, fmt.Errorf("cannot load thirdParty data privacy agreements: %w", err)
	}

	dpaByThirdParty := make(map[gid.GID]*coredata.ThirdPartyDataPrivacyAgreement, len(allDPAs))
	for _, d := range allDPAs {
		dpaByThirdParty[d.ThirdPartyID] = d
	}

	rows := make([]docgen.ThirdPartyListRow, 0, len(thirdParties))
	for _, v := range thirdParties {
		row := docgen.ThirdPartyListRow{
			Name:                          v.Name,
			LegalName:                     derefStringOrNotSpecified(v.LegalName),
			Description:                   derefStringOrNotSpecified(v.Description),
			Category:                      formatThirdPartyCategory(v.Category),
			HeadquarterAddress:            derefStringOrNotSpecified(v.HeadquarterAddress),
			WebsiteURL:                    derefStringOrNotSpecified(v.WebsiteURL),
			PrivacyPolicyURL:              derefStringOrNotSpecified(v.PrivacyPolicyURL),
			ServiceLevelAgreementURL:      derefStringOrNotSpecified(v.ServiceLevelAgreementURL),
			DataProcessingAgreementURL:    derefStringOrNotSpecified(v.DataProcessingAgreementURL),
			BusinessAssociateAgreementURL: derefStringOrNotSpecified(v.BusinessAssociateAgreementURL),
			SubprocessorsListURL:          derefStringOrNotSpecified(v.SubprocessorsListURL),
			StatusPageURL:                 derefStringOrNotSpecified(v.StatusPageURL),
			TermsOfServiceURL:             derefStringOrNotSpecified(v.TermsOfServiceURL),
			SecurityPageURL:               derefStringOrNotSpecified(v.SecurityPageURL),
			TrustPageURL:                  derefStringOrNotSpecified(v.TrustPageURL),
			Certifications:                joinOrNotSpecified(v.Certifications),
			Countries:                     formatCountries(v.Countries),
			BusinessOwner:                 lookupProfileName(profileMap, v.BusinessOwnerID),
			SecurityOwner:                 lookupProfileName(profileMap, v.SecurityOwnerID),
		}

		for _, vs := range servicesByThirdParty[v.ID] {
			row.Services = append(row.Services, docgen.ThirdPartyListService{
				Name:        vs.Name,
				Description: derefStringOrNotSpecified(vs.Description),
			})
		}

		for _, c := range contactsByThirdParty[v.ID] {
			email := ""
			if c.Email != nil {
				email = c.Email.String()
			}

			row.Contacts = append(row.Contacts, docgen.ThirdPartyListContact{
				FullName: derefStringOrNotSpecified(c.FullName),
				Email:    stringOrNotSpecified(email),
				Phone:    derefStringOrNotSpecified(c.Phone),
				Role:     derefStringOrNotSpecified(c.Role),
			})
		}

		for _, ra := range assessmentsByThirdParty[v.ID] {
			row.RiskAssessments = append(row.RiskAssessments, docgen.ThirdPartyListRiskAssessment{
				AssessedAt:      ra.CreatedAt.Format("2006-01-02"),
				ExpiresAt:       ra.ExpiresAt.Format("2006-01-02"),
				DataSensitivity: formatDataSensitivity(ra.DataSensitivity),
				BusinessImpact:  formatBusinessImpact(ra.BusinessImpact),
				Notes:           derefStringOrNotSpecified(ra.Notes),
			})
		}

		for _, r := range reportsByThirdParty[v.ID] {
			row.ComplianceReports = append(row.ComplianceReports, docgen.ThirdPartyListComplianceReport{
				ReportName: r.ReportName,
				ReportDate: r.ReportDate.Format("2006-01-02"),
				ValidUntil: formatTimeOrNotSpecified(r.ValidUntil),
			})
		}

		if baa := baaByThirdParty[v.ID]; baa != nil {
			row.BusinessAssociateAgreement = &docgen.ThirdPartyListAgreement{
				ValidFrom:  formatTimeOrNotSpecified(baa.ValidFrom),
				ValidUntil: formatTimeOrNotSpecified(baa.ValidUntil),
			}
		}

		if dpa := dpaByThirdParty[v.ID]; dpa != nil {
			row.DataPrivacyAgreement = &docgen.ThirdPartyListAgreement{
				ValidFrom:  formatTimeOrNotSpecified(dpa.ValidFrom),
				ValidUntil: formatTimeOrNotSpecified(dpa.ValidUntil),
			}
		}

		rows = append(rows, row)
	}

	return docgen.ThirdPartyListData{
		Title:             "ThirdParties",
		OrganizationName:  organization.Name,
		CreatedAt:         time.Now(),
		TotalThirdParties: len(thirdParties),
		Rows:              rows,
	}, nil
}

func stringOrNotSpecified(s string) string {
	if s == "" {
		return "Not specified"
	}

	return s
}

func formatTimeOrNotSpecified(t *time.Time) string {
	if t == nil {
		return "Not specified"
	}

	return t.Format("2006-01-02")
}

func joinOrNotSpecified(items []string) string {
	if len(items) == 0 {
		return "Not specified"
	}

	return strings.Join(items, ", ")
}

func formatCountries(c coredata.CountryCodes) string {
	if len(c) == 0 {
		return "Not specified"
	}

	parts := make([]string, len(c))
	for i, cc := range c {
		parts[i] = string(cc)
	}

	return strings.Join(parts, ", ")
}

func lookupProfileName(profiles map[gid.GID]*coredata.MembershipProfile, id *gid.GID) string {
	if id == nil {
		return "Not assigned"
	}

	if p, ok := profiles[*id]; ok && p.FullName != "" {
		return p.FullName
	}

	return "Not assigned"
}

func formatDataSensitivity(s coredata.DataSensitivity) string {
	switch s {
	case coredata.DataSensitivityNone:
		return "None"
	case coredata.DataSensitivityLow:
		return "Low"
	case coredata.DataSensitivityMedium:
		return "Medium"
	case coredata.DataSensitivityHigh:
		return "High"
	case coredata.DataSensitivityCritical:
		return "Critical"
	default:
		return stringOrNotSpecified(string(s))
	}
}

func formatBusinessImpact(b coredata.BusinessImpact) string {
	switch b {
	case coredata.BusinessImpactLow:
		return "Low"
	case coredata.BusinessImpactMedium:
		return "Medium"
	case coredata.BusinessImpactHigh:
		return "High"
	case coredata.BusinessImpactCritical:
		return "Critical"
	default:
		return stringOrNotSpecified(string(b))
	}
}

func formatThirdPartyCategory(c coredata.ThirdPartyCategory) string {
	switch c {
	case coredata.ThirdPartyCategoryAnalytics:
		return "Analytics"
	case coredata.ThirdPartyCategoryCloudMonitoring:
		return "Cloud Monitoring"
	case coredata.ThirdPartyCategoryCloudProvider:
		return "Cloud Provider"
	case coredata.ThirdPartyCategoryCollaboration:
		return "Collaboration"
	case coredata.ThirdPartyCategoryCustomerSupport:
		return "Customer Support"
	case coredata.ThirdPartyCategoryDataStorageAndProcessing:
		return "Data Storage and Processing"
	case coredata.ThirdPartyCategoryDocumentManagement:
		return "Document Management"
	case coredata.ThirdPartyCategoryEmployeeManagement:
		return "Employee Management"
	case coredata.ThirdPartyCategoryEngineering:
		return "Engineering"
	case coredata.ThirdPartyCategoryFinance:
		return "Finance"
	case coredata.ThirdPartyCategoryIdentityProvider:
		return "Identity Provider"
	case coredata.ThirdPartyCategoryIT:
		return "IT"
	case coredata.ThirdPartyCategoryMarketing:
		return "Marketing"
	case coredata.ThirdPartyCategoryOfficeOperations:
		return "Office Operations"
	case coredata.ThirdPartyCategoryOther:
		return "Other"
	case coredata.ThirdPartyCategoryPasswordManagement:
		return "Password Management"
	case coredata.ThirdPartyCategoryProductAndDesign:
		return "Product and Design"
	case coredata.ThirdPartyCategoryProfessionalServices:
		return "Professional Services"
	case coredata.ThirdPartyCategoryRecruiting:
		return "Recruiting"
	case coredata.ThirdPartyCategorySales:
		return "Sales"
	case coredata.ThirdPartyCategorySecurity:
		return "Security"
	case coredata.ThirdPartyCategoryVersionControl:
		return "Version Control"
	default:
		return stringOrNotSpecified(string(c))
	}
}

var thirdPartyListTemplate = template.Must(
	template.New("third_party_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
			"printf": fmt.Sprintf,
			"add":    func(a, b int) int { return a + b },
			"default": func(fallback string, v any) string {
				switch s := v.(type) {
				case nil:
					return fallback
				case string:
					if s == "" {
						return fallback
					}

					return s
				default:
					str := fmt.Sprint(v)
					if str == "" {
						return fallback
					}

					return str
				}
			},
		}).
		ParseFS(Templates, "templates/third_party_list.json.tmpl"),
)

func BuildThirdPartyListDocument(data docgen.ThirdPartyListData) (string, error) {
	var buf bytes.Buffer
	if err := thirdPartyListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute thirdParty list template: %w", err)
	}

	return buf.String(), nil
}

var riskListTemplate = template.Must(
	template.New("risk_list.json.tmpl").
		Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}

				return string(b), nil
			},
			"printf": fmt.Sprintf,
			"add":    func(a, b int) int { return a + b },
		}).
		ParseFS(Templates, "templates/risk_list.json.tmpl"),
)

func BuildRiskListDocument(data docgen.RiskListData) (string, error) {
	var buf bytes.Buffer
	if err := riskListTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute risk list template: %w", err)
	}

	return buf.String(), nil
}

func (s *GeneratedDocumentService) PublishRiskList(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var (
		document        *coredata.Document
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			documentData, err := s.buildRiskListDocumentData(ctx, scope, tx, organization)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildRiskListDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			now := time.Now()

			risk := coredata.Risk{}

			riskDocumentID, err := risk.GetGeneratedDocumentID(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot query generated documents: %w", err)
			}

			var existingDoc *coredata.Document

			if riskDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *riskDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load risk list document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					if err := risk.ClearGeneratedDocumentID(ctx, tx, []gid.GID{*riskDocumentID}); err != nil {
						return fmt.Errorf("cannot clear document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				if err := risk.UpsertGeneratedDocumentID(ctx, tx, organizationID, scope.GetTenantID(), documentID); err != nil {
					return fmt.Errorf("cannot upsert generated documents: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion = &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          "Risks",
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationConfidential,
				DocumentType:   coredata.DocumentTypeRegister,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, approverIDs, minor, now)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *GeneratedDocumentService) GetRisksDocumentID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (*gid.GID, error) {
	var riskDocumentID *gid.GID

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		risk := coredata.Risk{}

		var err error

		riskDocumentID, err = risk.GetGeneratedDocumentID(ctx, conn, organizationID)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get risk list document ID: %w", err)
	}

	return riskDocumentID, nil
}

func (s *GeneratedDocumentService) buildRiskListDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
) (docgen.RiskListData, error) {
	risks, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskOrderField]{
			Field:     coredata.RiskOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskOrderField]) ([]*coredata.Risk, error) {
			var batch coredata.Risks
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor, coredata.NewRiskFilter(nil)); err != nil {
				return nil, fmt.Errorf("cannot load risks: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return docgen.RiskListData{}, err
	}

	if len(risks) == 0 {
		return docgen.RiskListData{
			Title:            "Risks",
			OrganizationName: organization.Name,
			CreatedAt:        time.Now(),
			TotalRisks:       0,
		}, nil
	}

	ownerIDs := make([]gid.GID, 0, len(risks))
	ownerIDSet := make(map[gid.GID]struct{})

	for _, r := range risks {
		if r.OwnerID != nil {
			if _, ok := ownerIDSet[*r.OwnerID]; !ok {
				ownerIDs = append(ownerIDs, *r.OwnerID)
				ownerIDSet[*r.OwnerID] = struct{}{}
			}
		}
	}

	profileMap := make(map[gid.GID]*coredata.MembershipProfile)

	if len(ownerIDs) > 0 {
		var profiles coredata.MembershipProfiles
		if err := profiles.LoadByIDs(ctx, conn, scope, ownerIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return docgen.RiskListData{}, fmt.Errorf("cannot load profiles: %w", err)
		}

		for _, p := range profiles {
			profileMap[p.ID] = p
		}
	}

	rows := make([]docgen.RiskListRow, 0, len(risks))
	for _, r := range risks {
		rows = append(rows, docgen.RiskListRow{
			Name:                    r.Name,
			Description:             derefStringOrNotSpecified(r.Description),
			Category:                stringOrNotSpecified(r.Category),
			Treatment:               formatRiskTreatment(r.Treatment),
			Owner:                   lookupProfileName(profileMap, r.OwnerID),
			InherentLikelihood:      r.InherentLikelihood,
			InherentLikelihoodLabel: riskLikelihoodLabel(r.InherentLikelihood),
			InherentImpact:          r.InherentImpact,
			InherentImpactLabel:     riskImpactLabel(r.InherentImpact),
			InherentRiskScore:       r.InherentRiskScore,
			InherentSeverity:        riskSeverityLabel(r.InherentRiskScore),
			ResidualLikelihood:      r.ResidualLikelihood,
			ResidualLikelihoodLabel: riskLikelihoodLabel(r.ResidualLikelihood),
			ResidualImpact:          r.ResidualImpact,
			ResidualImpactLabel:     riskImpactLabel(r.ResidualImpact),
			ResidualRiskScore:       r.ResidualRiskScore,
			ResidualSeverity:        riskSeverityLabel(r.ResidualRiskScore),
			Note:                    stringOrNotSpecified(r.Note),
		})
	}

	return docgen.RiskListData{
		Title:            "Risks",
		OrganizationName: organization.Name,
		CreatedAt:        time.Now(),
		TotalRisks:       len(risks),
		Rows:             rows,
	}, nil
}

func riskLikelihoodLabel(v int) string {
	switch v {
	case 1:
		return "Improbable"
	case 2:
		return "Remote"
	case 3:
		return "Occasional"
	case 4:
		return "Probable"
	case 5:
		return "Frequent"
	default:
		return "Unknown"
	}
}

func riskImpactLabel(v int) string {
	switch v {
	case 1:
		return "Negligible"
	case 2:
		return "Low"
	case 3:
		return "Moderate"
	case 4:
		return "Significant"
	case 5:
		return "Catastrophic"
	default:
		return "Unknown"
	}
}

func riskSeverityLabel(score int) string {
	switch {
	case score >= 15:
		return "Critical"
	case score >= 5:
		return "High"
	default:
		return "Low"
	}
}

func formatRiskTreatment(t coredata.RiskTreatment) string {
	switch t {
	case coredata.RiskTreatmentMitigated:
		return "Mitigated"
	case coredata.RiskTreatmentAccepted:
		return "Accepted"
	case coredata.RiskTreatmentAvoided:
		return "Avoided"
	case coredata.RiskTreatmentTransferred:
		return "Transferred"
	default:
		return stringOrNotSpecified(string(t))
	}
}

// publishOrRequestApproval finalises a freshly built generated document
// version. The version's Major, Minor, Status and PublishedAt fields are
// computed here based on the document's current published state, the minor
// flag, and whether approvers were provided. When minor is true the version
// is always published as a minor version and approvers are ignored. When
// minor is false a non-empty approverIDs triggers an approval request at
// (currentMajor+1).0; otherwise the version is published at (currentMajor+1).0.
func (s *GeneratedDocumentService) publishOrRequestApproval(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	document *coredata.Document,
	version *coredata.DocumentVersion,
	organizationID gid.GID,
	approverIDs []gid.GID,
	minor bool,
	now time.Time,
) error {
	previousVersion := &coredata.DocumentVersion{}

	isFirstVersion := false

	err := previousVersion.LoadLatestVersion(ctx, tx, scope, document.ID)
	switch {
	case err == nil:
		version.Title = previousVersion.Title
		version.Classification = previousVersion.Classification
		version.DocumentType = previousVersion.DocumentType
	case errors.Is(err, coredata.ErrResourceNotFound):
		// First publish: keep the caller-provided defaults.
		isFirstVersion = true
	default:
		return fmt.Errorf("cannot load previous document version: %w", err)
	}

	if minor {
		if document.CurrentPublishedMajor != nil && document.CurrentPublishedMinor != nil {
			version.Major = *document.CurrentPublishedMajor
			version.Minor = *document.CurrentPublishedMinor + 1
		} else {
			version.Major = 0
			version.Minor = 1
		}

		version.Status = coredata.DocumentVersionStatusPublished
		version.PublishedAt = &now
		approverIDs = nil
	} else {
		if document.CurrentPublishedMajor != nil {
			version.Major = *document.CurrentPublishedMajor + 1
		} else {
			version.Major = 1
		}

		version.Minor = 0
		if len(approverIDs) > 0 {
			version.Status = coredata.DocumentVersionStatusDraft
			version.PublishedAt = nil
		} else {
			version.Status = coredata.DocumentVersionStatusPublished
			version.PublishedAt = &now
		}
	}

	if err := version.Insert(ctx, tx, scope); err != nil {
		if errors.Is(err, coredata.ErrResourceAlreadyExists) {
			switch previousVersion.Status {
			case coredata.DocumentVersionStatusDraft:
				return fmt.Errorf("a draft version exists, publish or delete it before publishing a new one: %w", err)
			case coredata.DocumentVersionStatusPendingApproval:
				return fmt.Errorf("a version is pending approval, approve or reject it before publishing a new one: %w", err)
			default:
				return fmt.Errorf("a version already exists at this number: %w", err)
			}
		}

		return fmt.Errorf("cannot insert document version: %w", err)
	}

	if len(approverIDs) > 0 {
		profiles := &coredata.MembershipProfiles{}
		if err := profiles.LoadByIDs(ctx, tx, scope, approverIDs); err != nil {
			return fmt.Errorf("cannot load approver profiles: %w", err)
		}

		defaultApprovers := &coredata.DocumentDefaultApprovers{}
		if err := defaultApprovers.MergeByDocumentID(ctx, tx, scope, document.ID, organizationID, approverIDs); err != nil {
			return fmt.Errorf("cannot save default approvers: %w", err)
		}

		quorum, err := s.svc.DocumentApprovals.RequestApprovalInTx(ctx, scope, tx, document, version, approverIDs, nil)
		if err != nil {
			return fmt.Errorf("cannot request approval: %w", err)
		}

		if isFirstVersion {
			if err := s.svc.Documents.emitDocumentEventInTx(ctx, scope, tx, document.ID, coredata.WebhookEventTypeDocumentCreated, nil, nil, nil); err != nil {
				return fmt.Errorf("cannot emit document created webhook: %w", err)
			}
		}

		if err := s.svc.Documents.emitDocumentEventInTx(
			ctx,
			scope,
			tx,
			version.DocumentID,
			coredata.WebhookEventTypeDocumentVersionApprovalQuorumRequested,
			version,
			nil,
			&quorum.ID,
		); err != nil {
			return fmt.Errorf("cannot emit approval quorum requested webhook: %w", err)
		}

		return nil
	}

	document.CurrentPublishedMajor = &version.Major
	document.CurrentPublishedMinor = &version.Minor
	document.UpdatedAt = now

	if err := document.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update document: %w", err)
	}

	if !minor {
		if err := s.svc.Documents.cancelPreviousMajorSignatureRequestsInTx(ctx, scope, tx, document.ID, version.Major); err != nil {
			return fmt.Errorf("cannot cancel signature requests from previous major versions: %w", err)
		}
	}

	if isFirstVersion {
		if err := s.svc.Documents.emitDocumentEventInTx(ctx, scope, tx, document.ID, coredata.WebhookEventTypeDocumentCreated, nil, nil, nil); err != nil {
			return fmt.Errorf("cannot emit document created webhook: %w", err)
		}
	}

	if err := s.svc.Documents.emitDocumentEventInTx(
		ctx,
		scope,
		tx,
		version.DocumentID,
		coredata.WebhookEventTypeDocumentVersionPublished,
		version,
		nil,
		nil,
	); err != nil {
		return fmt.Errorf("cannot emit document version published webhook: %w", err)
	}

	return nil
}
