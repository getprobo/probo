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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessSourceOrderBy         OrderBy[coredata.AccessSourceOrderField]
	AccessReviewCampaignOrderBy OrderBy[coredata.AccessReviewCampaignOrderField]
	AccessEntryOrderBy          OrderBy[coredata.AccessEntryOrderField]

	AccessSourceConnection struct {
		TotalCount int
		Edges      []*AccessSourceEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	AccessReviewCampaignConnection struct {
		TotalCount int
		Edges      []*AccessReviewCampaignEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	AccessEntryConnection struct {
		TotalCount int
		Edges      []*AccessEntryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		SourceID *gid.GID
	}
)

// AccessSource helpers

func NewAccessSourceConnection(
	p *page.Page[*coredata.AccessSource, coredata.AccessSourceOrderField],
	parentType any,
	parentID gid.GID,
) *AccessSourceConnection {
	edges := make([]*AccessSourceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessSourceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessSourceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAccessSourceEdge(s *coredata.AccessSource, orderBy coredata.AccessSourceOrderField) *AccessSourceEdge {
	return &AccessSourceEdge{
		Cursor: s.CursorKey(orderBy),
		Node:   NewAccessSource(s),
	}
}

func NewAccessSource(s *coredata.AccessSource) *AccessSource {
	return &AccessSource{
		ID: s.ID,
		AccessReview: &AccessReview{
			ID: s.AccessReviewID,
		},
		ConnectorID: s.ConnectorID,
		Name:        s.Name,
		CSVData:     s.CsvData,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func NewAccessReviewCampaignScopeSource(
	source *coredata.AccessSource,
	fetch *coredata.AccessReviewCampaignSourceFetch,
) *AccessReviewCampaignScopeSource {
	status := coredata.AccessReviewCampaignSourceFetchStatusQueued
	fetchedAccountsCount := 0
	attemptCount := 0
	var lastError *string
	var fetchStartedAt *time.Time
	var fetchCompletedAt *time.Time
	if fetch != nil {
		status = fetch.Status
		fetchedAccountsCount = fetch.FetchedAccountsCount
		attemptCount = fetch.AttemptCount
		lastError = fetch.LastError
		fetchStartedAt = fetch.StartedAt
		fetchCompletedAt = fetch.CompletedAt
	}

	return &AccessReviewCampaignScopeSource{
		ID:                   source.ID,
		Source:               NewAccessSource(source),
		Name:                 source.Name,
		FetchStatus:          status,
		FetchedAccountsCount: fetchedAccountsCount,
		AttemptCount:         attemptCount,
		LastError:            lastError,
		FetchStartedAt:       fetchStartedAt,
		FetchCompletedAt:     fetchCompletedAt,
	}
}

// AccessReviewCampaign helpers

func NewAccessReviewCampaignConnection(
	p *page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField],
	parentType any,
	parentID gid.GID,
) *AccessReviewCampaignConnection {
	edges := make([]*AccessReviewCampaignEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessReviewCampaignEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessReviewCampaignConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAccessReviewCampaignEdge(c *coredata.AccessReviewCampaign, orderBy coredata.AccessReviewCampaignOrderField) *AccessReviewCampaignEdge {
	return &AccessReviewCampaignEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewAccessReviewCampaign(c),
	}
}

func NewAccessReviewCampaign(c *coredata.AccessReviewCampaign) *AccessReviewCampaign {
	return &AccessReviewCampaign{
		ID: c.ID,
		AccessReview: &AccessReview{
			ID: c.AccessReviewID,
		},
		Name:              c.Name,
		Status:            c.Status,
		StartedAt:         c.StartedAt,
		CompletedAt:       c.CompletedAt,
		FrameworkControls: c.FrameworkControls,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// AccessEntry helpers

func NewAccessEntryConnection(
	p *page.Page[*coredata.AccessEntry, coredata.AccessEntryOrderField],
	parentType any,
	parentID gid.GID,
	sourceID *gid.GID,
) *AccessEntryConnection {
	edges := make([]*AccessEntryEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessEntryEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessEntryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		SourceID: sourceID,
	}
}

func NewAccessEntryEdge(e *coredata.AccessEntry, orderBy coredata.AccessEntryOrderField) *AccessEntryEdge {
	return &AccessEntryEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewAccessEntry(e),
	}
}

func NewAccessEntry(e *coredata.AccessEntry) *AccessEntry {
	entry := &AccessEntry{
		ID: e.ID,
		Campaign: &AccessReviewCampaign{
			ID: e.AccessReviewCampaignID,
		},
		AccessSource: &AccessSource{
			ID: e.AccessSourceID,
		},
		Email:            e.Email,
		FullName:         e.FullName,
		Role:             e.Role,
		JobTitle:         e.JobTitle,
		IsAdmin:          e.IsAdmin,
		MfaStatus:        e.MFAStatus,
		AuthMethod:       e.AuthMethod,
		LastLogin:        e.LastLogin,
		AccountCreatedAt: e.AccountCreatedAt,
		ExternalID:       e.ExternalID,
		IncrementalTag:   e.IncrementalTag,
		Flag:             e.Flag,
		FlagReason:       e.FlagReason,
		Decision:         e.Decision,
		DecisionNote:     e.DecisionNote,
		DecidedAt:        e.DecidedAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}

	if e.DecidedBy != nil {
		entry.DecidedBy = &People{ID: *e.DecidedBy}
	}

	return entry
}

// AccessReview helper

func NewAccessReview(ar *coredata.AccessReview) *AccessReview {
	review := &AccessReview{
		ID: ar.ID,
		Organization: &Organization{
			ID: ar.OrganizationID,
		},
		CreatedAt: ar.CreatedAt,
		UpdatedAt: ar.UpdatedAt,
	}

	if ar.IdentitySourceID != nil {
		review.IdentitySource = &AccessSource{
			ID: *ar.IdentitySourceID,
		}
	}

	return review
}
