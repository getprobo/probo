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

package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/vetting"
)

const (
	vettingErrorMessageMaxLen  = 512
	vettingWebsiteURLMaxLength = 2048
	vettingProcedureMaxLength  = 5000
)

var (
	ErrVettingDisabled   = errors.New("thirdParty vetting is not configured on this deployment")
	ErrVettingInProgress = errors.New("a vetting job is already in progress for this third party")
)

type (
	Vetter interface {
		Assess(
			ctx context.Context,
			websiteURL string,
			procedure string,
			reporter agent.ProgressReporter,
			extraTools []agent.Tool,
		) (*vetting.Result, error)
	}

	DisabledVetter struct{}

	VetRequest struct {
		ID         gid.GID
		WebsiteURL string
		Procedure  *string
	}
)

var _ Vetter = DisabledVetter{}

func (DisabledVetter) Assess(
	_ context.Context,
	_ string,
	_ string,
	_ agent.ProgressReporter,
	_ []agent.Tool,
) (*vetting.Result, error) {
	return nil, ErrVettingDisabled
}

func (req VetRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(req.WebsiteURL, "website_url", validator.Required(), validator.SafeText(vettingWebsiteURLMaxLength))
	v.Check(req.Procedure, "procedure", validator.SafeText(vettingProcedureMaxLength))

	return v.Error()
}

func sanitizeVettingError(err error) string {
	msg := err.Error()
	if len(msg) <= vettingErrorMessageMaxLen {
		return msg
	}

	cut := vettingErrorMessageMaxLen
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}

	return msg[:cut] + "…"
}

func (s *Service) Vet(
	ctx context.Context,
	predicate coredata.Predicater,
	req VetRequest,
) (*coredata.ThirdParty, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if !s.vettingEnabled {
		return nil, ErrVettingDisabled
	}

	thirdParty := &coredata.ThirdParty{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByIDForUpdate(ctx, conn, predicate, req.ID); err != nil {
				return fmt.Errorf("cannot load thirdParty %q: %w", req.ID, err)
			}

			if thirdParty.VettingStatus != nil && thirdParty.VettingStatus.IsActive() {
				return ErrVettingInProgress
			}

			pending := coredata.ThirdPartyVettingStatusPending
			websiteURL := req.WebsiteURL

			thirdParty.VettingStatus = &pending
			thirdParty.VettingWebsiteURL = &websiteURL
			thirdParty.VettingProcedure = req.Procedure
			thirdParty.VettingProcessingStartedAt = nil
			thirdParty.VettingErrorMessage = nil
			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, conn, predicate); err != nil {
				return fmt.Errorf("cannot enqueue vetting: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s *Service) VettingStatus(
	ctx context.Context,
	predicate coredata.Predicater,
	thirdPartyID gid.GID,
) (*coredata.ThirdPartyVettingStatus, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParty.LoadByID(ctx, conn, predicate, thirdPartyID)
		},
	)
	if err != nil {
		return nil, err
	}

	if thirdParty.VettingStatus == nil {
		return nil, nil
	}

	return thirdParty.VettingStatus, nil
}
