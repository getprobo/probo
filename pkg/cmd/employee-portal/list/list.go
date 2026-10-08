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
	"strconv"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: EmployeePortalOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      employeePortals(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            name
            active
            capabilities {
              deviceAgent
            }
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

type employeePortal struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Capabilities struct {
		DeviceAgent bool `json:"deviceAgent"`
	} `json:"capabilities"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List employee portals",
		Aliases: []string{"ls"},
		Example: `  # List employee portals for the default organization
  prb employee-portal list

  # List employee portals for a specific organization
  prb employee-portal ls --org <organization-id>`,
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

			if err := cmdutil.ValidateEnum("order-direction", flagOrderDir, []string{"ASC", "DESC"}); err != nil {
				return err
			}

			if flagLimit < 1 {
				return fmt.Errorf("--limit must be greater than 0")
			}

			variables := map[string]any{
				"id": flagOrg,
				"orderBy": map[string]any{
					"field":     "CREATED_AT",
					"direction": flagOrderDir,
				},
			}

			portals, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[employeePortal], error) {
					var resp struct {
						Node *struct {
							Typename        string                         `json:"__typename"`
							EmployeePortals api.Connection[employeePortal] `json:"employeePortals"`
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

					return &resp.Node.EmployeePortals, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, portals)
			}

			if len(portals) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No employee portals found.")
				return nil
			}

			rows := make([][]string, 0, len(portals))
			for _, portal := range portals {
				rows = append(rows, []string{
					portal.ID,
					portal.Name,
					strconv.FormatBool(portal.Active),
					strconv.FormatBool(portal.Capabilities.DeviceAgent),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "ACTIVE", "DEVICE AGENT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(portals) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d employee portals\n",
					len(portals),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of employee portals to list")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
