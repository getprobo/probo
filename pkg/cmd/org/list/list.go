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
query($first: Int, $after: CursorKey, $orderBy: ProfileOrder, $filter: ProfileFilter) {
  viewer {
    profiles(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
      totalCount
      edges {
        node {
          id
          state
          organization {
            id
            name
          }
          membership {
            role
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
`

type (
	organization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	membership struct {
		Role string `json:"role"`
	}

	profile struct {
		ID           string        `json:"id"`
		State        string        `json:"state"`
		Organization *organization `json:"organization"`
		Membership   *membership   `json:"membership"`
	}
)

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var flagLimit int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List organizations you have access to",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

			variables := map[string]any{
				"filter": map[string]any{
					"state": "ACTIVE",
				},
			}

			profiles, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[profile], error) {
					var resp struct {
						Viewer struct {
							Profiles api.Connection[profile] `json:"profiles"`
						} `json:"viewer"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					return &resp.Viewer.Profiles, nil
				},
			)
			if err != nil {
				return err
			}

			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No organizations found.")
				return nil
			}

			headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
			cellStyle := lipgloss.NewStyle().Padding(0, 1)

			rows := make([][]string, 0, len(profiles))
			for _, p := range profiles {
				orgID := ""
				orgName := ""
				if p.Organization != nil {
					orgID = p.Organization.ID
					orgName = p.Organization.Name
				}

				role := ""
				if p.Membership != nil {
					role = p.Membership.Role
				}

				rows = append(rows, []string{
					orgID,
					orgName,
					role,
				})
			}

			t := table.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
				Headers("ID", "NAME", "ROLE").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}
					return cellStyle
				}).
				Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(profiles) {
				fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d organizations\n",
					len(profiles),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of organizations to list")

	return cmd
}
