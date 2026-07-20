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
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrCampaignNoScopeSources             = errors.New("cannot start campaign: no scope sources configured")
	ErrCampaignInvalidStatus              = errors.New("campaign status does not allow this operation")
	ErrCampaignSourceOrganizationMismatch = errors.New("access source does not belong to the same organization")
)

type CampaignInvalidStatusError struct {
	Operation string
	Status    coredata.AccessReviewCampaignStatus
	Expected  coredata.AccessReviewCampaignStatus
}

func (e *CampaignInvalidStatusError) Error() string {
	switch e.Operation {
	case "add scope source", "remove scope source":
		return fmt.Sprintf(
			"cannot %s: campaign status is %s, expected %s",
			e.Operation,
			e.Status,
			e.Expected,
		)
	default:
		return fmt.Sprintf(
			"cannot %s campaign: status is %s, expected %s",
			e.Operation,
			e.Status,
			e.Expected,
		)
	}
}

func (e *CampaignInvalidStatusError) Is(target error) bool {
	return target == ErrCampaignInvalidStatus
}

type CampaignSourceOrganizationMismatchError struct {
	Operation string
	SourceID  gid.GID
}

func (e *CampaignSourceOrganizationMismatchError) Error() string {
	switch e.Operation {
	case "create":
		return fmt.Sprintf(
			"cannot create campaign: access source %s does not belong to the same organization",
			e.SourceID,
		)
	case "update":
		return fmt.Sprintf(
			"cannot update campaign: access source %s does not belong to the same organization",
			e.SourceID,
		)
	default:
		return fmt.Sprintf(
			"cannot add scope source: access source %q does not belong to the same organization",
			e.SourceID,
		)
	}
}

func (e *CampaignSourceOrganizationMismatchError) Is(target error) bool {
	return target == ErrCampaignSourceOrganizationMismatch
}

func IsCampaignClientError(err error) bool {
	return errors.Is(err, ErrCampaignNoScopeSources) ||
		errors.Is(err, ErrCampaignInvalidStatus) ||
		errors.Is(err, ErrCampaignSourceOrganizationMismatch)
}
