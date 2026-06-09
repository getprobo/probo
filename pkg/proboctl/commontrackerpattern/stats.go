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
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdStats(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Summarize the common tracker pattern catalog by enrichment and link state",
		Args:  cobra.NoArgs,
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		stats := map[string]int{}

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				counts := []struct {
					key    string
					filter *coredata.CommonTrackerPatternFilter
				}{
					{"total", coredata.NewCommonTrackerPatternFilter()},
					{"queued", coredata.NewCommonTrackerPatternFilter().WithState(new(coredata.CommonTrackerPatternEnrichmentStateQueued))},
					{"enriched", coredata.NewCommonTrackerPatternFilter().WithState(new(coredata.CommonTrackerPatternEnrichmentStateEnriched)).WithDescribed(new(true))},
					{"enriched (no description)", coredata.NewCommonTrackerPatternFilter().WithState(new(coredata.CommonTrackerPatternEnrichmentStateEnriched)).WithDescribed(new(false))},
					{"unenriched", coredata.NewCommonTrackerPatternFilter().WithState(new(coredata.CommonTrackerPatternEnrichmentStateUnenriched))},
					{"linked", coredata.NewCommonTrackerPatternFilter().WithLinked(new(true))},
					{"unlinked", coredata.NewCommonTrackerPatternFilter().WithLinked(new(false))},
				}

				for _, c := range counts {
					var ps coredata.CommonTrackerPatterns

					n, err := ps.CountAll(ctx, conn, c.filter)
					if err != nil {
						return err
					}

					stats[c.key] = n
				}

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, stats)
		}

		table := clicmdutil.NewTable("METRIC", "COUNT")
		for _, key := range []string{"total", "queued", "enriched", "enriched (no description)", "unenriched", "linked", "unlinked"} {
			table.Row(key, fmt.Sprintf("%d", stats[key]))
		}

		_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())

		return nil
	}

	return cmd
}
