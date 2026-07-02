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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Document struct {
	ID                    gid.GID                        `json:"id"`
	OrganizationID        gid.GID                        `json:"organizationId"`
	Title                 string                         `json:"title"`
	DocumentType          coredata.DocumentType          `json:"documentType"`
	Status                coredata.DocumentStatus        `json:"status"`
	TrustCenterVisibility coredata.TrustCenterVisibility `json:"trustCenterVisibility"`
	CurrentPublishedMajor *int                           `json:"currentPublishedMajor"`
	CurrentPublishedMinor *int                           `json:"currentPublishedMinor"`
	ArchivedAt            *time.Time                     `json:"archivedAt"`
	CreatedAt             time.Time                      `json:"createdAt"`
	UpdatedAt             time.Time                      `json:"updatedAt"`
}

type DocumentVersion struct {
	ID             gid.GID                         `json:"id"`
	DocumentID     gid.GID                         `json:"documentId"`
	Title          string                          `json:"title"`
	Major          int                             `json:"major"`
	Minor          int                             `json:"minor"`
	Classification coredata.DocumentClassification `json:"classification"`
	DocumentType   coredata.DocumentType           `json:"documentType"`
	Changelog      string                          `json:"changelog"`
	Status         coredata.DocumentVersionStatus  `json:"status"`
	PublishedAt    *time.Time                      `json:"publishedAt"`
	CreatedAt      time.Time                       `json:"createdAt"`
	UpdatedAt      time.Time                       `json:"updatedAt"`
	Document       *Document                       `json:"document"`
}

type DocumentVersionSignature struct {
	ID                gid.GID                                `json:"id"`
	DocumentVersionID gid.GID                                `json:"documentVersionId"`
	State             coredata.DocumentVersionSignatureState `json:"state"`
	SignedBy          gid.GID                                `json:"signedBy"`
	SignedAt          *time.Time                             `json:"signedAt"`
	RequestedAt       time.Time                              `json:"requestedAt"`
	CreatedAt         time.Time                              `json:"createdAt"`
	UpdatedAt         time.Time                              `json:"updatedAt"`
	Version           *DocumentVersion                       `json:"version"`
}

type DocumentApprovalQuorum struct {
	ID        gid.GID                                      `json:"id"`
	VersionID gid.GID                                      `json:"versionId"`
	Status    coredata.DocumentVersionApprovalQuorumStatus `json:"status"`
	CreatedAt time.Time                                    `json:"createdAt"`
	UpdatedAt time.Time                                    `json:"updatedAt"`
	Decisions []*DocumentApprovalDecision                  `json:"decisions"`
	Version   *DocumentVersion                             `json:"version"`
}

type DocumentApprovalDecision struct {
	ID         gid.GID                                       `json:"id"`
	ApproverID gid.GID                                       `json:"approverId"`
	State      coredata.DocumentVersionApprovalDecisionState `json:"state"`
	Comment    *string                                       `json:"comment"`
	DecidedAt  *time.Time                                    `json:"decidedAt"`
	CreatedAt  time.Time                                     `json:"createdAt"`
	UpdatedAt  time.Time                                     `json:"updatedAt"`
}

func NewDocument(d *coredata.Document) *Document {
	return &Document{
		ID:                    d.ID,
		OrganizationID:        d.OrganizationID,
		Title:                 d.Title,
		DocumentType:          d.DocumentType,
		Status:                d.Status,
		TrustCenterVisibility: d.TrustCenterVisibility,
		CurrentPublishedMajor: d.CurrentPublishedMajor,
		CurrentPublishedMinor: d.CurrentPublishedMinor,
		ArchivedAt:            d.ArchivedAt,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}

func NewDocumentVersion(v *coredata.DocumentVersion, d *coredata.Document) *DocumentVersion {
	return &DocumentVersion{
		ID:             v.ID,
		DocumentID:     v.DocumentID,
		Title:          v.Title,
		Major:          v.Major,
		Minor:          v.Minor,
		Classification: v.Classification,
		DocumentType:   v.DocumentType,
		Changelog:      v.Changelog,
		Status:         v.Status,
		PublishedAt:    v.PublishedAt,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		Document:       NewDocument(d),
	}
}

func NewDocumentVersionSignature(
	s *coredata.DocumentVersionSignature,
	v *coredata.DocumentVersion,
	d *coredata.Document,
) *DocumentVersionSignature {
	return &DocumentVersionSignature{
		ID:                s.ID,
		DocumentVersionID: s.DocumentVersionID,
		State:             s.State,
		SignedBy:          s.SignedBy,
		SignedAt:          s.SignedAt,
		RequestedAt:       s.RequestedAt,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
		Version:           NewDocumentVersion(v, d),
	}
}

func NewDocumentApprovalQuorum(
	quorum *coredata.DocumentVersionApprovalQuorum,
	decisions coredata.DocumentVersionApprovalDecisions,
	v *coredata.DocumentVersion,
	d *coredata.Document,
) *DocumentApprovalQuorum {
	decisionPayloads := make([]*DocumentApprovalDecision, 0, len(decisions))
	for _, decision := range decisions {
		decisionPayloads = append(decisionPayloads, &DocumentApprovalDecision{
			ID:         decision.ID,
			ApproverID: decision.ApproverID,
			State:      decision.State,
			Comment:    decision.Comment,
			DecidedAt:  decision.DecidedAt,
			CreatedAt:  decision.CreatedAt,
			UpdatedAt:  decision.UpdatedAt,
		})
	}

	return &DocumentApprovalQuorum{
		ID:        quorum.ID,
		VersionID: quorum.VersionID,
		Status:    quorum.Status,
		CreatedAt: quorum.CreatedAt,
		UpdatedAt: quorum.UpdatedAt,
		Decisions: decisionPayloads,
		Version:   NewDocumentVersion(v, d),
	}
}
