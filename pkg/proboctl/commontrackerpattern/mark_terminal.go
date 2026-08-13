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

package commontrackerpattern

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// newCmdMarkFirstParty records the terminal FIRST_PARTY verdict: the artifact
// belongs to the site operator.
func newCmdMarkFirstParty(f *cmdutil.Factory) *cobra.Command {
	return newCmdMarkTerminal(
		f,
		"mark-first-party",
		"Mark common tracker patterns as first-party (the operator's own)",
		"Record the terminal FIRST_PARTY verdict: the artifact belongs to the site "+
			"operator - its own tracker, or a bundled library that stores state "+
			"locally and egresses nothing.\n\n",
		coredata.CommonTrackerPatternAttributionFirstParty,
	)
}

// newCmdMarkNotAttributable records the terminal NOT_ATTRIBUTABLE verdict: real
// software belonging to neither the operator nor a vendor they engaged.
func newCmdMarkNotAttributable(f *cmdutil.Factory) *cobra.Command {
	return newCmdMarkTerminal(
		f,
		"mark-not-attributable",
		"Mark common tracker patterns as belonging to nobody's register",
		"Record the terminal NOT_ATTRIBUTABLE verdict: the artifact comes from real "+
			"software that is neither the operator's nor a third party they engaged, "+
			"so it belongs in nobody's register. Browser extensions and other "+
			"visitor-installed tooling inject keys into a page the operator does not "+
			"control.\n\nUse this rather than mark-first-party for extensions: the "+
			"operator did not write that code and cannot remove it, so calling it "+
			"their own would state something false in a privacy register.\n\n",
		coredata.CommonTrackerPatternAttributionNotAttributable,
	)
}

func newCmdMarkTerminal(
	f *cmdutil.Factory,
	use string,
	short string,
	longPrefix string,
	verdict coredata.CommonTrackerPatternAttribution,
) *cobra.Command {
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
		Use:   use,
		Short: short,
		Long: longPrefix +
			"Any vendor link is cleared and the now-stale description - which may name " +
			"the wrong vendor - is blanked on both the catalog row and the uncategorised " +
			"org tracker patterns linked to it. Those org patterns are remapped (org " +
			"third party cleared, mapping re-armed) so the pipeline drops the stale " +
			"vendor; because the verdict is terminal the mapping worker leaves them " +
			"unattributed. User-categorised and excluded org patterns are left " +
			"untouched. Selection mirrors 'reenrich'. To re-attribute a row later, use " +
			"'link' (which returns it to THIRD_PARTY).",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringSliceVar(&flagIDs, "id", nil, "Common tracker pattern GID(s) to mark (repeatable)")
	cmd.Flags().StringVar(&flagLinkedBanner, "linked-banner", "", "Select catalog rows linked to a cookie banner's patterns (GID)")
	cmd.Flags().StringVar(&flagLinkedOrg, "linked-org", "", "Select catalog rows linked to an organization's patterns (GID)")
	cmd.Flags().StringVar(&flagCommonThirdParty, "common-third-party", "", "Select patterns currently linked to a common third party (slug or GID)")
	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Filter selected patterns by tracker type")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter selected patterns by a pattern/description substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter selected patterns by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().BoolVar(&flagWithoutDescription, "without-description", false, "Only patterns with a blank description")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the selected patterns without marking")
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
			_, _ = fmt.Fprintf(out, "Would mark %d common tracker pattern(s) as %s.\n", len(ids), verdict)
			printSample(out, ids)

			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to mark %d pattern(s) as %s; pass --yes to proceed or --dry-run to preview", len(ids), verdict)
		}

		var (
			marked   int64
			remapped int64
			cleared  int64
		)

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var ps coredata.CommonTrackerPatterns

				marked, err = ps.SetAttributionByIDs(ctx, tx, ids, verdict)
				if err != nil {
					return err
				}

				if _, err = ps.ClearDescriptionByIDs(ctx, tx, ids); err != nil {
					return err
				}

				var tps coredata.TrackerPatterns

				remapped, err = tps.RequestMappingForUncategorisedByCommonTrackerPatternIDs(ctx, tx, ids)
				if err != nil {
					return err
				}

				cleared, err = tps.ClearDescriptionForUncategorisedByCommonTrackerPatternIDs(ctx, tx, ids)
				if err != nil {
					return err
				}

				return nil
			},
		); err != nil {
			return fmt.Errorf("cannot mark common tracker patterns %s: %w", verdict, err)
		}

		_, _ = fmt.Fprintf(
			out,
			"Marked %d pattern(s) %s, remapped %d uncategorised org tracker pattern(s), cleared %d stale org description(s).\n",
			marked,
			verdict,
			remapped,
			cleared,
		)

		return nil
	}

	return cmd
}
