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
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/service-account/shared"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on Organization {
      serviceAccounts(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            organizationId
            name
            description
            scopes
            disabledAt
            createdAt
            updatedAt
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

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List service accounts",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
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

			organizationID, err := shared.ResolveOrganization(flagOrg, hc)
			if err != nil {
				return err
			}

			accounts, totalCount, err := api.Paginate(
				shared.NewClient(cfg, host, hc),
				listQuery,
				map[string]any{"id": organizationID},
				flagLimit,
				func(data json.RawMessage) (*api.Connection[shared.ServiceAccount], error) {
					var resp struct {
						Node *struct {
							Typename        string                                `json:"__typename"`
							ServiceAccounts api.Connection[shared.ServiceAccount] `json:"serviceAccounts"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", organizationID)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.ServiceAccounts, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, accounts)
			}
			if len(accounts) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No service accounts found.")
				return nil
			}

			rows := make([][]string, 0, len(accounts))
			for _, account := range accounts {
				status := "active"
				if account.DisabledAt != nil {
					status = "disabled"
				}
				rows = append(rows, []string{
					account.ID,
					account.Name,
					strings.Join(account.Scopes, ", "),
					status,
					cmdutil.FormatTime(account.CreatedAt),
				})
			}

			table := cmdutil.NewTable("ID", "NAME", "SCOPES", "STATUS", "CREATED").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, table)

			if totalCount > len(accounts) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d service accounts\n",
					len(accounts),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of service accounts to list")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
