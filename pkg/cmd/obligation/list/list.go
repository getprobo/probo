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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ObligationOrder, $filter: ObligationFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      obligations(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            area
            source
            status
            type
            dueDate
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

type obligation struct {
	ID      string  `json:"id"`
	Area    string  `json:"area"`
	Source  string  `json:"source"`
	Status  string  `json:"status"`
	Type    string  `json:"type"`
	DueDate *string `json:"dueDate"`
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
		Short:   "List obligations in an organization",
		Aliases: []string{"ls"},
		Example: `  # List obligations in the default organization
  prb obligation list

  # List obligations sorted by due date
  prb obligation ls --order-by DUE_DATE --json`,
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

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "LAST_REVIEW_DATE", "DUE_DATE", "STATUS"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			obligations, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[obligation], error) {
					var resp struct {
						Node *struct {
							Typename    string                     `json:"__typename"`
							Obligations api.Connection[obligation] `json:"obligations"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Obligations, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, obligations)
			}

			if len(obligations) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No obligations found.")
				return nil
			}

			rows := make([][]string, 0, len(obligations))
			for _, o := range obligations {
				dueDate := ""
				if o.DueDate != nil {
					dueDate = *o.DueDate
				}
				rows = append(rows, []string{
					o.ID,
					o.Area,
					o.Source,
					o.Status,
					o.Type,
					dueDate,
				})
			}

			t := cmdutil.NewTable("ID", "AREA", "SOURCE", "STATUS", "TYPE", "DUE DATE").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(obligations) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d obligations\n",
					len(obligations),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of obligations to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, LAST_REVIEW_DATE, DUE_DATE, STATUS)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
