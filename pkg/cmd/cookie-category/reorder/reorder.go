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
