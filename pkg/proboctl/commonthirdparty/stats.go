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

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdStats(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Summarize the common third party catalog by enrichment state and status",
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
					filter *coredata.CommonThirdPartyFilter
				}{
					{"total", coredata.NewCommonThirdPartyFilter(nil)},
					{"queued", coredata.NewCommonThirdPartyFilter(nil).WithState(new(coredata.CommonThirdPartyEnrichmentStateQueued))},
					{"enriched", coredata.NewCommonThirdPartyFilter(nil).WithState(new(coredata.CommonThirdPartyEnrichmentStateEnriched))},
					{"unenriched", coredata.NewCommonThirdPartyFilter(nil).WithState(new(coredata.CommonThirdPartyEnrichmentStateUnenriched))},
					{"status: done", coredata.NewCommonThirdPartyFilter(nil).WithEnrichmentStatus(new("done"))},
					{"status: partial", coredata.NewCommonThirdPartyFilter(nil).WithEnrichmentStatus(new("partial"))},
					{"status: failed", coredata.NewCommonThirdPartyFilter(nil).WithEnrichmentStatus(new("failed"))},
				}

				for _, c := range counts {
					var parties coredata.CommonThirdParties

					n, err := parties.CountAll(ctx, conn, c.filter)
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

		order := []string{"total", "queued", "enriched", "unenriched", "status: done", "status: partial", "status: failed"}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, stats)
		}

		table := clicmdutil.NewTable("METRIC", "COUNT")
		for _, key := range order {
			table.Row(key, fmt.Sprintf("%d", stats[key]))
		}

		_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())

		return nil
	}

	return cmd
}
