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
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ResourceTagOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      resourceTags(first: $first, after: $after, orderBy: $orderBy) {
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

type resourceTag struct {
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
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List resource tags in an organization",
		Aliases: []string{"ls"},
		Example: `  # List tags in the default organization
  prb resource-tag list

  # List tags for a specific organization
  prb resource-tag list --organization-id prborg_... --json`,
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
				return fmt.Errorf("organization ID is required; pass --organization-id or set a default with 'prb auth login'")
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

			tags, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[resourceTag], error) {
					var resp struct {
						Node *struct {
							Typename     string                      `json:"__typename"`
							ResourceTags api.Connection[resourceTag] `json:"resourceTags"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}

					return &resp.Node.ResourceTags, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				if tags == nil {
					tags = []resourceTag{}
				}

				return cmdutil.PrintJSON(f.IOStreams.Out, tags)
			}

			if len(tags) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No resource tags found.")
				return nil
			}

			rows := make([][]string, 0, len(tags))
			for _, t := range tags {
				color := ""
				if t.Color != nil {
					color = *t.Color
				}

				rows = append(rows, []string{
					t.ID,
					t.Key,
					t.Value,
					color,
					cmdutil.FormatTime(t.CreatedAt),
				})
			}

			table := cmdutil.NewTable("ID", "KEY", "VALUE", "COLOR", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, table)

			if totalCount > len(tags) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d resource tags\n",
					len(tags),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "organization-id", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of resource tags to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, KEY)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
