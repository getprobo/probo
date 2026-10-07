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

package view

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on Connector {
      id
      provider
      protocol
      createdAt
      accounts(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            externalAccountId
            name
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

type (
	account struct {
		ID                string `json:"id"`
		ExternalAccountID string `json:"externalAccountId"`
		Name              string `json:"name"`
	}

	accountEdge struct {
		Node account `json:"node"`
	}

	connectorView struct {
		Typename  string `json:"__typename"`
		ID        string `json:"id"`
		Provider  string `json:"provider"`
		Protocol  string `json:"protocol"`
		CreatedAt string `json:"createdAt"`
		Accounts  struct {
			TotalCount int           `json:"totalCount"`
			Edges      []accountEdge `json:"edges"`
		} `json:"accounts"`
	}
)

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a connector",
		Args:  cobra.ExactArgs(1),
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

			var viewed *connectorView

			// Walk every page. Paginate stops at this limit, which is not a page size.
			accounts, totalCount, err := api.Paginate(
				client,
				viewQuery,
				map[string]any{"id": args[0]},
				math.MaxInt32,
				func(data json.RawMessage) (*api.Connection[account], error) {
					var resp struct {
						Node *struct {
							Typename  string                  `json:"__typename"`
							ID        string                  `json:"id"`
							Provider  string                  `json:"provider"`
							Protocol  string                  `json:"protocol"`
							CreatedAt string                  `json:"createdAt"`
							Accounts  api.Connection[account] `json:"accounts"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("connector %s not found", args[0])
					}

					if resp.Node.Typename != "Connector" {
						return nil, fmt.Errorf("expected Connector node, got %s", resp.Node.Typename)
					}

					if viewed == nil {
						viewed = &connectorView{
							Typename:  resp.Node.Typename,
							ID:        resp.Node.ID,
							Provider:  resp.Node.Provider,
							Protocol:  resp.Node.Protocol,
							CreatedAt: resp.Node.CreatedAt,
						}
					}

					return &resp.Node.Accounts, nil
				},
			)
			if err != nil {
				return err
			}

			edges := make([]accountEdge, len(accounts))
			for i, a := range accounts {
				edges[i] = accountEdge{Node: a}
			}

			viewed.Accounts.TotalCount = totalCount
			viewed.Accounts.Edges = edges

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, viewed)
			}

			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Connector"))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), viewed.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Provider:"), viewed.Provider)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Protocol:"), viewed.Protocol)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(viewed.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Accounts:"), totalCount)

			if len(accounts) > 0 {
				_, _ = fmt.Fprintln(out)

				rows := make([][]string, 0, len(accounts))
				for _, a := range accounts {
					rows = append(rows, []string{
						a.ID,
						a.ExternalAccountID,
						a.Name,
					})
				}

				t := cmdutil.NewTable("ACCOUNT ID", "EXTERNAL ID", "NAME").Rows(rows...)
				_, _ = fmt.Fprintln(out, t)
			}

			if totalCount > len(accounts) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d accounts\n",
					len(accounts),
					totalCount,
				)
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
