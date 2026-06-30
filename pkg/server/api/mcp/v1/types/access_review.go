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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewAccessReviewSource(s *coredata.AccessReviewSource) *AccessReviewSource {
	return &AccessReviewSource{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		ConnectorID:    s.ConnectorID,
		Name:           s.Name,
		CsvData:        s.CsvData,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewListAccessReviewSourcesOutput(
	p *page.Page[*coredata.AccessReviewSource, coredata.AccessReviewSourceOrderField],
) ListAccessReviewSourcesOutput {
	sources := make([]*AccessReviewSource, 0, len(p.Data))
	for _, s := range p.Data {
		sources = append(sources, NewAccessReviewSource(s))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessReviewSourcesOutput{
		NextCursor:          nextCursor,
		AccessReviewSources: sources,
	}
}

func NewAccessReviewCampaign(c *coredata.AccessReviewCampaign) *AccessReviewCampaign {
	return &AccessReviewCampaign{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		Name:           c.Name,
		Description:    &c.Description,
		Status:         c.Status,
		StartedAt:      c.StartedAt,
		CompletedAt:    c.CompletedAt,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func NewAccessReviewEntry(e *coredata.AccessReviewEntry) *AccessReviewEntry {
	roles := e.Roles
	if roles == nil {
		roles = []string{}
	}

	entry := &AccessReviewEntry{
		ID:                           e.ID,
		CampaignID:                   e.AccessReviewCampaignID,
		AccessReviewCampaignSourceID: e.AccessReviewCampaignSourceID,
		Email:                        e.Email,
		FullName:                     e.FullName,
		Roles:                        roles,
		JobTitle:                     e.JobTitle,
		IsAdmin:                      e.IsAdmin,
		Active:                       e.Active,
		MfaStatus:                    e.MFAStatus,
		AuthMethod:                   e.AuthMethod,
		AccountType:                  e.AccountType,
		LastLogin:                    e.LastLogin,
		AccountCreatedAt:             e.AccountCreatedAt,
		ExternalID:                   e.ExternalID,
		IncrementalTag:               e.IncrementalTag,
		Flags:                        e.Flags,
		FlagReasons:                  e.FlagReasons,
		Decision:                     e.Decision,
		DecisionNote:                 e.DecisionNote,
		DecidedBy:                    e.DecidedBy,
		DecidedAt:                    e.DecidedAt,
		CreatedAt:                    e.CreatedAt,
		UpdatedAt:                    e.UpdatedAt,
	}

	return entry
}

func NewListAccessReviewCampaignsOutput(
	p *page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField],
) ListAccessReviewCampaignsOutput {
	campaigns := make([]*AccessReviewCampaign, 0, len(p.Data))
	for _, c := range p.Data {
		campaigns = append(campaigns, NewAccessReviewCampaign(c))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessReviewCampaignsOutput{
		NextCursor: nextCursor,
		Campaigns:  campaigns,
	}
}

func NewListAccessEntriesOutput(
	p *page.Page[*coredata.AccessReviewEntry, coredata.AccessReviewEntryOrderField],
) ListAccessEntriesOutput {
	entries := make([]*AccessReviewEntry, 0, len(p.Data))
	for _, e := range p.Data {
		entries = append(entries, NewAccessReviewEntry(e))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAccessEntriesOutput{
		NextCursor: nextCursor,
		Entries:    entries,
	}
}

func NewAccessReviewStatistics(s *coredata.AccessReviewStatistics) *AccessReviewStatistics {
	decisionCounts := make(map[string]any, len(s.DecisionCounts))
	for k, v := range s.DecisionCounts {
		decisionCounts[string(k)] = v
	}

	flagCounts := make(map[string]any, len(s.FlagCounts))
	for k, v := range s.FlagCounts {
		flagCounts[string(k)] = v
	}

	incrementalTagCounts := make(map[string]any, len(s.IncrementalTagCounts))
	for k, v := range s.IncrementalTagCounts {
		incrementalTagCounts[string(k)] = v
	}

	return &AccessReviewStatistics{
		TotalCount:           s.TotalCount,
		DecisionCounts:       decisionCounts,
		FlagCounts:           flagCounts,
		IncrementalTagCounts: incrementalTagCounts,
	}
}
