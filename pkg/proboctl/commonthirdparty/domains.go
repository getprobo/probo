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

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdDomains(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains <gid|slug>",
		Short: "List the domains of a common third party",
		Args:  cobra.ExactArgs(1),
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

		var domains coredata.CommonThirdPartyDomains

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				party, err := resolveCommonThirdParty(ctx, conn, args[0])
				if err != nil {
					return err
				}

				if err := domains.LoadByCommonThirdPartyID(ctx, conn, party.ID); err != nil {
					return fmt.Errorf("cannot load domains: %w", err)
				}

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, domains)
		}

		if len(domains) == 0 {
			_, _ = fmt.Fprintln(f.IOStreams.Out, "No domains found.")
			return nil
		}

		table := clicmdutil.NewTable("DOMAIN", "ID")
		for _, d := range domains {
			table.Row(d.Domain, d.ID.String())
		}

		_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())

		return nil
	}

	return cmd
}
