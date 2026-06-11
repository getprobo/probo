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
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessReviewSourceOrderBy   OrderBy[coredata.AccessReviewSourceOrderField]
	AccessReviewCampaignOrderBy OrderBy[coredata.AccessReviewCampaignOrderField]
	AccessReviewEntryOrderBy    OrderBy[coredata.AccessReviewEntryOrderField]

	AccessReviewSourceConnection struct {
		TotalCount int
		Edges      []*AccessReviewSourceEdge
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

	AccessReviewEntryConnection struct {
		TotalCount int
		Edges      []*AccessReviewEntryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		SourceID *gid.GID
		Filter   *coredata.AccessReviewEntryFilter
	}
)

// AccessReviewSource helpers

func NewAccessReviewSourceConnection(
	p *page.Page[*coredata.AccessReviewSource, coredata.AccessReviewSourceOrderField],
	parentType any,
	parentID gid.GID,
) *AccessReviewSourceConnection {
	edges := make([]*AccessReviewSourceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessReviewSourceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessReviewSourceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAccessReviewSourceEdge(s *coredata.AccessReviewSource, orderBy coredata.AccessReviewSourceOrderField) *AccessReviewSourceEdge {
	return &AccessReviewSourceEdge{
		Cursor: s.CursorKey(orderBy),
		Node:   NewAccessReviewSource(s),
	}
}

func NewAccessReviewSource(s *coredata.AccessReviewSource) *AccessReviewSource {
	return &AccessReviewSource{
		ID: s.ID,
		Organization: &Organization{
			ID: s.OrganizationID,
		},
		ConnectorID: s.ConnectorID,
		Name:        s.Name,
		CSVData:     s.CsvData,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// NewAccessReviewCampaignSource builds the GraphQL scope source from a
// campaign source snapshot. The current fetch state is derived from the latest
// fetch attempt (nil when the source has never been fetched). The live access
// source is resolved lazily via the source field resolver from SourceID.
func NewAccessReviewCampaignSource(
	campaignSource *coredata.AccessReviewCampaignSource,
	latestAttempt *coredata.AccessReviewCampaignSourceFetchAttempt,
) *AccessReviewCampaignSource {
	status := coredata.AccessReviewCampaignSourceFetchStatusQueued
	fetchedAccountsCount := 0
	attemptCount := 0

	var (
		lastError        *string
		fetchStartedAt   *time.Time
		fetchCompletedAt *time.Time
	)

	if latestAttempt != nil {
		status = latestAttempt.Status
		fetchedAccountsCount = latestAttempt.FetchedAccountsCount
		attemptCount = latestAttempt.AttemptNumber
		lastError = latestAttempt.Error
		fetchStartedAt = latestAttempt.StartedAt
		fetchCompletedAt = latestAttempt.CompletedAt
	}

	return &AccessReviewCampaignSource{
		ID:                   campaignSource.ID,
		CampaignID:           campaignSource.AccessReviewCampaignID,
		SourceID:             campaignSource.AccessReviewSourceID,
		Name:                 campaignSource.Name,
		FetchStatus:          status,
		FetchedAccountsCount: fetchedAccountsCount,
		AttemptCount:         attemptCount,
		LastError:            lastError,
		FetchStartedAt:       fetchStartedAt,
		FetchCompletedAt:     fetchCompletedAt,
	}
}

// NewAccessReviewCampaignSourceFetchAttempt builds the GraphQL representation of a
// single append-only fetch attempt.
func NewAccessReviewCampaignSourceFetchAttempt(a *coredata.AccessReviewCampaignSourceFetchAttempt) *AccessReviewCampaignSourceFetchAttempt {
	return &AccessReviewCampaignSourceFetchAttempt{
		ID:                   a.ID,
		AttemptNumber:        a.AttemptNumber,
		Status:               a.Status,
		FetchedAccountsCount: a.FetchedAccountsCount,
		Error:                a.Error,
		StartedAt:            a.StartedAt,
		CompletedAt:          a.CompletedAt,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
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
	campaign := &AccessReviewCampaign{
		ID: c.ID,
		Organization: &Organization{
			ID: c.OrganizationID,
		},
		Name:              c.Name,
		Description:       c.Description,
		Status:            c.Status,
		StartedAt:         c.StartedAt,
		CompletedAt:       c.CompletedAt,
		FrameworkControls: c.FrameworkControls,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}

	return campaign
}

func NewAccessReviewEntryDecisionHistoryEntry(h *coredata.AccessReviewEntryDecisionHistory) *AccessReviewEntryDecisionHistoryEntry {
	entry := &AccessReviewEntryDecisionHistoryEntry{
		ID:           h.ID,
		Decision:     h.Decision,
		DecisionNote: h.DecisionNote,
		DecidedAt:    h.DecidedAt,
		CreatedAt:    h.CreatedAt,
	}

	if h.DecidedBy != nil {
		entry.DecidedBy = h.DecidedBy
	}

	return entry
}

// AccessReviewEntry helpers

func NewAccessReviewEntryConnection(
	p *page.Page[*coredata.AccessReviewEntry, coredata.AccessReviewEntryOrderField],
	parentType any,
	parentID gid.GID,
	sourceID *gid.GID,
	filter *coredata.AccessReviewEntryFilter,
) *AccessReviewEntryConnection {
	edges := make([]*AccessReviewEntryEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessReviewEntryEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessReviewEntryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		SourceID: sourceID,
		Filter:   filter,
	}
}

func NewAccessReviewEntryEdge(e *coredata.AccessReviewEntry, orderBy coredata.AccessReviewEntryOrderField) *AccessReviewEntryEdge {
	return &AccessReviewEntryEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewAccessReviewEntry(e),
	}
}

func NewAccessReviewEntry(e *coredata.AccessReviewEntry) *AccessReviewEntry {
	entry := &AccessReviewEntry{
		ID: e.ID,
		Campaign: &AccessReviewCampaign{
			ID: e.AccessReviewCampaignID,
		},
		Email:            e.Email,
		FullName:         e.FullName,
		Role:             e.Role,
		JobTitle:         e.JobTitle,
		IsAdmin:          e.IsAdmin,
		Active:           e.Active,
		MfaStatus:        e.MFAStatus,
		AuthMethod:       e.AuthMethod,
		AccountType:      e.AccountType,
		LastLogin:        e.LastLogin,
		AccountCreatedAt: e.AccountCreatedAt,
		ExternalID:       e.ExternalID,
		IncrementalTag:   e.IncrementalTag,
		Flags:            e.Flags,
		FlagReasons:      e.FlagReasons,
		Decision:         e.Decision,
		DecisionNote:     e.DecisionNote,
		DecidedAt:        e.DecidedAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}

	if e.DecidedBy != nil {
		entry.DecidedBy = e.DecidedBy
	}

	return entry
}

func NewAccessReviewStatistics(stats *coredata.AccessReviewStatistics) *AccessReviewStatistics {
	decisionCounts := make([]*AccessReviewEntryDecisionCount, 0, len(stats.DecisionCounts))
	for decision, count := range stats.DecisionCounts {
		decisionCounts = append(
			decisionCounts,
			&AccessReviewEntryDecisionCount{Decision: decision, Count: count},
		)
	}

	flagCounts := make([]*AccessReviewEntryFlagCount, 0, len(stats.FlagCounts))
	for flag, count := range stats.FlagCounts {
		flagCounts = append(
			flagCounts,
			&AccessReviewEntryFlagCount{Flag: flag, Count: count},
		)
	}

	incrementalTagCounts := make([]*AccessReviewEntryIncrementalTagCount, 0, len(stats.IncrementalTagCounts))
	for tag, count := range stats.IncrementalTagCounts {
		incrementalTagCounts = append(
			incrementalTagCounts,
			&AccessReviewEntryIncrementalTagCount{IncrementalTag: tag, Count: count},
		)
	}

	return &AccessReviewStatistics{
		TotalCount:           stats.TotalCount,
		DecisionCounts:       decisionCounts,
		FlagCounts:           flagCounts,
		IncrementalTagCounts: incrementalTagCounts,
	}
}
