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

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskOrder, $filter: RiskFilter) {
  node(id: $id) {
    ... on Organization {
      risks(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            name
            category
            treatment
            inherentRiskScore
            residualRiskScore
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

type risk struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Category          string `json:"category"`
	Treatment         string `json:"treatment"`
	InherentRiskScore int    `json:"inherentRiskScore"`
	ResidualRiskScore int    `json:"residualRiskScore"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg     string
		flagLimit   int
		flagOrderBy string
		flagFilter  string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List risks in an organization",
		Aliases: []string{"ls"},
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
				cfg.HTTPTimeoutDuration(),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'proboctl auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": "DESC",
				}
			}

			if flagFilter != "" {
				variables["filter"] = map[string]any{
					"query": flagFilter,
				}
			}

			risks, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[risk], error) {
					var resp struct {
						Node struct {
							Risks api.Connection[risk] `json:"risks"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					return &resp.Node.Risks, nil
				},
			)
			if err != nil {
				return err
			}

			if len(risks) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No risks found.")
				return nil
			}

			headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
			cellStyle := lipgloss.NewStyle().Padding(0, 1)

			rows := make([][]string, 0, len(risks))
			for _, r := range risks {
				rows = append(rows, []string{
					r.ID,
					r.Name,
					r.Category,
					r.Treatment,
					fmt.Sprintf("%d", r.InherentRiskScore),
					fmt.Sprintf("%d", r.ResidualRiskScore),
				})
			}

			t := table.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
				Headers("ID", "NAME", "CATEGORY", "TREATMENT", "INHERENT", "RESIDUAL").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}
					return cellStyle
				}).
				Rows(rows...)

			fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(risks) {
				fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d risks\n",
					len(risks),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of risks to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME, CATEGORY, TREATMENT, INHERENT_RISK_SCORE, RESIDUAL_RISK_SCORE)")
	cmd.Flags().StringVarP(&flagFilter, "filter", "q", "", "Filter risks by search query")

	return cmd
}
