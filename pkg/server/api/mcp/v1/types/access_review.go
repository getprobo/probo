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
