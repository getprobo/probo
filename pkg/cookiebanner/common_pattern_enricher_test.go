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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

// TestInterpretEnrichmentAttribution pins the enricher's post-agent
// decision, which used to accept a vendor on name and confidence alone.
//
// Two cases carry the bug this replaced. A rejected attribution that also
// carries an explicit first-party verdict must still yield that terminal
// verdict — the enricher previously ignored IsFirstParty entirely. And a
// denylisted library name without a first-party verdict must yield nothing
// rather than a synthesized first-party verdict, because the guards judge
// the proposed name and a wrong terminal verdict is unrecoverable.
func TestInterpretEnrichmentAttribution(t *testing.T) {
	t.Parallel()

	base := func(mut func(*TrackerMappingAgentResult)) TrackerMappingAgentResult {
		r := TrackerMappingAgentResult{
			ThirdPartyName:       "PostHog",
			Category:             coredata.ThirdPartyCategoryAnalytics,
			ThirdPartyConfidence: 0.9,
			EvidenceSource:       evidenceSourceNamingConvention,
		}
		if mut != nil {
			mut(&r)
		}

		return r
	}

	tests := []struct {
		name              string
		mutate            func(*TrackerMappingAgentResult)
		expectVendor      string
		expectFirstParty  bool
		expectNil         bool
		expectedRejection attributionRejection
	}{
		{
			name:              "accepts a confident evidence-backed vendor",
			expectVendor:      "PostHog",
			expectedRejection: attributionAccepted,
		},
		{
			name:              "trims the vendor name",
			mutate:            func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "  PostHog  " },
			expectVendor:      "PostHog",
			expectedRejection: attributionAccepted,
		},
		{
			name:              "below the confidence threshold yields nothing",
			mutate:            func(r *TrackerMappingAgentResult) { r.ThirdPartyConfidence = 0.4 },
			expectNil:         true,
			expectedRejection: attributionRejectedConfidence,
		},
		{
			name: "no evidence yields nothing",
			mutate: func(r *TrackerMappingAgentResult) {
				r.ThirdPartyName = "Acme"
				r.EvidenceSource = evidenceSourceNone
			},
			expectNil:         true,
			expectedRejection: attributionRejectedNoEvidence,
		},
		{
			name: "a terminal flag outranks a missing evidence source",
			mutate: func(r *TrackerMappingAgentResult) {
				r.ThirdPartyName = "Acme"
				r.EvidenceSource = evidenceSourceNone
				r.IsFirstParty = true
			},
			expectFirstParty:  true,
			expectedRejection: attributionRejectedTerminalVerdict,
		},
		{
			name: "a terminal flag outranks the name guards",
			mutate: func(r *TrackerMappingAgentResult) {
				r.ThirdPartyName = "Cookiepedia"
				r.ThirdPartyConfidence = 0.95
				r.IsFirstParty = true
			},
			expectFirstParty:  true,
			expectedRejection: attributionRejectedTerminalVerdict,
		},
		{
			name: "rejected name without a first-party verdict yields nothing",
			mutate: func(r *TrackerMappingAgentResult) {
				r.ThirdPartyName = "Cookiepedia"
				r.ThirdPartyConfidence = 0.95
			},
			expectNil:         true,
			expectedRejection: attributionRejectedAggregator,
		},
		{
			name:              "cookie-database aggregator yields nothing",
			mutate:            func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "Cookiepedia" },
			expectNil:         true,
			expectedRejection: attributionRejectedAggregator,
		},
		{
			// A catalog pattern has no scanned site, so the scanned-site
			// backstop cannot fire and a vendor whose name resembles some
			// domain is still attributed.
			name:              "scanned-site backstop cannot fire on a catalog pattern",
			mutate:            func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "Example" },
			expectVendor:      "Example",
			expectedRejection: attributionAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				ident, rejection := interpretEnrichmentAttribution(base(tt.mutate))

				assert.Equal(t, tt.expectedRejection, rejection)

				if tt.expectNil {
					assert.Nil(t, ident)
					return
				}

				require.NotNil(t, ident)
				assert.Equal(t, tt.expectFirstParty, ident.firstParty)

				if tt.expectVendor != "" {
					assert.Equal(t, tt.expectVendor, ident.result.ThirdPartyName)
				}
			},
		)
	}
}

// TestPersistFirstPartyVerdict_DoesNotRequeue pins the terminal state of a
// first-party verdict, and above all that the stale-recovery sweep never
// re-queues it.
//
// The sweep re-arms rows that were claimed but carry no enrichment
// payload. A first-party verdict writes no description and links no
// vendor, so if it also skipped the payload the row would be re-queued on
// every sweep and re-run the agent forever.
func TestPersistFirstPartyVerdict_DoesNotRequeue(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedEnricherCommonThirdParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	requested := now.Add(-time.Hour)

	// Stage a claimed row: an attempt recorded, a vendor still linked, and
	// a description that names it. All three must be undone.
	cp := coredata.CommonTrackerPattern{
		ID:                      gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID:      &party.ID,
		TrackerType:             coredata.TrackerTypeLocalStorage,
		Pattern:                 "i18nextLng-" + party.Slug,
		MatchType:               coredata.TrackerPatternMatchTypeExact,
		Description:             "Stores the visitor's language for Acme Analytics.",
		Confidence:              0.8,
		Attribution:             coredata.CommonTrackerPatternAttributionThirdParty,
		EnrichmentRequestedAt:   nil,
		EnrichmentAttempts:      1,
		LastEnrichmentAttemptAt: &requested,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cp.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, cp.ID)
			return err
		})
	})

	enricher := NewCommonPatternEnricher(
		client,
		log.NewLogger(log.WithOutput(io.Discard)),
		TrackerEnrichmentAgentConfig{Model: "model-x"},
		TrackerMappingAgentConfig{},
	)

	require.NoError(t, enricher.persistFirstPartyVerdict(
		ctx,
		cp,
		&agentIdentification{firstParty: true},
	))

	var stored coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return stored.LoadByID(ctx, conn, cp.ID)
	}))

	assert.Equal(t, coredata.CommonTrackerPatternAttributionFirstParty, stored.Attribution)
	assert.Nil(t, stored.CommonThirdPartyID, "a first-party row must carry no vendor link")
	assert.Empty(t, stored.Description, "a terminal non-vendor row keeps no vendor-naming prose")
	assert.Nil(t, stored.EnrichmentRequestedAt)
	require.NotEmpty(t, stored.Enrichment, "the payload is what stops the stale sweep re-queueing")

	// Back-date the attempt clock well past the staleness window used
	// below, so the sweep's eligibility does not hinge on clock skew
	// between this process and the database.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE common_tracker_patterns SET last_enrichment_attempt_at = NOW() - interval '1 hour' WHERE id = $1`,
			cp.ID,
		)

		return err
	}))

	// The fence: this sweep re-queues every claimed row that carries no
	// payload, so a first-party verdict that skipped the payload would be
	// re-armed here and loop forever.
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.ResetStaleEnrichments(ctx, conn, time.Minute, 10)
	}))

	var afterSweep coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return afterSweep.LoadByID(ctx, conn, cp.ID)
	}))

	assert.Nil(t, afterSweep.EnrichmentRequestedAt, "first-party row must not be re-queued")
	assert.Equal(t, coredata.CommonTrackerPatternAttributionFirstParty, afterSweep.Attribution)
}

// seedEnricherCommonThirdParty inserts a catalog vendor with a
// collision-free name and slug. The catalog is global and uniquely indexed
// on slug, so parallel tests must namespace their rows.
func seedEnricherCommonThirdParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	suffix := slug.Make(id.String())

	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Acme Analytics " + suffix,
		Slug:           "acme-analytics-" + suffix,
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		Review:         coredata.CommonThirdPartyReviewUnreviewed,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
			return err
		})
	})

	return party
}

// TestBuildCommonPatternTerminalMetadata pins that a terminal verdict reads as
// a fully resolved run, and that the payload records which verdict applied.
// Both enrichment targets are settled — the artifact has no vendor to name and
// so no vendor-informed description to write — so reporting no_result would
// misrepresent a definitive answer as a failed one. And flattening the verdict
// to "first party" would claim an extension is the operator's own code.
func TestBuildCommonPatternTerminalMetadata(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)

	for _, verdict := range []coredata.CommonTrackerPatternAttribution{
		coredata.CommonTrackerPatternAttributionFirstParty,
		coredata.CommonTrackerPatternAttributionNotAttributable,
	} {
		t.Run(string(verdict), func(t *testing.T) {
			t.Parallel()

			meta := buildCommonPatternTerminalMetadata("model-x", verdict, now)

			assert.Equal(t, commonPatternStatusDone, meta.Status)
			assert.Equal(t, "model-x", meta.Model)

			require.NotNil(t, meta.Attribution)
			assert.Equal(t, string(verdict), meta.Attribution.TerminalVerdict)
			assert.Empty(t, meta.Attribution.ThirdPartyName)

			require.Len(t, meta.Fields, 2)

			for _, field := range []string{commonPatternFieldDescription, commonPatternFieldThirdParty} {
				assert.Equal(t, commonPatternFieldStatusTerminal, meta.Fields[field].Status, field)
				assert.True(t, commonPatternFieldResolved(meta.Fields[field].Status), field)
			}
		})
	}
}
