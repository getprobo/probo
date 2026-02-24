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

package probo

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

// AccessReviewCampaignService defines the interface for access review campaign operations.
// The implementation lives in pkg/accessreview.
type AccessReviewCampaignService interface {
	Create(ctx context.Context, req CreateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	Get(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	Update(ctx context.Context, req UpdateAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	Delete(ctx context.Context, campaignID gid.GID) error
	Start(ctx context.Context, req StartAccessReviewCampaignRequest) (*coredata.AccessReviewCampaign, error)
	RetryStart(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	Close(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
	ValidateForClose(ctx context.Context, campaignID gid.GID, validatedBy *gid.GID, note *string) error
	ExportEvidence(ctx context.Context, campaignID gid.GID) (string, string, error)
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
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *UpdateAccessReviewCampaignRequest) Validate() error {
	v := validator.New()

	v.Check(r.CampaignID, "campaign_id", validator.Required(), validator.GID(coredata.AccessReviewCampaignEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}
