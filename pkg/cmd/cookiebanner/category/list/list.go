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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      categories(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            name
            description
            required
            rank
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

type cookieCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Rank        int    `json:"rank"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBanner string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cookie categories for a banner",
		Aliases: []string{"ls"},
		Example: `  # List categories for a cookie banner
  prb cookie-banner category list --banner <banner-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

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
			)

			if flagBanner == "" {
				return fmt.Errorf("banner is required; pass --banner")
			}

			variables := map[string]any{
				"id": flagBanner,
			}

			categories, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[cookieCategory], error) {
					var resp struct {
						Node *struct {
							Typename   string                         `json:"__typename"`
							Categories api.Connection[cookieCategory] `json:"categories"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("cookie banner %s not found", flagBanner)
					}
					if resp.Node.Typename != "CookieBanner" {
						return nil, fmt.Errorf("expected CookieBanner node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Categories, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, categories)
			}

			if len(categories) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No cookie categories found.")
				return nil
			}

			rows := make([][]string, 0, len(categories))
			for _, c := range categories {
				required := "no"
				if c.Required {
					required = "yes"
				}
				rows = append(rows, []string{
					c.ID,
					c.Name,
					required,
					fmt.Sprintf("%d", c.Rank),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "REQUIRED", "RANK").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(categories) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d categories\n",
					len(categories),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBanner, "banner", "", "Cookie banner ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of categories to list")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("banner")

	return cmd
}
