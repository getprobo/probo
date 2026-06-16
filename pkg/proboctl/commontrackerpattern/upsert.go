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

package commontrackerpattern

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdUpsert(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTrackerType      string
		flagPattern          string
		flagMatchType        string
		flagDescription      string
		flagMaxAgeSeconds    int
		flagConfidence       float32
		flagAttribution      string
		flagCommonThirdParty string
		flagDryRun           bool
		flagEnrich           bool
	)

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a common tracker pattern in the global catalog",
		Long: "Insert a new common tracker pattern or update an existing one keyed by " +
			"its natural key (--tracker-type, --pattern, --max-age-seconds). Only " +
			"--tracker-type and --pattern are required; every other field is updated " +
			"only when its flag is passed, so an existing row's other columns are " +
			"preserved. An empty --description never overwrites an existing one — " +
			"descriptions are owned by the enrichment worker. A FIRST_PARTY row stays " +
			"vendor-free and keeps its terminal verdict regardless of --common-third-party " +
			"or --attribution.\n\n" +
			"Pass --enrich to queue the row for the async enrichment worker after " +
			"writing, so a minimal pattern row gets its description (and vendor " +
			"attribution) researched. Enrichment is expensive (LLM per row).",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Tracker type (required)")
	cmd.Flags().StringVar(&flagPattern, "pattern", "", "Pattern key (required)")
	cmd.Flags().StringVar(&flagMatchType, "match-type", string(coredata.TrackerPatternMatchTypeExact), "Match type (EXACT, PREFIX, GLOB)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Human-readable description")
	cmd.Flags().IntVar(&flagMaxAgeSeconds, "max-age-seconds", 0, "Cookie max-age in seconds (part of the natural key)")
	cmd.Flags().Float32Var(&flagConfidence, "confidence", 1.0, "Confidence score (default: 1.0 for a manual entry)")
	cmd.Flags().StringVar(&flagAttribution, "attribution", "", "Attribution verdict (UNDETERMINED, THIRD_PARTY, FIRST_PARTY)")
	cmd.Flags().StringVar(&flagCommonThirdParty, "common-third-party", "", "Common third party to link (slug or GID); pass empty to unlink")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the resulting row without writing")
	cmd.Flags().BoolVar(&flagEnrich, "enrich", false, "Queue the row for the async enrichment worker after writing")

	_ = cmd.MarkFlagRequired("tracker-type")
	_ = cmd.MarkFlagRequired("pattern")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		trackerType := coredata.TrackerType(flagTrackerType)
		if !trackerType.IsValid() {
			return fmt.Errorf("invalid --tracker-type value %q", flagTrackerType)
		}

		matchType := coredata.TrackerPatternMatchType(flagMatchType)
		if !matchType.IsValid() {
			return fmt.Errorf("invalid --match-type value %q", flagMatchType)
		}

		if flagPattern == "" {
			return fmt.Errorf("--pattern must not be empty")
		}

		var attribution coredata.CommonTrackerPatternAttribution
		if cmd.Flags().Changed("attribution") {
			attribution = coredata.CommonTrackerPatternAttribution(flagAttribution)
			if !attribution.IsValid() {
				return fmt.Errorf("invalid --attribution value %q", flagAttribution)
			}
		}

		// max_age_seconds is part of the natural key, so an unset flag means
		// "the row with no max age" rather than "leave it untouched".
		var maxAge *int
		if cmd.Flags().Changed("max-age-seconds") {
			maxAge = &flagMaxAgeSeconds
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		var (
			pattern  coredata.CommonTrackerPattern
			inserted bool
		)

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				now := time.Now()

				existing := coredata.CommonTrackerPattern{}

				err := existing.LoadByPattern(ctx, tx, trackerType, flagPattern, maxAge)
				switch {
				case err == nil:
					pattern = existing
				case errors.Is(err, coredata.ErrResourceNotFound):
					pattern = coredata.CommonTrackerPattern{
						ID:         gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
						Confidence: flagConfidence,
						CreatedAt:  now,
					}
				default:
					return fmt.Errorf("cannot load common tracker pattern by pattern: %w", err)
				}

				pattern.TrackerType = trackerType
				pattern.Pattern = flagPattern
				pattern.MatchType = matchType
				pattern.MaxAgeSeconds = maxAge
				pattern.UpdatedAt = now

				if cmd.Flags().Changed("description") {
					pattern.Description = flagDescription
				}

				if cmd.Flags().Changed("confidence") {
					pattern.Confidence = flagConfidence
				}

				if cmd.Flags().Changed("attribution") {
					pattern.Attribution = attribution
				}

				if cmd.Flags().Changed("common-third-party") {
					if flagCommonThirdParty == "" {
						pattern.CommonThirdPartyID = nil

						// Removing the vendor invalidates a THIRD_PARTY
						// verdict, which by definition carries one. When the
						// operator did not set an explicit --attribution,
						// downgrade the now-stale verdict to UNDETERMINED so
						// the mapping pipeline probes the vendor-free row
						// again. A FIRST_PARTY row is already vendor-free and
						// terminal, so it is left untouched.
						if !cmd.Flags().Changed("attribution") &&
							pattern.Attribution == coredata.CommonTrackerPatternAttributionThirdParty {
							pattern.Attribution = coredata.CommonTrackerPatternAttributionUndetermined
						}
					} else {
						thirdPartyID, err := resolveCommonThirdPartyID(ctx, tx, flagCommonThirdParty)
						if err != nil {
							return err
						}

						pattern.CommonThirdPartyID = &thirdPartyID

						// A vendor-linked row is, by definition, attributed to
						// a third party. When the operator did not set an
						// explicit --attribution, normalize the verdict to
						// THIRD_PARTY so the row never persists with an
						// UNDETERMINED (or unset) verdict. A FIRST_PARTY row is
						// terminal and stays vendor-free — the upsert discards
						// the incoming vendor and keeps the verdict — so it is
						// left untouched.
						if !cmd.Flags().Changed("attribution") &&
							pattern.Attribution != coredata.CommonTrackerPatternAttributionFirstParty {
							pattern.Attribution = coredata.CommonTrackerPatternAttributionThirdParty
						}
					}
				}

				if flagDryRun {
					return nil
				}

				inserted, err = pattern.Upsert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot upsert common tracker pattern: %w", err)
				}

				// Arm enrichment explicitly rather than via the receiver:
				// Upsert only queues a description-less row on insert and
				// otherwise leaves the enrichment columns untouched, so this
				// is the only path that re-arms an existing row uniformly. It
				// also resets the attempt budget for a fresh run.
				if flagEnrich {
					var patterns coredata.CommonTrackerPatterns

					if _, err := patterns.RequestEnrichmentByIDs(ctx, tx, []gid.GID{pattern.ID}); err != nil {
						return fmt.Errorf("cannot queue enrichment: %w", err)
					}
				}

				return nil
			},
		); err != nil {
			return err
		}

		enrichSuffix := ""
		if flagEnrich {
			enrichSuffix = " (queued for enrichment)"
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would upsert common tracker pattern %q (tracker type %s, match type %s)%s.\n", pattern.Pattern, pattern.TrackerType, pattern.MatchType, enrichSuffix)
			return nil
		}

		action := "Updated"
		if inserted {
			action = "Created"
		}

		_, _ = fmt.Fprintf(out, "%s common tracker pattern %s (%q, tracker type %s)%s.\n", action, pattern.ID.String(), pattern.Pattern, pattern.TrackerType, enrichSuffix)

		return nil
	}

	return cmd
}
