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

package cookiebanner

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdResetTrackers(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBanner      string
		flagOrg         string
		flagMappingOnly bool
		flagDryRun      bool
		flagYes         bool
	)

	cmd := &cobra.Command{
		Use:   "reset-trackers",
		Short: "Rebuild a banner's tracker patterns from detections and re-arm the analysis + mapping workers",
		Long: "Destructive, tenant-scoped operator action. For a banner's uncategorised, " +
			"non-excluded patterns it clears catalog/vendor links, rebuilds the raw exact " +
			"patterns from detected_trackers (decomposing derived globs), and re-arms the " +
			"pattern-analysis and mapping workers. User-categorised and excluded patterns are " +
			"preserved. With --mapping-only it skips the rebuild and only re-arms mapping.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&flagBanner, "banner", "", "Cookie banner GID to reset")
	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization GID: reset every banner of the org")
	cmd.Flags().BoolVar(&flagMappingOnly, "mapping-only", false, "Only re-arm mapping (skip the detection rebuild and analysis)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the target banners without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if (flagBanner == "") == (flagOrg == "") {
			return fmt.Errorf("exactly one of --banner or --org is required")
		}

		ctx := cmd.Context()

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		bannerIDs, scope, err := resolveTargetBanners(ctx, pgClient, flagBanner, flagOrg)
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		if len(bannerIDs) == 0 {
			_, _ = fmt.Fprintln(out, "No cookie banners matched.")
			return nil
		}

		mode := "full reset"
		if flagMappingOnly {
			mode = "mapping-only reset"
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would run %s on %d banner(s):\n", mode, len(bannerIDs))
			for _, id := range bannerIDs {
				_, _ = fmt.Fprintf(out, "  %s\n", id.String())
			}

			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to run %s on %d banner(s); pass --yes to proceed or --dry-run to preview", mode, len(bannerIDs))
		}

		for _, id := range bannerIDs {
			result, err := cookiebanner.ResetBannerTrackers(ctx, pgClient, scope, id, flagMappingOnly)
			if err != nil {
				return fmt.Errorf("cannot reset banner %s: %w", id, err)
			}

			_, _ = fmt.Fprintf(
				out,
				"%s: reset %d pattern(s), decomposed %d glob(s) into %d exact(s), relinked %d detection(s), analysis_requested=%t\n",
				id.String(),
				result.PatternsReset,
				result.GlobsDecomposed,
				result.ExactsCreated,
				result.DetectionsRelinked,
				result.AnalysisRequested,
			)
		}

		return nil
	}

	return cmd
}

// resolveTargetBanners returns the banner ids to reset and a tenant scope
// derived from the provided GID. The scope is keyed off the banner or org
// GID so every downstream write stays tenant-isolated.
func resolveTargetBanners(
	ctx context.Context,
	pgClient *pg.Client,
	bannerFlag, orgFlag string,
) ([]gid.GID, coredata.Scoper, error) {
	if bannerFlag != "" {
		id, err := gid.ParseGID(bannerFlag)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --banner GID %q: %w", bannerFlag, err)
		}

		return []gid.GID{id}, coredata.NewScopeFromObjectID(id), nil
	}

	orgID, err := gid.ParseGID(orgFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --org GID %q: %w", orgFlag, err)
	}

	scope := coredata.NewScopeFromObjectID(orgID)

	var ids []gid.GID

	if err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			banners, err := cmdutil.Paginate(
				ctx,
				page.OrderBy[coredata.CookieBannerOrderField]{
					Field:     coredata.CookieBannerOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				0,
				func(ctx context.Context, cursor *page.Cursor[coredata.CookieBannerOrderField]) ([]*coredata.CookieBanner, error) {
					var bs coredata.CookieBanners
					if err := bs.LoadByOrganizationID(ctx, conn, scope, orgID, cursor, coredata.NewCookieBannerFilter(nil)); err != nil {
						return nil, err
					}

					return bs, nil
				},
			)
			if err != nil {
				return err
			}

			for _, b := range banners {
				ids = append(ids, b.ID)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return ids, scope, nil
}
