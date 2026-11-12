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

package cookiebanner

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const analyticsFlushInterval = 10 * time.Second

type (
	pageLoadKey struct {
		bannerID      gid.GID
		day           time.Time
		family        coredata.BrowserFamily
		presentation  coredata.CookieBannerPresentation
		consentAction coredata.CookieConsentAction
	}

	patternHitKey struct {
		bannerID        gid.GID
		patternID       gid.GID
		day             time.Time
		family          coredata.BrowserFamily
		presentation    coredata.CookieBannerPresentation
		consentAction   coredata.CookieConsentAction
		categoryGranted bool
	}
)

func (s *Service) addPageLoad(key pageLoadKey) {
	actual, _ := s.pageLoadCounts.LoadOrStore(key, &atomic.Int64{})
	actual.(*atomic.Int64).Add(1)
}

func (s *Service) addPatternHit(key patternHitKey) {
	actual, _ := s.patternHitCounts.LoadOrStore(key, &atomic.Int64{})
	actual.(*atomic.Int64).Add(1)
}

// pageLoadGranted reports whether non-necessary trackers may fire on a view
// with this presentation and consent action. An empty presentation is a
// legacy discovery bucket and counts as granted.
func pageLoadGranted(presentation coredata.CookieBannerPresentation, action *coredata.CookieConsentAction) bool {
	if presentation == "" {
		return true
	}

	actionValue := ""
	if action != nil {
		actionValue = action.String()
	}

	switch presentation {
	case coredata.CookieBannerPresentationOptIn:
		return actionValue == coredata.CookieConsentActionAcceptAll.String()
	case coredata.CookieBannerPresentationOptOut, coredata.CookieBannerPresentationNotice:
		return actionValue != coredata.CookieConsentActionRejectAll.String() &&
			actionValue != coredata.CookieConsentActionGPC.String()
	default:
		return false
	}
}

// categoryGrantedForHit reports whether this hit's category was allowed to
// fire. CUSTOMIZE uses the SDK flag. A missing flag on a customize view is
// not granted. Every other action is derived from the presentation.
func categoryGrantedForHit(
	presentation coredata.CookieBannerPresentation,
	action *coredata.CookieConsentAction,
	reported *bool,
) bool {
	if action != nil && *action == coredata.CookieConsentActionCustomize {
		return reported != nil && *reported
	}

	return pageLoadGranted(presentation, action)
}

func consentActionValue(action *coredata.CookieConsentAction) coredata.CookieConsentAction {
	if action == nil {
		return ""
	}

	return *action
}

func (s *Service) recordAnalytics(
	bannerID gid.GID,
	family coredata.BrowserFamily,
	presentation coredata.CookieBannerPresentation,
	action *coredata.CookieConsentAction,
	pageView bool,
	patternGrants map[gid.GID]bool,
) {
	if !family.IsValid() {
		family = coredata.BrowserFamilyOther
	}

	day := utcDay(s.now())
	consentAction := consentActionValue(action)

	if pageView {
		s.addPageLoad(pageLoadKey{
			bannerID:      bannerID,
			day:           day,
			family:        family,
			presentation:  presentation,
			consentAction: consentAction,
		})
	}

	for patternID, granted := range patternGrants {
		s.addPatternHit(patternHitKey{
			bannerID:        bannerID,
			patternID:       patternID,
			day:             day,
			family:          family,
			presentation:    presentation,
			consentAction:   consentAction,
			categoryGranted: granted,
		})
	}
}

// FlushAnalytics writes the buffered counts to their day buckets. A
// counter of a past day is dropped once a flush finds it empty: increments
// pick their day right before adding, so none can still be in flight a
// whole flush interval after midnight.
func (s *Service) FlushAnalytics(ctx context.Context) error {
	var firstErr error

	today := utcDay(s.now())

	s.pageLoadCounts.Range(
		func(key, value any) bool {
			loadKey := key.(pageLoadKey)

			count := value.(*atomic.Int64).Swap(0)
			if count == 0 {
				if loadKey.day.Before(today) {
					s.pageLoadCounts.CompareAndDelete(key, value)
				}

				return true
			}

			if err := s.writePageLoad(ctx, loadKey, count); err != nil && firstErr == nil {
				firstErr = err
			}

			return true
		},
	)

	s.patternHitCounts.Range(
		func(key, value any) bool {
			hitKey := key.(patternHitKey)

			count := value.(*atomic.Int64).Swap(0)
			if count == 0 {
				if hitKey.day.Before(today) {
					s.patternHitCounts.CompareAndDelete(key, value)
				}

				return true
			}

			if err := s.writePatternHit(ctx, hitKey, count); err != nil && firstErr == nil {
				firstErr = err
			}

			return true
		},
	)

	return firstErr
}

func (s *Service) writePageLoad(
	ctx context.Context,
	key pageLoadKey,
	count int64,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.RecordDailyPageLoad(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(key.bannerID),
				key.bannerID,
				key.day,
				key.family,
				key.presentation,
				key.consentAction,
				count,
			); err != nil {
				return fmt.Errorf("cannot record page load: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) writePatternHit(
	ctx context.Context,
	key patternHitKey,
	count int64,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.RecordDailyPatternHit(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(key.bannerID),
				key.bannerID,
				key.patternID,
				key.day,
				key.family,
				key.presentation,
				key.consentAction,
				key.categoryGranted,
				count,
			); err != nil {
				return fmt.Errorf("cannot record pattern hit: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) RunAnalyticsFlusher(ctx context.Context, logger *log.Logger) error {
	logger = logger.Named("analytics-flusher")

	ticker := time.NewTicker(analyticsFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			flushCtx := context.WithoutCancel(ctx)
			if err := s.FlushAnalytics(flushCtx); err != nil {
				logger.ErrorCtx(flushCtx, "cannot flush analytics on shutdown", log.Error(err))
			}

			return nil
		case <-ticker.C:
			if err := s.FlushAnalytics(ctx); err != nil {
				logger.ErrorCtx(ctx, "cannot flush analytics", log.Error(err))
			}
		}
	}
}
