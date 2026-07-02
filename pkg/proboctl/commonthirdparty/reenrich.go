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

package commonthirdparty

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
		flagIDs      []string
		flagSlugs    []string
		flagCategory string
		flagKeyword  string
		flagState    string
		flagStatus   string
		flagDryRun   bool
		flagYes      bool
	)

	cmd := &cobra.Command{
		Use:   "reenrich",
		Short: "Re-enrich common third parties via the enrichment worker",
		Long: "Re-arm the async common-third-party enrichment worker for selected " +
			"catalog rows. The worker re-resolves company profile, compliance " +
			"documents, owned domains, and the logo, merging results over prior " +
			"per-field provenance so curated seed data and human edits are never " +
			"overwritten. Already-enriched rows are re-processed. Enrichment is " +
			"expensive (LLM + browser per row), so a non-empty selection requires " +
			"--yes; use --dry-run to preview.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringSliceVar(&flagIDs, "id", nil, "Common third party GID(s) to re-enrich (repeatable)")
	cmd.Flags().StringSliceVar(&flagSlugs, "slug", nil, "Common third party slug(s) to re-enrich (repeatable)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Select rows by category")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Select rows by a name/slug substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Select rows by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Select rows by last enrichment status (done, partial, failed)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the selected rows without enriching")
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
			flagSlugs,
			flagCategory,
			flagKeyword,
			flagState,
			flagStatus,
		)
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		if len(ids) == 0 {
			_, _ = fmt.Fprintln(out, "No common third parties matched the selection.")
			return nil
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would re-enrich %d common third party(ies).\n", len(ids))
			printSample(out, ids)

			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to re-enrich %d common third party(ies); pass --yes to proceed or --dry-run to preview", len(ids))
		}

		var requeued int64

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var parties coredata.CommonThirdParties

				requeued, err = parties.RequestEnrichmentByIDs(ctx, tx, ids)

				return err
			},
		); err != nil {
			return fmt.Errorf("cannot enqueue enrichment: %w", err)
		}

		_, _ = fmt.Fprintf(out, "Queued %d common third party(ies) for the enrichment worker.\n", requeued)

		return nil
	}

	return cmd
}

// resolveReenrichIDs turns the selection flags into the set of common
// third party IDs to re-enrich. Explicit selection (--id and/or --slug)
// is used verbatim and the filtering flags do not apply. With no
// explicit selection the filtering flags (--category, --keyword,
// --state, --status) select across the whole catalog.
func resolveReenrichIDs(
	ctx context.Context,
	pgClient *pg.Client,
	rawIDs, rawSlugs []string,
	category, keyword, state, status string,
) ([]gid.GID, error) {
	explicit := len(rawIDs) > 0 || len(rawSlugs) > 0
	filtered := category != "" || keyword != "" || state != "" || status != ""

	if explicit && filtered {
		return nil, fmt.Errorf("--id/--slug cannot be combined with --category, --keyword, --state, or --status")
	}

	if explicit {
		return resolveExplicitIDs(ctx, pgClient, rawIDs, rawSlugs)
	}

	filter, err := buildReenrichFilter(category, keyword, state, status)
	if err != nil {
		return nil, err
	}

	var ids []gid.GID

	err = pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var parties coredata.CommonThirdParties

			ids, err = parties.LoadAllIDs(ctx, conn, filter)

			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// resolveExplicitIDs parses the --id GIDs and resolves the --slug values,
// preserving order and de-duplicating the combined set.
func resolveExplicitIDs(
	ctx context.Context,
	pgClient *pg.Client,
	rawIDs, rawSlugs []string,
) ([]gid.GID, error) {
	seen := make(map[gid.GID]struct{}, len(rawIDs)+len(rawSlugs))
	ids := make([]gid.GID, 0, len(rawIDs)+len(rawSlugs))

	add := func(id gid.GID) {
		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}

		ids = append(ids, id)
	}

	for _, raw := range rawIDs {
		id, err := gid.ParseGID(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --id value %q: %w", raw, err)
		}

		add(id)
	}

	if len(rawSlugs) > 0 {
		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				for _, slug := range rawSlugs {
					id, err := resolveCommonThirdPartyID(ctx, conn, slug)
					if err != nil {
						return err
					}

					add(id)
				}

				return nil
			},
		); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

func buildReenrichFilter(category, keyword, state, status string) (*coredata.CommonThirdPartyFilter, error) {
	filter := coredata.NewCommonThirdPartyFilter(nil)

	if category != "" {
		cat := coredata.ThirdPartyCategory(category)
		if !cat.IsValid() {
			return nil, fmt.Errorf("invalid --category value %q", category)
		}

		filter.WithCategory(&cat)
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

	if status != "" {
		if _, ok := validEnrichmentStatuses[status]; !ok {
			return nil, fmt.Errorf("invalid --status value %q: valid values are done, partial, failed", status)
		}

		filter.WithEnrichmentStatus(&status)
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
