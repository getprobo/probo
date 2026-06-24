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
	var (
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List organizations you have access to",
		Aliases: []string{"ls"},
		Example: `  # List all organizations
  prb org list

  # Output as JSON
  prb org ls --json`,
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
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

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, profiles)
			}

			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No organizations found.")
				return nil
			}

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

			t := cmdutil.NewTable("ID", "NAME", "ROLE").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(profiles) {
				_, _ = fmt.Fprintf(
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
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
