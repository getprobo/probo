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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ProfileOrder, $filter: ProfileFilter) {
  node(id: $id) {
    ... on Organization {
      profiles(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            fullName
            emailAddress
            state
            kind
            position
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

type profile struct {
	ID           string  `json:"id"`
	FullName     string  `json:"fullName"`
	EmailAddress string  `json:"emailAddress"`
	State        string  `json:"state"`
	Kind         *string `json:"kind"`
	Position     *string `json:"position"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagLimit  int
		flagOrder  string
		flagActive bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List users in an organization",
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
				"/console/v1/graphql",
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

			if flagOrder != "" {
				variables["orderBy"] = map[string]any{
					"field":     flagOrder,
					"direction": "ASC",
				}
			}

			if flagActive {
				variables["filter"] = map[string]any{
					"excludeContractEnded": true,
				}
			}

			profiles, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[profile], error) {
					var resp struct {
						Node struct {
							Profiles api.Connection[profile] `json:"profiles"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					return &resp.Node.Profiles, nil
				},
			)
			if err != nil {
				return err
			}

			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No users found.")
				return nil
			}

			headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
			cellStyle := lipgloss.NewStyle().Padding(0, 1)

			rows := make([][]string, 0, len(profiles))
			for _, p := range profiles {
				kind := ""
				if p.Kind != nil {
					kind = *p.Kind
				}

				position := ""
				if p.Position != nil {
					position = *p.Position
				}

				rows = append(rows, []string{
					p.ID,
					p.FullName,
					p.EmailAddress,
					p.State,
					kind,
					position,
				})
			}

			t := table.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
				Headers("ID", "NAME", "EMAIL", "STATE", "KIND", "POSITION").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}
					return cellStyle
				}).
				Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(profiles) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d users\n",
					len(profiles),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of users to list")
	cmd.Flags().StringVar(&flagOrder, "order-by", "", "Order by field (FULL_NAME, CREATED_AT, KIND)")
	cmd.Flags().BoolVar(&flagActive, "active", false, "Exclude users whose contract has ended")

	return cmd
}
