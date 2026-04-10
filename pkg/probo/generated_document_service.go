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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

type GeneratedDocumentService struct {
	svc *TenantService
}

func (s *GeneratedDocumentService) PublishStatementOfApplicability(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
	approverIDs []gid.GID,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var document *coredata.Document
	var documentVersion *coredata.DocumentVersion

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			documentData, err := s.buildStatementOfApplicabilityDocumentData(ctx, tx, statementOfApplicabilityID)
			if err != nil {
				return fmt.Errorf("cannot build document data: %w", err)
			}

			prosemirrorJSON, err := BuildStatementOfApplicabilityProseMirrorDocument(documentData)
			if err != nil {
				return fmt.Errorf("cannot build prosemirror document: %w", err)
			}

			soa := &coredata.StatementOfApplicability{}
			if err := soa.LoadByID(ctx, tx, s.svc.scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement of applicability: %w", err)
			}

			now := time.Now()
			landscape := coredata.DocumentVersionOrientationLandscape

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
					ContentSource:         coredata.DocumentContentSourceGenerated,
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

				latestVersion := &coredata.DocumentVersion{}
				if err := latestVersion.LoadLatestVersion(ctx, tx, s.svc.scope, document.ID); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						return fmt.Errorf("cannot load latest version: %w", err)
					}
				} else if latestVersion.Status == coredata.DocumentVersionStatusDraft {
					if err := latestVersion.Delete(ctx, tx, s.svc.scope); err != nil {
						return fmt.Errorf("cannot delete existing draft version: %w", err)
					}
				}
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
				Orientation:    landscape,
				PublishedAt:    publishedAt,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := documentVersion.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert document version: %w", err)
			}

			if hasApprovers {
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
	statementOfApplicabilityID gid.GID,
) (docgen.StatementOfApplicabilityData, error) {
	statementOfApplicability := &coredata.StatementOfApplicability{}
	if err := statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, statementOfApplicabilityID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load statement of applicability: %w", err)
	}

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, conn, s.svc.scope, statementOfApplicability.OrganizationID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load organization: %w", err)
	}

	var applicabilityStatements coredata.ApplicabilityStatements
	if err := applicabilityStatements.LoadAllByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicabilityID); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load applicability statements: %w", err)
	}

	if len(applicabilityStatements) == 0 {
		return docgen.StatementOfApplicabilityData{
			Title:            statementOfApplicability.Name,
			OrganizationName: organization.Name,
			CreatedAt:        statementOfApplicability.CreatedAt,
			TotalControls:    0,
			FrameworkGroups:  []docgen.FrameworkControlGroup{},
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

	obligationCounts, err := coredata.CountObligationsByControlIDs(ctx, conn, s.svc.scope, controlIDs)
	if err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot count obligations: %w", err)
	}

	type obligationKey struct {
		controlID gid.GID
		oblType   coredata.ObligationType
	}
	oblCountMap := make(map[obligationKey]int, len(obligationCounts))
	for _, oc := range obligationCounts {
		oblCountMap[obligationKey{oc.ControlID, oc.ObligationType}] = oc.Count
	}

	var controlsWithRisk coredata.ControlsWithRisk
	if err := controlsWithRisk.LoadByControlIDs(ctx, conn, s.svc.scope, controlIDs); err != nil {
		return docgen.StatementOfApplicabilityData{}, fmt.Errorf("cannot load controls with risks: %w", err)
	}

	riskSet := make(map[gid.GID]struct{}, len(controlsWithRisk))
	for _, cwr := range controlsWithRisk {
		riskSet[cwr.ControlID] = struct{}{}
	}

	frameworkControlsMap := make(map[string][]docgen.ControlData)
	frameworkOrder := []string{}

	for _, stmt := range applicabilityStatements {
		control := controlMap[stmt.ControlID]
		framework := frameworkMap[control.FrameworkID]

		legalCount := oblCountMap[obligationKey{stmt.ControlID, coredata.ObligationTypeLegal}]
		contractualCount := oblCountMap[obligationKey{stmt.ControlID, coredata.ObligationTypeContractual}]
		_, hasRisk := riskSet[stmt.ControlID]

		if _, exists := frameworkControlsMap[framework.Name]; !exists {
			frameworkOrder = append(frameworkOrder, framework.Name)
			frameworkControlsMap[framework.Name] = []docgen.ControlData{}
		}

		var (
			regulatory     *bool
			contractual    *bool
			bestPractice   *bool
			riskAssessment *bool
		)

		if stmt.Applicability {
			regulatory = new(legalCount > 0)
			contractual = new(contractualCount > 0)
			riskAssessment = new(hasRisk)
			bestPractice = &control.BestPractice
		}

		applicability := stmt.Applicability

		implemented := control.Implemented.String()
		frameworkControlsMap[framework.Name] = append(
			frameworkControlsMap[framework.Name],
			docgen.ControlData{
				FrameworkName: framework.Name,
				SectionTitle:  control.SectionTitle,
				Name:          control.Name,
				Applicability: &applicability,
				Justification: stmt.Justification,
				BestPractice:  bestPractice,
				Implemented:   &implemented,
				NotImplementedJustification: func() *string {
					if control.Implemented == coredata.ControlImplementationStateImplemented {
						return nil
					}
					return control.NotImplementedJustification
				}(),
				Regulatory:     regulatory,
				Contractual:    contractual,
				RiskAssessment: riskAssessment,
			},
		)
	}

	frameworkGroups := make([]docgen.FrameworkControlGroup, len(frameworkOrder))
	for i, frameworkName := range frameworkOrder {
		frameworkGroups[i] = docgen.FrameworkControlGroup{
			FrameworkName: frameworkName,
			Controls:      frameworkControlsMap[frameworkName],
		}
	}

	horizontalLogoBase64 := ""
	if organization.HorizontalLogoFileID != nil {
		fileRecord := &coredata.File{}
		fileErr := fileRecord.LoadByID(ctx, conn, s.svc.scope, *organization.HorizontalLogoFileID)
		if fileErr == nil {
			base64Data, mimeType, logoErr := s.svc.fileManager.GetFileBase64(ctx, fileRecord)
			if logoErr == nil {
				horizontalLogoBase64 = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
			}
		}
	}

	return docgen.StatementOfApplicabilityData{
		Title:                       statementOfApplicability.Name,
		OrganizationName:            organization.Name,
		CreatedAt:                   statementOfApplicability.CreatedAt,
		TotalControls:               len(applicabilityStatements),
		FrameworkGroups:             frameworkGroups,
		CompanyHorizontalLogoBase64: horizontalLogoBase64,
		Version:                     0,
		PublishedAt:                 time.Now(),
		Approver:                    "",
	}, nil
}

func BuildStatementOfApplicabilityProseMirrorDocument(data docgen.StatementOfApplicabilityData) (string, error) {
	content := []prosemirror.Node{
		pmHeading(1, "1. Purpose"),
		pmParagraph(
			"This document provides a comprehensive overview of the statement of applicability for controls within the organization. " +
				"It serves as a record of which controls are applicable or not applicable to the organization, along with their " +
				"relationships to regulatory requirements, contractual obligations, risk assessments, and best practices.",
		),
	}

	// Column widths in pixels (total ~1100 for landscape A4).
	// Framework 120, Control 250, Applicability 70, Justification 130,
	// Implemented 70, Not-impl justification 110, Reg 60, Contract 60, BP 60, Risk 60.
	colwidths := [][]int{
		{120}, {250}, {70}, {130}, {70}, {110}, {60}, {60}, {60}, {60},
	}

	content = append(content,
		prosemirror.Node{Type: prosemirror.NodeHorizontalRule},
		pmHeading(1, "2. Controls"),
	)

	headerRow1 := prosemirror.Node{
		Type: prosemirror.NodeTableRow,
		Content: []prosemirror.Node{
			pmTableHeaderCellSpanW("Framework", 1, 2, colwidths[0]),
			pmTableHeaderCellSpanW("Control", 1, 2, colwidths[1]),
			pmTableHeaderCellSpanW("Applicability", 1, 2, colwidths[2]),
			pmTableHeaderCellSpanW("Justification for non-applicability", 1, 2, colwidths[3]),
			pmTableHeaderCellSpanW("Implemented", 1, 2, colwidths[4]),
			pmTableHeaderCellSpanW("Justification for non-implementation", 1, 2, colwidths[5]),
			pmTableHeaderCellSpanW("Justification for inclusion", 4, 1, nil),
		},
	}
	headerRow2 := prosemirror.Node{
		Type: prosemirror.NodeTableRow,
		Content: []prosemirror.Node{
			pmTableHeaderCellW("Regulatory", colwidths[6]),
			pmTableHeaderCellW("Contractual", colwidths[7]),
			pmTableHeaderCellW("Best Practice", colwidths[8]),
			pmTableHeaderCellW("Risk Assessment", colwidths[9]),
		},
	}

	rows := []prosemirror.Node{headerRow1, headerRow2}

	for _, group := range data.FrameworkGroups {
		for _, ctrl := range group.Controls {
			justification := "-"
			if ctrl.Applicability != nil && !*ctrl.Applicability && ctrl.Justification != nil {
				justification = *ctrl.Justification
			}

			implemented := "-"
			if ctrl.Implemented != nil {
				if ctrl.Applicability != nil && !*ctrl.Applicability {
					implemented = "-"
				} else if *ctrl.Implemented == "IMPLEMENTED" {
					implemented = "Yes"
				} else {
					implemented = "No"
				}
			}

			notImplJustification := "-"
			if ctrl.Implemented != nil && *ctrl.Implemented == "NOT_IMPLEMENTED" && ctrl.NotImplementedJustification != nil {
				notImplJustification = *ctrl.NotImplementedJustification
			}

			controlCell := pmTableCellW(colwidths[1], pmControlParagraph(ctrl.SectionTitle, ctrl.Name))

			cells := []prosemirror.Node{
				pmTableCellW(colwidths[0], pmParagraph(group.FrameworkName)),
				controlCell,
				pmTableCellW(colwidths[2], pmParagraph(boolLabel(ctrl.Applicability))),
				pmTableCellW(colwidths[3], pmParagraph(justification)),
				pmTableCellW(colwidths[4], pmParagraph(implemented)),
				pmTableCellW(colwidths[5], pmParagraph(notImplJustification)),
				pmTableCellW(colwidths[6], pmParagraph(boolLabel(ctrl.Regulatory))),
				pmTableCellW(colwidths[7], pmParagraph(boolLabel(ctrl.Contractual))),
				pmTableCellW(colwidths[8], pmParagraph(boolLabel(ctrl.BestPractice))),
				pmTableCellW(colwidths[9], pmParagraph(boolLabel(ctrl.RiskAssessment))),
			}

			rows = append(rows, prosemirror.Node{
				Type:    prosemirror.NodeTableRow,
				Content: cells,
			})
		}
	}

	content = append(content, prosemirror.Node{
		Type:    prosemirror.NodeTable,
		Content: rows,
	})

	content = append(content, buildStatementOfApplicabilityAnnex()...)

	doc := prosemirror.Node{
		Type:    prosemirror.NodeDoc,
		Content: content,
	}

	b, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("cannot marshal prosemirror document: %w", err)
	}

	return string(b), nil
}

func pmHeading(level int, text string) prosemirror.Node {
	attrs, _ := json.Marshal(prosemirror.HeadingAttrs{Level: level})
	return prosemirror.Node{
		Type:  prosemirror.NodeHeading,
		Attrs: attrs,
		Content: []prosemirror.Node{
			pmText(text),
		},
	}
}

func pmParagraph(text string) prosemirror.Node {
	if text == "" {
		return prosemirror.Node{
			Type: prosemirror.NodeParagraph,
		}
	}
	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			pmText(text),
		},
	}
}

func pmText(text string) prosemirror.Node {
	return prosemirror.Node{
		Type: prosemirror.NodeText,
		Text: &text,
	}
}

func pmTableCellW(colwidth []int, content ...prosemirror.Node) prosemirror.Node {
	attrs, _ := json.Marshal(prosemirror.TableCellAttrs{
		Colspan:  1,
		Rowspan:  1,
		Colwidth: colwidth,
	})
	return prosemirror.Node{
		Type:    prosemirror.NodeTableCell,
		Attrs:   attrs,
		Content: content,
	}
}

func pmTableHeaderCellSpanW(text string, colspan, rowspan int, colwidth []int) prosemirror.Node {
	attrs, _ := json.Marshal(prosemirror.TableCellAttrs{
		Colspan:  colspan,
		Rowspan:  rowspan,
		Colwidth: colwidth,
	})
	return prosemirror.Node{
		Type:  prosemirror.NodeTableHeader,
		Attrs: attrs,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeParagraph,
				Content: []prosemirror.Node{
					{
						Type:  prosemirror.NodeText,
						Text:  &text,
						Marks: []prosemirror.Mark{{Type: prosemirror.MarkStrong}},
					},
				},
			},
		},
	}
}

func pmTableHeaderCellW(text string, colwidth []int) prosemirror.Node {
	return pmTableHeaderCellSpanW(text, 1, 1, colwidth)
}

func pmControlParagraph(sectionTitle, name string) prosemirror.Node {
	tag := fmt.Sprintf("[%s] ", sectionTitle)
	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{
				Type:  prosemirror.NodeText,
				Text:  &tag,
				Marks: []prosemirror.Mark{{Type: prosemirror.MarkCode}},
			},
			pmText(name),
		},
	}
}

func buildStatementOfApplicabilityAnnex() []prosemirror.Node {
	naApplicability := pmBulletList(
		pmBoldTextParagraph("Yes: ", "The control is applicable to the organization."),
		pmBoldTextParagraph("No: ", "The control is not applicable to the organization (with justification provided)."),
	)

	naImplemented := pmBulletList(
		pmBoldTextParagraph("Yes: ", "The control has been implemented by the organization."),
		pmBoldTextParagraph("No: ", "The control has not been implemented (with justification provided)."),
		pmBoldTextParagraph("-: ", "Not applicable (control is not applicable)."),
	)

	naYesNoNA := func(yes, no string) prosemirror.Node {
		return pmBulletList(
			pmBoldTextParagraph("Yes: ", yes),
			pmBoldTextParagraph("No: ", no),
			pmBoldTextParagraph("-: ", "Not applicable (control is not applicable)."),
		)
	}

	return []prosemirror.Node{
		{Type: prosemirror.NodeHorizontalRule},
		pmHeading(1, "3. Annexes"),
		pmHeading(2, "3.1 Column Definitions"),

		pmHeading(3, "Framework"),
		pmParagraph("The name of the compliance framework or standard to which the control belongs (e.g., ISO 27001, SOC 2, GDPR)."),

		pmHeading(3, "Control"),
		pmParagraph("The specific control identifier and name within the framework, including its section reference."),

		pmHeading(3, "Applicability"),
		naApplicability,

		pmHeading(3, "Justification for non-applicability"),
		pmParagraph("Provides the rationale when a control is not applicable. This field is empty for applicable controls."),

		pmHeading(3, "Implemented"),
		naImplemented,

		pmHeading(3, "Justification for non-implementation"),
		pmParagraph("Provides the rationale when a control is not implemented. This field is empty for implemented controls or when the control is not applicable."),

		pmHeading(3, "Justification for inclusion"),
		pmParagraph("For applicable controls, this section provides additional context on why the control is included, based on regulatory requirements, contractual obligations, best practices, or risk assessments."),

		pmHeading(4, "Regulatory"),
		naYesNoNA(
			"The control is linked to one or more legal or regulatory obligations.",
			"The control is not associated with any legal or regulatory obligations.",
		),

		pmHeading(4, "Contractual"),
		naYesNoNA(
			"The control is linked to one or more contractual obligations.",
			"The control is not associated with any contractual obligations.",
		),

		pmHeading(4, "Best Practice"),
		naYesNoNA(
			"The control is designated as a best practice recommendation.",
			"The control is not designated as a best practice.",
		),

		pmHeading(4, "Risk Assessment"),
		naYesNoNA(
			"The control is associated with one or more identified risks through risk mitigation measures.",
			"The control is not currently associated with any identified risks.",
		),
	}
}

func pmBulletList(items ...prosemirror.Node) prosemirror.Node {
	listItems := make([]prosemirror.Node, len(items))
	for i, item := range items {
		listItems[i] = prosemirror.Node{
			Type:    prosemirror.NodeListItem,
			Content: []prosemirror.Node{item},
		}
	}
	return prosemirror.Node{
		Type:    prosemirror.NodeBulletList,
		Content: listItems,
	}
}

func pmBoldTextParagraph(bold, text string) prosemirror.Node {
	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{
				Type:  prosemirror.NodeText,
				Text:  &bold,
				Marks: []prosemirror.Mark{{Type: prosemirror.MarkStrong}},
			},
			pmText(text),
		},
	}
}

func boolLabel(v *bool) string {
	if v == nil {
		return "-"
	}
	if *v {
		return "Yes"
	}
	return "No"
}
