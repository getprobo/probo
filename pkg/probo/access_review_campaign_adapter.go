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
		AccessReviewID:    req.AccessReviewID,
		Name:              req.Name,
		FrameworkControls: req.FrameworkControls,
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

func (a *accessReviewCampaignAdapter) Start(ctx context.Context, req StartAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Start(ctx, accessreview.StartAccessReviewCampaignRequest{
		CampaignID:      req.CampaignID,
		AccessSourceIDs: req.AccessSourceIDs,
	})
}

func (a *accessReviewCampaignAdapter) RetryStart(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.RetryStart(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) Close(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Close(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) ValidateForClose(ctx context.Context, campaignID gid.GID, validatedBy *gid.GID, note *string) error {
	return a.svc.ValidateForClose(ctx, campaignID, validatedBy, note)
}

func (a *accessReviewCampaignAdapter) ExportEvidence(ctx context.Context, campaignID gid.GID) (string, string, error) {
	return a.svc.ExportEvidence(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) Cancel(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error) {
	return a.svc.Cancel(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) ListForAccessReviewID(
	ctx context.Context,
	accessReviewID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	return a.svc.ListForAccessReviewID(ctx, accessReviewID, cursor)
}

func (a *accessReviewCampaignAdapter) ListSourceFetches(ctx context.Context, campaignID gid.GID) (coredata.AccessReviewCampaignSourceFetches, error) {
	return a.svc.ListSourceFetches(ctx, campaignID)
}

func (a *accessReviewCampaignAdapter) CountForAccessReviewID(ctx context.Context, accessReviewID gid.GID) (int, error) {
	return a.svc.CountForAccessReviewID(ctx, accessReviewID)
}
