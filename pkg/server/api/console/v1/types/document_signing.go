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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewDocumentSigningConfig(c *coredata.DocumentSigningConfig) *DocumentSigningConfig {
	return &DocumentSigningConfig{
		ID: c.ID,
		Document: &Document{
			ID: c.DocumentID,
		},
		PublishableKey:     c.PublishableKey,
		State:              c.State,
		Mode:               c.Mode,
		Kind:               c.Kind,
		ReacceptancePolicy: c.ReacceptancePolicy,
		AllowedOrigins:     c.AllowedOrigins,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

func NewClickwrapAcceptance(a *coredata.ClickwrapAcceptance) *ClickwrapAcceptance {
	return &ClickwrapAcceptance{
		ID: a.ID,
		SigningConfig: &DocumentSigningConfig{
			ID: a.SigningConfigID,
		},
		DocumentVersion: &DocumentVersion{
			ID: a.DocumentVersionID,
		},
		Subject: &ClickwrapSubject{
			ID: a.SubjectID,
		},
		AcceptedEmail:    a.AcceptedEmail,
		AcceptedFullName: a.AcceptedFullName,
		ConsentText:      a.ConsentText,
		ContentSha256:    a.ContentSha256,
		IPAddress:        a.IPAddress,
		Mode:             a.Mode,
		Sealed:           a.SealedAt != nil,
		RevokedAt:        a.RevokedAt,
		AcceptedAt:       a.AcceptedAt,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func NewClickwrapAcceptanceEdge(
	a *coredata.ClickwrapAcceptance,
	orderBy coredata.ClickwrapAcceptanceOrderField,
) *ClickwrapAcceptanceEdge {
	return &ClickwrapAcceptanceEdge{
		Cursor: a.CursorKey(orderBy).String(),
		Node:   NewClickwrapAcceptance(a),
	}
}

func NewClickwrapAcceptanceConnection(
	p *page.Page[*coredata.ClickwrapAcceptance, coredata.ClickwrapAcceptanceOrderField],
) *ClickwrapAcceptanceConnection {
	edges := make([]*ClickwrapAcceptanceEdge, len(p.Data))
	for i, a := range p.Data {
		edges[i] = NewClickwrapAcceptanceEdge(a, p.Cursor.OrderBy.Field)
	}
	return &ClickwrapAcceptanceConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewClickwrapSubject(s *coredata.ClickwrapSubject) *ClickwrapSubject {
	return &ClickwrapSubject{
		ID: s.ID,
		Organization: &Organization{
			ID: s.OrganizationID,
		},
		ExternalUserID:    s.ExternalUserID,
		Email:             s.Email,
		FullName:          s.FullName,
		ExternalAccountID: s.ExternalAccountID,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

func NewClickwrapSubjectEdge(
	s *coredata.ClickwrapSubject,
	orderBy coredata.ClickwrapSubjectOrderField,
) *ClickwrapSubjectEdge {
	return &ClickwrapSubjectEdge{
		Cursor: s.CursorKey(orderBy).String(),
		Node:   NewClickwrapSubject(s),
	}
}

func NewClickwrapSubjectConnection(
	p *page.Page[*coredata.ClickwrapSubject, coredata.ClickwrapSubjectOrderField],
) *ClickwrapSubjectConnection {
	edges := make([]*ClickwrapSubjectEdge, len(p.Data))
	for i, s := range p.Data {
		edges[i] = NewClickwrapSubjectEdge(s, p.Cursor.OrderBy.Field)
	}
	return &ClickwrapSubjectConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewDataPlaneRequestLogEntry(l *coredata.DataPlaneRequestLog) *DataPlaneRequestLogEntry {
	return &DataPlaneRequestLogEntry{
		ID:          l.ID,
		RequestPath: l.RequestPath,
		Method:      l.Method,
		StatusCode:  l.StatusCode,
		ErrorCode:   l.ErrorCode,
		RequestID:   l.RequestID,
		OccurredAt:  l.OccurredAt,
	}
}
