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

package reorder

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const reorderMutation = `
mutation($input: ReorderCookieCategoryInput!) {
  reorderCookieCategory(input: $input) {
    cookieBanner {
      id
    }
  }
}
`

func NewCmdReorder(f *cmdutil.Factory) *cobra.Command {
	var flagRank int

	cmd := &cobra.Command{
		Use:   "reorder <id>",
		Short: "Change the rank of a cookie category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			_, err = client.Do(reorderMutation, map[string]any{
				"input": map[string]any{
					"cookieCategoryId": args[0],
					"rank":             flagRank,
				},
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Reordered cookie category %s to rank %d\n", args[0], flagRank)

			return nil
		},
	}

	cmd.Flags().IntVar(&flagRank, "rank", 0, "New rank position (required)")
	_ = cmd.MarkFlagRequired("rank")

	return cmd
}
