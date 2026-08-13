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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	seed "go.probo.inc/probo/pkg/proboctl/seed/common-third-parties"
)

func newCmdMerge(f *cmdutil.Factory) *cobra.Command {
	var (
		flagInto            string
		flagReenrichWinner  bool
		flagDryRun          bool
		flagYes             bool
		flagContinueOnError bool
		flagAllowSeeded     bool
	)

	cmd := &cobra.Command{
		Use:   "merge --into <winner> <loser>...",
		Short: "Fold duplicate catalog entries into one",
		Long: "Move every reference from each loser onto the winner, then delete the " +
			"losers.\n\n" +
			"References are repointed across all tenants, because a global catalog " +
			"entry is referenced from every organization that imported it: catalog " +
			"tracker patterns are re-attributed, owned domains are moved (dropping " +
			"any the winner already claims), and organization third parties are " +
			"relinked. An organization that already links the winner keeps its own " +
			"row and is reported instead, since two rows in one organization " +
			"pointing at the same catalog entry would make one of them permanently " +
			"invisible.\n\n" +
			"This is not reversible: after the merge nothing records which of the " +
			"folded entries a given reference originally meant. Run --dry-run first.\n\n" +
			"Each loser is merged in its own transaction, so a failure part-way " +
			"leaves earlier merges committed.",
		Args: cobra.MinimumNArgs(1),
	}

	cmd.Flags().StringVar(&flagInto, "into", "", "Winner: the entry that absorbs the losers (slug or GID)")
	cmd.Flags().BoolVar(&flagReenrichWinner, "reenrich-winner", false, "Re-arm enrichment on the winner after merging")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print what each merge would move without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")
	cmd.Flags().BoolVar(&flagContinueOnError, "continue-on-error", false, "Keep going past a failed loser")
	cmd.Flags().BoolVar(&flagAllowSeeded, "allow-seeded", false, "Allow merging away a curated seed entry (it will be recreated by the next seed run)")

	_ = cmd.MarkFlagRequired("into")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		out := f.IOStreams.Out
		errOut := f.IOStreams.ErrOut

		var (
			winner coredata.CommonThirdParty
			losers []coredata.CommonThirdParty
		)

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				winner, err = resolveCommonThirdParty(ctx, conn, flagInto)
				if err != nil {
					return err
				}

				losers, err = resolveMergeLosers(ctx, conn, winner, args)

				return err
			},
		); err != nil {
			return err
		}

		if len(losers) == 0 {
			_, _ = fmt.Fprintln(out, "Nothing to merge.")
			return nil
		}

		// A seeded slug is recreated by the next seed run, so merging one
		// away silently reintroduces the duplicate on the following deploy.
		// That is a hard error rather than a warning: in a scripted run a
		// warning goes unread, and the failure resurfaces long after.
		if !flagAllowSeeded {
			seededSlugs, err := seed.SeededSlugs()
			if err != nil {
				return fmt.Errorf("cannot load curated seed slugs: %w", err)
			}

			for _, loser := range losers {
				if _, seeded := seededSlugs[loser.Slug]; seeded {
					return fmt.Errorf(
						"loser %q is a curated seed entry, so the next `seed common-third-parties` run would recreate it; "+
							"merge in the other direction (--into %s), remove it from the seed data, or pass --allow-seeded",
						loser.Slug,
						loser.Slug,
					)
				}
			}
		}

		if flagDryRun {
			return previewMerges(ctx, pgClient, out, winner, losers)
		}

		if !flagYes {
			return fmt.Errorf(
				"about to merge %d common third party(ies) into %q and delete them; pass --yes to proceed or --dry-run to preview",
				len(losers),
				winner.Slug,
			)
		}

		var failed int

		for _, loser := range losers {
			var result coredata.MergeCommonThirdPartyResult

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					result, err = coredata.MergeCommonThirdParty(ctx, tx, winner.ID, loser.ID)

					return err
				},
			); err != nil {
				failed++

				_, _ = fmt.Fprintf(errOut, "error: cannot merge %q: %v\n", loser.Slug, err)

				if !flagContinueOnError {
					return fmt.Errorf("cannot merge %q: %w", loser.Slug, err)
				}

				continue
			}

			printMergeResult(out, winner, loser, result, false)

			if len(result.ThirdPartiesSkipped) > 0 {
				_, _ = fmt.Fprintf(
					errOut,
					"warning: %d organization third party(ies) were left unlinked because their organization already links %q: %s\n",
					len(result.ThirdPartiesSkipped),
					winner.Slug,
					formatGIDs(result.ThirdPartiesSkipped),
				)
			}
		}

		if flagReenrichWinner {
			var requeued int64

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					var parties coredata.CommonThirdParties

					requeued, err = parties.RequestEnrichmentByIDs(ctx, tx, []gid.GID{winner.ID})

					return err
				},
			); err != nil {
				return fmt.Errorf("cannot enqueue enrichment for the winner: %w", err)
			}

			_, _ = fmt.Fprintf(out, "Queued %d common third party(ies) for the enrichment worker.\n", requeued)
		}

		if failed > 0 {
			return fmt.Errorf("%d of %d merge(s) failed", failed, len(losers))
		}

		return nil
	}

	return cmd
}

// resolveMergeLosers resolves the loser references, rejecting the winner
// among them and de-duplicating repeats so a slug passed twice is merged
// once.
func resolveMergeLosers(
	ctx context.Context,
	conn pg.Querier,
	winner coredata.CommonThirdParty,
	refs []string,
) ([]coredata.CommonThirdParty, error) {
	seen := make(map[gid.GID]struct{}, len(refs))
	losers := make([]coredata.CommonThirdParty, 0, len(refs))

	for _, ref := range refs {
		loser, err := resolveCommonThirdParty(ctx, conn, ref)
		if err != nil {
			return nil, err
		}

		if loser.ID == winner.ID {
			return nil, fmt.Errorf(
				"%q is the winner, so it cannot also be a loser: %w",
				ref,
				coredata.ErrCannotMergeIntoSelf,
			)
		}

		if _, ok := seen[loser.ID]; ok {
			continue
		}

		seen[loser.ID] = struct{}{}
		losers = append(losers, loser)
	}

	return losers, nil
}

// previewMerges reports what each merge would move, without writing.
func previewMerges(
	ctx context.Context,
	pgClient *pg.Client,
	out io.Writer,
	winner coredata.CommonThirdParty,
	losers []coredata.CommonThirdParty,
) error {
	return pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			for i, loser := range losers {
				if i > 0 {
					_, _ = fmt.Fprintln(out)
				}

				result, err := coredata.PreviewMergeCommonThirdParty(ctx, conn, winner.ID, loser.ID)
				if err != nil {
					if errors.Is(err, coredata.ErrCannotMergeIntoSelf) {
						return err
					}

					return fmt.Errorf("cannot preview merge of %q: %w", loser.Slug, err)
				}

				printMergeResult(out, winner, loser, result, true)
			}

			_, _ = fmt.Fprintf(out, "\nWould delete %d common third party(ies).\n", len(losers))

			return nil
		},
	)
}

// printMergeResult renders one merge's per-table account. The skipped
// organization third parties are named because they are the one outcome
// needing follow-up: they end up unlinked rather than moved.
func printMergeResult(
	out io.Writer,
	winner coredata.CommonThirdParty,
	loser coredata.CommonThirdParty,
	result coredata.MergeCommonThirdPartyResult,
	dryRun bool,
) {
	verb := "Merged"
	if dryRun {
		verb = "Would merge"
	}

	_, _ = fmt.Fprintf(
		out,
		"%s %q (%s) into %q (%s):\n",
		verb,
		loser.Name,
		loser.Slug,
		winner.Name,
		winner.Slug,
	)

	_, _ = fmt.Fprintf(
		out,
		"  domains:            %d moved, %d dropped as duplicate\n",
		result.DomainsMoved,
		result.DomainsDroppedAsDup,
	)

	_, _ = fmt.Fprintf(
		out,
		"  tracker patterns:   %d repointed, %d re-queued for enrichment\n",
		result.TrackerPatternsRepointed,
		result.TrackerPatternsRequeued,
	)

	_, _ = fmt.Fprintf(
		out,
		"  org patterns:       %d relinked to the managed third party\n",
		result.OrgTrackerPatternsRelinked,
	)

	_, _ = fmt.Fprintf(
		out,
		"  org third parties:  %d repointed, %d skipped (organization already links the winner)\n",
		result.ThirdPartiesRepointed,
		len(result.ThirdPartiesSkipped),
	)

	if len(result.ThirdPartiesSkipped) > 0 {
		_, _ = fmt.Fprintf(out, "                      skipped: %s\n", formatGIDs(result.ThirdPartiesSkipped))
	}

	logo := "unchanged"
	if result.LogoAdopted {
		logo = "adopted from the loser (the winner had none)"
	}

	_, _ = fmt.Fprintf(out, "  logo:               %s\n", logo)
}

// formatGIDs renders ids for an operator, capped so a wide blast radius
// stays readable while still reporting the true total.
func formatGIDs(ids []gid.GID) string {
	const sampleSize = 10

	var out strings.Builder

	for i, id := range ids {
		if i >= sampleSize {
			fmt.Fprintf(&out, " ... and %d more", len(ids)-sampleSize)
			break
		}

		if i > 0 {
			out.WriteString(" ")
		}

		out.WriteString(id.String())
	}

	return out.String()
}
