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

package accessreview

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

const campaignNameMaxLength = 255

type (
	CreateAccessReviewCampaignRequest struct {
		OrganizationID        gid.GID
		Name                  string
		Description           string
		AccessReviewSourceIDs []gid.GID
	}

	UpdateAccessReviewCampaignRequest struct {
		CampaignID            gid.GID
		Name                  **string
		Description           **string
		AccessReviewSourceIDs *[]gid.GID
	}

	AddCampaignSourceRequest struct {
		CampaignID           gid.GID
		AccessReviewSourceID gid.GID
	}

	RemoveCampaignSourceRequest struct {
		CampaignID           gid.GID
		AccessReviewSourceID gid.GID
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
