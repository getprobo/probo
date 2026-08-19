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

package commonthirdparty

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	seed "go.probo.inc/probo/pkg/proboctl/seed/common-third-parties"
)

// defaultPruneMinAge keeps entries that may still be in flight. An entry is
// created before enrichment fills it in and before the pattern that
// triggered it is linked, so a young entry with no references is normal
// rather than garbage.
const defaultPruneMinAge = 24 * time.Hour

func newCmdPrune(f *cmdutil.Factory) *cobra.Command {
	var (
		flagUnenrichedOnly bool
		flagOlderThan      time.Duration
		flagLimit          int
		flagDryRun         bool
		flagYes            bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete catalog entries that nothing references",
		Long: "Delete catalog entries with no catalog tracker pattern, no organization " +
			"third party in any tenant.\n\n" +
			"Owned domains are deleted with the entry. A domain is part of the " +
			"entry's own record rather than something referencing it, so it " +
			"carries no meaning once the entry is gone.\n\n" +
			"These are the leftovers of a bad attribution: an entry created for a " +
			"vendor that nothing ever linked. They are not duplicates of anything, so " +
			"`merge` cannot clean them up.\n\n" +
			"Curated seed entries are always excluded, with no override: the next seed " +
			"run would recreate them anyway.\n\n" +
			"Entries younger than --older-than are kept, because an entry is created " +
			"before enrichment fills it in and before its triggering pattern is " +
			"linked, so a young unreferenced entry is usually still in flight.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().BoolVar(&flagUnenrichedOnly, "unenriched-only", false, "Only entries that never completed enrichment")
	cmd.Flags().DurationVar(&flagOlderThan, "older-than", defaultPruneMinAge, "Only entries created longer ago than this")
	cmd.Flags().IntVar(&flagLimit, "limit", 0, "Maximum entries to delete (0 for all)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the entries that would be deleted")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if flagOlderThan < 0 {
			return fmt.Errorf("invalid --older-than value %s: must not be negative", flagOlderThan)
		}

		// A negative limit would fall through the cap below and delete every
		// candidate, the opposite of what a maximum means.
		if flagLimit < 0 {
			return fmt.Errorf("invalid --limit value %d: must not be negative (0 means no limit)", flagLimit)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		seededSlugs, err := seed.SeededSlugs()
		if err != nil {
			return fmt.Errorf("cannot load curated seed slugs: %w", err)
		}

		out := f.IOStreams.Out
		errOut := f.IOStreams.ErrOut

		var (
			candidates   []coredata.CommonThirdParty
			domainCounts = map[gid.GID]int{}
		)

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				var parties coredata.CommonThirdParties

				ids, err := parties.LoadAllUnreferencedIDs(
					ctx,
					conn,
					time.Now().Add(-flagOlderThan),
					flagUnenrichedOnly,
				)
				if err != nil {
					return err
				}

				for _, id := range ids {
					var party coredata.CommonThirdParty
					if err := party.LoadByID(ctx, conn, id); err != nil {
						return fmt.Errorf("cannot load common third party %s: %w", id, err)
					}

					// Curated entries are never pruned: the seed recreates
					// them, so deleting one is churn at best.
					if _, seeded := seededSlugs[party.Slug]; seeded {
						continue
					}

					candidates = append(candidates, party)
				}

				// Reported in the preview so deleting a row's domains with it
				// is visible rather than a silent side effect.
				var domains coredata.CommonThirdPartyDomains

				byParty, err := domains.LoadAllGroupedByCommonThirdPartyID(ctx, conn)
				if err != nil {
					return err
				}

				for id, list := range byParty {
					domainCounts[id] = len(list)
				}

				return nil
			},
		); err != nil {
			return err
		}

		if flagLimit > 0 && len(candidates) > flagLimit {
			_, _ = fmt.Fprintf(
				errOut,
				"Limiting to %d of %d candidate(s).\n",
				flagLimit,
				len(candidates),
			)

			candidates = candidates[:flagLimit]
		}

		if len(candidates) == 0 {
			_, _ = fmt.Fprintln(out, "No unreferenced common third parties to prune.")
			return nil
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would delete %d unreferenced common third party(ies):\n", len(candidates))
			printPruneSample(out, candidates, domainCounts)

			return nil
		}

		if !flagYes {
			return fmt.Errorf(
				"about to delete %d unreferenced common third party(ies); pass --yes to proceed or --dry-run to preview",
				len(candidates),
			)
		}

		var (
			deleted int
			skipped int
			failed  int
		)

		// One transaction per row so a single failure does not roll back the
		// rows already pruned. The delete re-checks the selection predicates,
		// so an entry that gained a reference, completed enrichment under
		// --unenriched-only, or disappeared since the selection is left alone.
		for _, party := range candidates {
			var gone bool

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					var err error

					gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(
						ctx,
						tx,
						party.ID,
						flagUnenrichedOnly,
					)

					return err
				},
			); err != nil {
				failed++

				_, _ = fmt.Fprintf(errOut, "error: cannot delete %q: %v\n", party.Slug, err)

				continue
			}

			if !gone {
				skipped++

				_, _ = fmt.Fprintf(
					errOut,
					"skipped %q: it is no longer eligible\n",
					party.Slug,
				)

				continue
			}

			deleted++
		}

		_, _ = fmt.Fprintf(out, "Deleted %d unreferenced common third party(ies).\n", deleted)

		if skipped > 0 {
			_, _ = fmt.Fprintf(out, "Skipped %d that are no longer eligible.\n", skipped)
		}

		if failed > 0 {
			return fmt.Errorf("%d of %d deletion(s) failed", failed, len(candidates))
		}

		return nil
	}

	return cmd
}

func printPruneSample(out io.Writer, parties []coredata.CommonThirdParty, domainCounts map[gid.GID]int) {
	const sampleSize = 10

	for i, party := range parties {
		if i >= sampleSize {
			_, _ = fmt.Fprintf(out, "  ... and %d more\n", len(parties)-sampleSize)
			break
		}

		_, _ = fmt.Fprintf(
			out,
			"  %s (%s) created=%s domains=%d\n",
			party.Name,
			party.Slug,
			party.CreatedAt.Format("2006-01-02 15:04:05"),
			domainCounts[party.ID],
		)
	}
}
