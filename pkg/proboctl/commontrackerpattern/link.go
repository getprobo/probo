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
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdLink(f *cmdutil.Factory) *cobra.Command {
	var (
		flagIDs                []string
		flagLinkedBanner       string
		flagLinkedOrg          string
		flagCommonThirdParty   string
		flagTrackerType        string
		flagKeyword            string
		flagState              string
		flagWithoutDescription bool
		flagTo                 string
		flagDryRun             bool
		flagYes                bool
	)

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link common tracker patterns to a common third party",
		Long: "Point selected common tracker patterns at a common third party " +
			"(--to-common-third-party). The catalog rows are re-armed for enrichment " +
			"so a description is re-derived for the vendor, and the uncategorised org " +
			"tracker patterns linked to them are remapped (org third party cleared, " +
			"mapping re-armed) so the mapping worker re-resolves the vendor. " +
			"User-categorised and excluded org patterns are left untouched. Selection " +
			"mirrors 'reenrich'. To detach patterns, use 'unlink'.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringSliceVar(&flagIDs, "id", nil, "Common tracker pattern GID(s) to link (repeatable)")
	cmd.Flags().StringVar(&flagLinkedBanner, "linked-banner", "", "Select catalog rows linked to a cookie banner's patterns (GID)")
	cmd.Flags().StringVar(&flagLinkedOrg, "linked-org", "", "Select catalog rows linked to an organization's patterns (GID)")
	cmd.Flags().StringVar(&flagCommonThirdParty, "common-third-party", "", "Select patterns currently linked to a common third party (slug or GID)")
	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Filter selected patterns by tracker type")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter selected patterns by a pattern/description substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter selected patterns by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().BoolVar(&flagWithoutDescription, "without-description", false, "Only patterns with a blank description")
	cmd.Flags().StringVar(&flagTo, "to-common-third-party", "", "Target common third party to link to (slug or GID)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the selected patterns without linking")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	_ = cmd.MarkFlagRequired("to-common-third-party")

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
			_, _ = fmt.Fprintf(out, "Would link %d common tracker pattern(s) to %s.\n", len(ids), flagTo)
			printSample(out, ids)

			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to link %d pattern(s) to %s; pass --yes to proceed or --dry-run to preview", len(ids), flagTo)
		}

		var (
			linked   int64
			requeued int64
			remapped int64
		)

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				thirdPartyID, err := resolveCommonThirdPartyID(ctx, tx, flagTo)
				if err != nil {
					return err
				}

				var ps coredata.CommonTrackerPatterns

				linked, err = ps.RelinkCommonThirdPartyByIDs(ctx, tx, ids, &thirdPartyID)
				if err != nil {
					return err
				}

				requeued, err = ps.RequestEnrichmentByIDs(ctx, tx, ids)
				if err != nil {
					return err
				}

				var tps coredata.TrackerPatterns

				remapped, err = tps.RequestMappingForUncategorisedByCommonTrackerPatternIDs(ctx, tx, ids)
				if err != nil {
					return err
				}

				return nil
			},
		); err != nil {
			return fmt.Errorf("cannot link common tracker patterns: %w", err)
		}

		_, _ = fmt.Fprintf(
			out,
			"Linked %d pattern(s) to %s, re-queued %d for enrichment, remapped %d uncategorised org tracker pattern(s).\n",
			linked,
			flagTo,
			requeued,
			remapped,
		)

		return nil
	}

	return cmd
}
