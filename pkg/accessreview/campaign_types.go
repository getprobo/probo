package accessreview

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const campaignNameMaxLength = 1000

type AccessReviewCampaignService interface {
	Create(ctx context.Context, req CreateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	Get(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	Update(ctx context.Context, req UpdateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	Delete(ctx context.Context, campaignID gid.GID) error
	Start(ctx context.Context, req StartAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	Close(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	Cancel(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	ListForAccessReviewID(ctx context.Context, accessReviewID gid.GID, cursor *page.Cursor[coredata.AccessReviewCampaignOrderField]) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error)
	ListSourceFetches(ctx context.Context, campaignID gid.GID) (coredata.AccessReviewCampaignSourceFetches, error)
	CountForAccessReviewID(ctx context.Context, accessReviewID gid.GID) (int, error)
}

type (
	CreateAccessReviewCampaignRequest struct {
		AccessReviewID    gid.GID
		Name              string
		FrameworkControls []string
	}

	UpdateAccessReviewCampaignRequest struct {
		CampaignID        gid.GID
		Name              *string
		FrameworkControls *[]string
	}

	StartAccessReviewCampaignRequest struct {
		CampaignID      gid.GID
		AccessSourceIDs []gid.GID
	}
)

func (r *CreateAccessReviewCampaignRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessReviewID, "access_review_id", validator.Required(), validator.GID(coredata.AccessReviewEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(campaignNameMaxLength))

	return v.Error()
}

func (r *UpdateAccessReviewCampaignRequest) Validate() error {
	v := validator.New()

	v.Check(r.CampaignID, "campaign_id", validator.Required(), validator.GID(coredata.AccessReviewCampaignEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(campaignNameMaxLength))

	return v.Error()
}
