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

package list

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TagOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      tags(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            key
            value
            color
            createdAt
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

type tag struct {
	ID        string  `json:"id"`
	Key       string  `json:"key"`
	Value     string  `json:"value"`
	Color     *string `json:"color"`
	CreatedAt string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tags in an organization",
		Aliases: []string{"ls"},
		Example: `  # List tags in the default organization
  prb tag list

  # List tags for a specific organization
  prb tag list --org prborg_... --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if f.IOStreams.IsInteractive() {
				if flagOrg == "" {
					err := huh.NewInput().
						Title("Organization ID").
						Value(&flagOrg).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagOrg == "" {
				return fmt.Errorf("organization ID is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "KEY"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			tags, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[tag](
					"tags",
					"Organization",
					fmt.Sprintf("organization %s not found", flagOrg),
				),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				tags,
				totalCount,
				cmdutil.ListView[tag]{
					Empty:      "No tags found.",
					Noun:       "tags",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "KEY", "VALUE", "COLOR", "CREATED"},
					Row: func(t tag) ([]string, error) {
						color := ""
						if t.Color != nil {
							color = *t.Color
						}

						return []string{
							t.ID,
							t.Key,
							t.Value,
							color,
							cmdutil.FormatTime(t.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of tags to list")
	cmd.Flags().StringVar(&flagAfter, "after", "", "Cursor to continue listing from")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, KEY)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
