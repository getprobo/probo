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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessReviewSourceOrderBy                     OrderBy[coredata.AccessReviewSourceOrderField]
	AccessReviewCampaignOrderBy                   OrderBy[coredata.AccessReviewCampaignOrderField]
	AccessReviewEntryOrderBy                      OrderBy[coredata.AccessReviewEntryOrderField]
	AccessReviewCampaignSourceFetchAttemptOrderBy OrderBy[coredata.AccessReviewCampaignSourceFetchAttemptOrderField]

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

	AccessReviewCampaignSourceFetchAttemptConnection struct {
		TotalCount int
		Edges      []*AccessReviewCampaignSourceFetchAttemptEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
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

// NewAccessReviewCampaignSource builds the GraphQL campaign source from a
// snapshot row. Fetch state is resolved lazily via field resolvers on
// AccessReviewCampaignSource. The live access source is resolved via the
// source field resolver from SourceID.
func NewAccessReviewCampaignSource(
	campaignSource *coredata.AccessReviewCampaignSource,
) *AccessReviewCampaignSource {
	return &AccessReviewCampaignSource{
		ID: campaignSource.ID,
		Campaign: &AccessReviewCampaign{
			ID: campaignSource.AccessReviewCampaignID,
		},
		SourceID: campaignSource.AccessReviewSourceID,
		Name:     campaignSource.Name,
	}
}

// NewAccessReviewCampaignSourceFetchAttempt builds the GraphQL representation of a
// single append-only fetch attempt. attemptNumber is the 1-based position in the
// snapshot's history, counting up from the oldest attempt.
func NewAccessReviewCampaignSourceFetchAttempt(
	a *coredata.AccessReviewCampaignSourceFetchAttempt,
	attemptNumber int,
) *AccessReviewCampaignSourceFetchAttempt {
	if a.AttemptNumber > 0 {
		attemptNumber = a.AttemptNumber
	}

	return &AccessReviewCampaignSourceFetchAttempt{
		ID:                   a.ID,
		AttemptNumber:        attemptNumber,
		Status:               a.Status,
		FetchedAccountsCount: a.FetchedAccountsCount,
		Error:                a.Error,
		StartedAt:            a.StartedAt,
		CompletedAt:          a.CompletedAt,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
	}
}

func NewAccessReviewCampaignSourceFetchAttemptConnection(
	p *page.Page[*coredata.AccessReviewCampaignSourceFetchAttempt, coredata.AccessReviewCampaignSourceFetchAttemptOrderField],
	parentType any,
	parentID gid.GID,
) *AccessReviewCampaignSourceFetchAttemptConnection {
	edges := make([]*AccessReviewCampaignSourceFetchAttemptEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessReviewCampaignSourceFetchAttemptEdge(p.Data[i], p.Cursor.OrderBy.Field, 0)
	}

	return &AccessReviewCampaignSourceFetchAttemptConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAccessReviewCampaignSourceFetchAttemptEdge(
	a *coredata.AccessReviewCampaignSourceFetchAttempt,
	orderBy coredata.AccessReviewCampaignSourceFetchAttemptOrderField,
	attemptNumber int,
) *AccessReviewCampaignSourceFetchAttemptEdge {
	return &AccessReviewCampaignSourceFetchAttemptEdge{
		Cursor: a.CursorKey(orderBy),
		Node:   NewAccessReviewCampaignSourceFetchAttempt(a, attemptNumber),
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
		Name:        c.Name,
		Description: c.Description,
		Status:      c.Status,
		StartedAt:   c.StartedAt,
		CompletedAt: c.CompletedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
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
	roles := e.Roles
	if roles == nil {
		roles = []string{}
	}

	entry := &AccessReviewEntry{
		ID: e.ID,
		Campaign: &AccessReviewCampaign{
			ID: e.AccessReviewCampaignID,
		},
		CampaignSource: &AccessReviewCampaignSource{
			ID: e.AccessReviewCampaignSourceID,
		},
		Email:            e.Email,
		FullName:         e.FullName,
		Roles:            roles,
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
