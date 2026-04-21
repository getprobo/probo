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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ProcessingActivityOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      processingActivities(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            name
            role
            lawfulBasis
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

type processingActivity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	LawfulBasis string `json:"lawfulBasis"`
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
		Short:   "List processing activities in an organization",
		Aliases: []string{"ls"},
		Example: `  # List processing activities in the default organization
  prb processing-activity list

  # List processing activities sorted by name
  prb pa ls --order-by NAME --json`,
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
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "NAME"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			activities, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[processingActivity], error) {
					var resp struct {
						Node *struct {
							Typename             string                             `json:"__typename"`
							ProcessingActivities api.Connection[processingActivity] `json:"processingActivities"`
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
					return &resp.Node.ProcessingActivities, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, activities)
			}

			if len(activities) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No processing activities found.")
				return nil
			}

			rows := make([][]string, 0, len(activities))
			for _, a := range activities {
				rows = append(rows, []string{
					a.ID,
					a.Name,
					a.Role,
					a.LawfulBasis,
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "ROLE", "LAWFUL BASIS").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(activities) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d processing activities\n",
					len(activities),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of processing activities to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
