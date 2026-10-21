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
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const sampleCountFlushInterval = 10 * time.Second

type (
	sampleLoadKey struct {
		bannerID      gid.GID
		family        coredata.BrowserFamily
		presentation  string
		consentAction string
	}

	sampleHitKey struct {
		bannerID        gid.GID
		patternID       gid.GID
		family          coredata.BrowserFamily
		presentation    string
		consentAction   string
		categoryGranted bool
	}
)

func (s *Service) addPageLoad(key sampleLoadKey) {
	actual, _ := s.pageLoadCounts.LoadOrStore(key, &atomic.Int64{})
	actual.(*atomic.Int64).Add(1)
}

func (s *Service) addPatternHit(key sampleHitKey) {
	actual, _ := s.patternHitCounts.LoadOrStore(key, &atomic.Int64{})
	actual.(*atomic.Int64).Add(1)
}

// pageLoadGranted reports whether non-necessary trackers may fire on a view
// with this presentation and consent action. An empty presentation is a
// legacy discovery bucket and counts as granted.
func pageLoadGranted(presentation Presentation, action *coredata.CookieConsentAction) bool {
	if presentation == "" {
		return true
	}

	actionValue := ""
	if action != nil {
		actionValue = action.String()
	}

	switch presentation {
	case PresentationOptIn:
		return actionValue == coredata.CookieConsentActionAcceptAll.String()
	case PresentationOptOut, PresentationNotice:
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
	presentation Presentation,
	action *coredata.CookieConsentAction,
	reported *bool,
) bool {
	if action != nil && *action == coredata.CookieConsentActionCustomize {
		return reported != nil && *reported
	}

	return pageLoadGranted(presentation, action)
}

func consentActionValue(action *coredata.CookieConsentAction) string {
	if action == nil {
		return ""
	}

	return action.String()
}

func (s *Service) recordSampleIncrements(
	bannerID gid.GID,
	family coredata.BrowserFamily,
	presentation Presentation,
	action *coredata.CookieConsentAction,
	pageView bool,
	patternGrants map[gid.GID]bool,
) {
	if !family.IsValid() {
		family = coredata.BrowserFamilyOther
	}

	consentAction := consentActionValue(action)

	if pageView {
		s.addPageLoad(sampleLoadKey{
			bannerID:      bannerID,
			family:        family,
			presentation:  string(presentation),
			consentAction: consentAction,
		})
	}

	for patternID, granted := range patternGrants {
		s.addPatternHit(sampleHitKey{
			bannerID:        bannerID,
			patternID:       patternID,
			family:          family,
			presentation:    string(presentation),
			consentAction:   consentAction,
			categoryGranted: granted,
		})
	}
}

func (s *Service) FlushSampleCounts(ctx context.Context) error {
	var firstErr error

	s.pageLoadCounts.Range(
		func(key, value any) bool {
			loadKey := key.(sampleLoadKey)

			count := value.(*atomic.Int64).Swap(0)
			if count == 0 {
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
			hitKey := key.(sampleHitKey)

			count := value.(*atomic.Int64).Swap(0)
			if count == 0 {
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
	key sampleLoadKey,
	count int64,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(key.bannerID)

			sampleID, ok, err := resolveWritableSample(ctx, tx, scope, key.bannerID)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			if err := coredata.RecordSamplePageLoad(
				ctx,
				tx,
				scope,
				sampleID,
				key.family,
				key.presentation,
				key.consentAction,
				count,
			); err != nil {
				return fmt.Errorf("cannot record sample page load: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) writePatternHit(
	ctx context.Context,
	key sampleHitKey,
	count int64,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(key.bannerID)

			sampleID, ok, err := resolveWritableSample(ctx, tx, scope, key.bannerID)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			if err := coredata.RecordSamplePatternHit(
				ctx,
				tx,
				scope,
				sampleID,
				key.bannerID,
				key.patternID,
				key.family,
				key.presentation,
				key.consentAction,
				key.categoryGranted,
				count,
			); err != nil {
				return fmt.Errorf("cannot record sample pattern hit: %w", err)
			}

			return nil
		},
	)
}

func resolveWritableSample(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	bannerID gid.GID,
) (gid.GID, bool, error) {
	var open coredata.CookieBannerSample

	err := open.LoadOpenByCookieBannerID(ctx, tx, scope, bannerID)
	if err == nil {
		return open.ID, true, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return gid.GID{}, false, fmt.Errorf("cannot load open cookie banner sample: %w", err)
	}

	var samples coredata.CookieBannerSamples

	count, err := samples.CountByCookieBannerID(ctx, tx, scope, bannerID)
	if err != nil {
		return gid.GID{}, false, err
	}

	if count > 0 {
		return gid.GID{}, false, nil
	}

	var banner coredata.CookieBanner
	if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
		return gid.GID{}, false, fmt.Errorf("cannot load cookie banner: %w", err)
	}

	now := time.Now()

	sample := coredata.CookieBannerSample{
		ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerSampleEntityType),
		OrganizationID: banner.OrganizationID,
		CookieBannerID: bannerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	inserted, err := sample.InsertOpen(ctx, tx, scope)
	if err != nil {
		return gid.GID{}, false, err
	}

	if !inserted {
		if err := open.LoadOpenByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
			return gid.GID{}, false, fmt.Errorf("cannot reload open cookie banner sample: %w", err)
		}

		return open.ID, true, nil
	}

	return sample.ID, true, nil
}

func (s *Service) RunSampleCountFlusher(ctx context.Context, logger *log.Logger) error {
	logger = logger.Named("sample-count-flusher")

	ticker := time.NewTicker(sampleCountFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			flushCtx := context.WithoutCancel(ctx)
			if err := s.FlushSampleCounts(flushCtx); err != nil {
				logger.ErrorCtx(flushCtx, "cannot flush sample counts on shutdown", log.Error(err))
			}

			return nil
		case <-ticker.C:
			if err := s.FlushSampleCounts(ctx); err != nil {
				logger.ErrorCtx(ctx, "cannot flush sample counts", log.Error(err))
			}
		}
	}
}
