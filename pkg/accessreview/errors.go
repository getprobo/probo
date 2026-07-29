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

	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrCampaignMissingSources    = errors.New("access review campaign missing scope sources")
	ErrCampaignNotDraft          = errors.New("access review campaign not draft")
	ErrCampaignNotPendingActions = errors.New("access review campaign not pending actions")
	ErrCampaignCompleted         = errors.New("access review campaign completed")
	ErrCampaignCancelled         = errors.New("access review campaign cancelled")
)

type (
	CampaignMissingSourcesError struct {
		CampaignID gid.GID
	}

	CampaignNotDraftError struct {
		CampaignID gid.GID
	}

	CampaignNotPendingActionsError struct {
		CampaignID gid.GID
	}

	CampaignCompletedError struct {
		CampaignID gid.GID
	}

	CampaignCancelledError struct {
		CampaignID gid.GID
	}
)

func NewCampaignMissingSourcesError(campaignID gid.GID) error {
	return &CampaignMissingSourcesError{CampaignID: campaignID}
}

func (e *CampaignMissingSourcesError) Error() string {
	return fmt.Sprintf(
		"access review campaign %q cannot be started: no scope sources configured",
		e.CampaignID,
	)
}

func (e *CampaignMissingSourcesError) Is(target error) bool {
	return target == ErrCampaignMissingSources
}

func NewCampaignNotDraftError(campaignID gid.GID) error {
	return &CampaignNotDraftError{CampaignID: campaignID}
}

func (e *CampaignNotDraftError) Error() string {
	return fmt.Sprintf("access review campaign %q is not in draft", e.CampaignID)
}

func (e *CampaignNotDraftError) Is(target error) bool {
	return target == ErrCampaignNotDraft
}

func NewCampaignNotPendingActionsError(campaignID gid.GID) error {
	return &CampaignNotPendingActionsError{CampaignID: campaignID}
}

func (e *CampaignNotPendingActionsError) Error() string {
	return fmt.Sprintf(
		"access review campaign %q cannot be closed unless it is pending actions",
		e.CampaignID,
	)
}

func (e *CampaignNotPendingActionsError) Is(target error) bool {
	return target == ErrCampaignNotPendingActions
}

func NewCampaignCompletedError(campaignID gid.GID) error {
	return &CampaignCompletedError{CampaignID: campaignID}
}

func (e *CampaignCompletedError) Error() string {
	return fmt.Sprintf("access review campaign %q is already completed", e.CampaignID)
}

func (e *CampaignCompletedError) Is(target error) bool {
	return target == ErrCampaignCompleted
}

func NewCampaignCancelledError(campaignID gid.GID) error {
	return &CampaignCancelledError{CampaignID: campaignID}
}

func (e *CampaignCancelledError) Error() string {
	return fmt.Sprintf("access review campaign %q is already cancelled", e.CampaignID)
}

func (e *CampaignCancelledError) Is(target error) bool {
	return target == ErrCampaignCancelled
}
