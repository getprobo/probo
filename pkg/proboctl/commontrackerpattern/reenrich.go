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

package commontrackerpattern

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdReenrich(f *cmdutil.Factory) *cobra.Command {
	var (
		flagIDs                []string
		flagLinkedBanner       string
		flagLinkedOrg          string
		flagCommonThirdParty   string
		flagTrackerType        string
		flagKeyword            string
		flagState              string
		flagWithoutDescription bool
		flagDryRun             bool
		flagYes                bool
	)

	cmd := &cobra.Command{
		Use:   "reenrich",
		Short: "Re-describe common tracker patterns via the enrichment worker",
		Long: "Re-describe selected common tracker patterns by arming the async " +
			"enrichment worker, which fills descriptions and fans them out to linked " +
			"tracker patterns. Re-describe a banner's catalog rows with --linked-banner " +
			"before running 'cookie-banner reset-trackers' so fresh descriptions copy down.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringSliceVar(&flagIDs, "id", nil, "Common tracker pattern GID(s) to re-enrich (repeatable)")
	cmd.Flags().StringVar(&flagLinkedBanner, "linked-banner", "", "Select catalog rows linked to a cookie banner's patterns (GID)")
	cmd.Flags().StringVar(&flagLinkedOrg, "linked-org", "", "Select catalog rows linked to an organization's patterns (GID)")
	cmd.Flags().StringVar(&flagCommonThirdParty, "common-third-party", "", "Select patterns linked to a common third party (slug or GID)")
	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Filter selected patterns by tracker type")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter selected patterns by a pattern/description substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter selected patterns by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().BoolVar(&flagWithoutDescription, "without-description", false, "Only patterns with a blank description")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the selected patterns without enriching")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		ids, err := resolveReenrichIDs(
			ctx,
			pgClient,
			flagIDs,
			flagLinkedBanner,
			flagLinkedOrg,
			flagCommonThirdParty,
			flagTrackerType,
			flagKeyword,
			flagState,
			flagWithoutDescription,
		)
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		if len(ids) == 0 {
			_, _ = fmt.Fprintln(out, "No common tracker patterns matched the selection.")
			return nil
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would re-enrich %d common tracker pattern(s).\n", len(ids))
			printSample(out, ids)

			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to re-enrich %d pattern(s); pass --yes to proceed or --dry-run to preview", len(ids))
		}

		var requeued int64

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var ps coredata.CommonTrackerPatterns

				requeued, err = ps.RequestEnrichmentByIDs(ctx, tx, ids)

				return err
			},
		); err != nil {
			return fmt.Errorf("cannot enqueue enrichment: %w", err)
		}

		_, _ = fmt.Fprintf(out, "Queued %d common tracker pattern(s) for the enrichment worker.\n", requeued)

		return nil
	}

	return cmd
}

// resolveReenrichIDs turns the selection flags into the set of common
// tracker pattern IDs to re-enrich. At most one selection anchor may be
// provided: --id, --linked-banner, --linked-org, or --common-third-party.
// The --tracker-type, --keyword, --state, and --without-description flags
// further narrow the anchor's result, except with --id, where the listed
// patterns are used verbatim. With no anchor, the filtering flags select
// across the whole catalog (e.g. --without-description re-enriches every
// pattern with a blank description).
func resolveReenrichIDs(
	ctx context.Context,
	pgClient *pg.Client,
	rawIDs []string,
	linkedBanner, linkedOrg, commonThirdParty string,
	trackerType, keyword, state string,
	withoutDescription bool,
) ([]gid.GID, error) {
	anchors := 0

	for _, set := range []bool{len(rawIDs) > 0, linkedBanner != "", linkedOrg != "", commonThirdParty != ""} {
		if set {
			anchors++
		}
	}

	if anchors > 1 {
		return nil, fmt.Errorf("--id, --linked-banner, --linked-org, and --common-third-party are mutually exclusive")
	}

	// --id selects patterns verbatim; the filtering flags do not apply.
	if len(rawIDs) > 0 {
		ids := make([]gid.GID, 0, len(rawIDs))

		for _, raw := range rawIDs {
			id, err := gid.ParseGID(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid --id value %q: %w", raw, err)
			}

			ids = append(ids, id)
		}

		return ids, nil
	}

	var ids []gid.GID

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			filter, err := buildReenrichFilter(trackerType, keyword, state, withoutDescription)
			if err != nil {
				return err
			}

			switch {
			case linkedBanner != "":
				bannerID, err := gid.ParseGID(linkedBanner)
				if err != nil {
					return fmt.Errorf("invalid --linked-banner GID %q: %w", linkedBanner, err)
				}

				var tps coredata.TrackerPatterns

				linkedIDs, err := tps.LoadAllLinkedCommonTrackerPatternIDsByCookieBannerID(ctx, conn, coredata.NewScopeFromObjectID(bannerID), bannerID)
				if err != nil {
					return err
				}

				if len(linkedIDs) == 0 {
					return nil
				}

				filter.WithIDs(linkedIDs)
			case linkedOrg != "":
				orgID, err := gid.ParseGID(linkedOrg)
				if err != nil {
					return fmt.Errorf("invalid --linked-org GID %q: %w", linkedOrg, err)
				}

				var tps coredata.TrackerPatterns

				linkedIDs, err := tps.LoadAllLinkedCommonTrackerPatternIDsByOrganizationID(ctx, conn, coredata.NewScopeFromObjectID(orgID), orgID)
				if err != nil {
					return err
				}

				if len(linkedIDs) == 0 {
					return nil
				}

				filter.WithIDs(linkedIDs)
			case commonThirdParty != "":
				thirdPartyID, err := resolveCommonThirdPartyID(ctx, conn, commonThirdParty)
				if err != nil {
					return err
				}

				filter.WithCommonThirdPartyID(&thirdPartyID)
			}

			var ps coredata.CommonTrackerPatterns

			ids, err = ps.LoadAllIDs(ctx, conn, filter)

			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func buildReenrichFilter(trackerType, keyword, state string, withoutDescription bool) (*coredata.CommonTrackerPatternFilter, error) {
	filter := coredata.NewCommonTrackerPatternFilter()

	if trackerType != "" {
		tt := coredata.TrackerType(trackerType)
		if !tt.IsValid() {
			return nil, fmt.Errorf("invalid --tracker-type value %q", trackerType)
		}

		filter.WithTrackerType(&tt)
	}

	if keyword != "" {
		filter.WithKeyword(&keyword)
	}

	if state != "" {
		st, err := parseEnrichmentState(state)
		if err != nil {
			return nil, err
		}

		filter.WithState(&st)
	}

	if withoutDescription {
		filter.WithDescribed(new(false))
	}

	return filter, nil
}

func printSample(out io.Writer, ids []gid.GID) {
	const sampleSize = 10

	for i, id := range ids {
		if i >= sampleSize {
			_, _ = fmt.Fprintf(out, "  ... and %d more\n", len(ids)-sampleSize)
			break
		}

		_, _ = fmt.Fprintf(out, "  %s\n", id.String())
	}
}
