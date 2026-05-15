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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type trackerMappingHandler struct {
	pg     *pg.Client
	logger *log.Logger
}

func NewTrackerMappingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.TrackerPattern] {
	h := &trackerMappingHandler{
		pg:     pgClient,
		logger: logger,
	}

	return worker.New(
		"tracker-mapping-worker",
		h,
		logger,
		opts...,
	)
}

func (h *trackerMappingHandler) Claim(ctx context.Context) (coredata.TrackerPattern, error) {
	var tp coredata.TrackerPattern

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := tp.LoadNextForMappingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return tp.ClearMappingRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.TrackerPattern{}, worker.ErrNoTask
		}
		return coredata.TrackerPattern{}, fmt.Errorf("cannot claim tracker mapping task: %w", err)
	}

	return tp, nil
}

func (h *trackerMappingHandler) Process(ctx context.Context, tp coredata.TrackerPattern) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var commonPatternID *gid.GID
			var thirdPartyID *gid.GID

			commonPatternID, thirdPartyID = h.matchByPattern(ctx, tx, tp)

			if commonPatternID == nil {
				commonPatternID, thirdPartyID = h.matchByDomain(ctx, tx, tp)
			}

			if commonPatternID != nil || thirdPartyID != nil {
				if err := tp.UpdateMapping(ctx, tx, commonPatternID, thirdPartyID); err != nil {
					return fmt.Errorf("cannot update tracker pattern mapping: %w", err)
				}

				h.logger.InfoCtx(
					ctx,
					"mapped tracker pattern",
					log.String("pattern", tp.Pattern),
					log.String("tracker_pattern_id", tp.ID.String()),
				)
			}

			return nil
		},
	)
}

func (h *trackerMappingHandler) matchByPattern(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByPattern(ctx, conn, tp.TrackerType, tp.Pattern, tp.MaxAgeSeconds); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			h.logger.ErrorCtx(ctx, "cannot load common tracker pattern", log.Error(err))
		}
		return nil, nil
	}

	var thirdPartyID *gid.GID
	if commonPattern.CommonThirdPartyID != nil {
		thirdPartyID = h.resolveThirdParty(ctx, conn, tp, &commonPattern)
	}

	return &commonPattern.ID, thirdPartyID
}

func (h *trackerMappingHandler) matchByDomain(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID) {
	var trackers coredata.DetectedTrackers
	commonThirdPartyID, err := trackers.LoadCommonThirdPartyIDByDomainMatch(ctx, tx, tp.ID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot load common third party ID from domain", log.Error(err))
		return nil, nil
	}

	if commonThirdPartyID == nil {
		return nil, nil
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		Description:        tp.Description,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         0.7,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	actualID, _, err := commonPattern.Upsert(ctx, tx)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot upsert common tracker pattern from domain match",
			log.Error(err),
		)
		return nil, nil
	}

	commonPattern.ID = actualID
	thirdPartyID := h.resolveThirdParty(ctx, tx, tp, &commonPattern)

	return &commonPattern.ID, thirdPartyID
}

func (h *trackerMappingHandler) resolveThirdParty(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
	commonPattern *coredata.CommonTrackerPattern,
) *gid.GID {
	if commonPattern.CommonThirdPartyID == nil {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(tp.ID)

	var t coredata.ThirdParty
	if err := t.LoadByOrganizationIDAndCommonThirdPartyID(
		ctx,
		conn,
		scope,
		tp.OrganizationID,
		*commonPattern.CommonThirdPartyID,
	); err != nil {
		return nil
	}

	return &t.ID
}
