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

package accessreview

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

const campaignNameMaxLength = 1000

type (
	CreateAccessReviewCampaignRequest struct {
		OrganizationID    gid.GID
		Name              string
		Description       string
		FrameworkControls []string
		AccessSourceIDs   []gid.GID
	}

	UpdateAccessReviewCampaignRequest struct {
		CampaignID        gid.GID
		Name              *string
		Description       *string
		FrameworkControls *[]string
	}

	AddCampaignScopeSourceRequest struct {
		CampaignID     gid.GID
		AccessSourceID gid.GID
	}

	RemoveCampaignScopeSourceRequest struct {
		CampaignID     gid.GID
		AccessSourceID gid.GID
	}
)

func (r *CreateAccessReviewCampaignRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(campaignNameMaxLength))

	return v.Error()
}

func (r *UpdateAccessReviewCampaignRequest) Validate() error {
	v := validator.New()

	v.Check(r.CampaignID, "campaign_id", validator.Required(), validator.GID(coredata.AccessReviewCampaignEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(campaignNameMaxLength))

	return v.Error()
}
