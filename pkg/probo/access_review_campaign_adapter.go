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

	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type accessReviewCampaignAdapter struct {
	svc accessreview.AccessReviewCampaignService
}

func newAccessReviewCampaignAdapter(svc accessreview.AccessReviewCampaignService) AccessReviewCampaignService {
	return &accessReviewCampaignAdapter{svc: svc}
}

func (a *accessReviewCampaignAdapter) Create(ctx context.Context, req CreateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Create(ctx, accessreview.CreateAccessReviewCampaignRequest{
		OrganizationID:    req.OrganizationID,
		Name:              req.Name,
		FrameworkControls: req.FrameworkControls,
		AccessSourceIDs:   req.AccessSourceIDs,
	})
}

func (a *accessReviewCampaignAdapter) Get(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Get(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) Update(ctx context.Context, req UpdateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Update(ctx, accessreview.UpdateAccessReviewCampaignRequest{
		CampaignID:        req.CampaignID,
		Name:              req.Name,
		FrameworkControls: req.FrameworkControls,
	})
}

func (a *accessReviewCampaignAdapter) Delete(ctx context.Context, campaignID gid.GID) error {
	return a.svc.Delete(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) AddScopeSource(ctx context.Context, req AddCampaignScopeSourceRequest) (*coredata.AccessReviewCampaign, error) {
	return a.svc.AddScopeSource(ctx, accessreview.AddCampaignScopeSourceRequest{
		CampaignID:     req.CampaignID,
		AccessSourceID: req.AccessSourceID,
	})
}

func (a *accessReviewCampaignAdapter) RemoveScopeSource(ctx context.Context, req RemoveCampaignScopeSourceRequest) (*coredata.AccessReviewCampaign, error) {
	return a.svc.RemoveScopeSource(ctx, accessreview.RemoveCampaignScopeSourceRequest{
		CampaignID:     req.CampaignID,
		AccessSourceID: req.AccessSourceID,
	})
}

func (a *accessReviewCampaignAdapter) Start(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Start(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) Close(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Close(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) Cancel(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Cancel(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	return a.svc.ListForOrganizationID(ctx, organizationID, cursor)
}

func (a *accessReviewCampaignAdapter) ListSourceFetches(ctx context.Context, campaignID gid.GID) (coredata.AccessReviewCampaignSourceFetches, error) {
	return a.svc.ListSourceFetches(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) CountForOrganizationID(ctx context.Context, organizationID gid.GID) (int, error) {
	return a.svc.CountForOrganizationID(ctx, organizationID)
}
