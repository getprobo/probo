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
		flagOrg      string
		flagLimit    int
		flagOrder    string
		flagOrderDir string
		flagActive   bool
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List users in an organization",
		Aliases: []string{"ls"},
		Example: `  # List users in the default organization
  proboctl user list

  # List only active users as JSON
  proboctl user ls --active --json`,
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
				if err := cmdutil.ValidateEnum("order-by", flagOrder, []string{"FULL_NAME", "CREATED_AT", "KIND"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrder,
					"direction": flagOrderDir,
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

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, profiles)
			}

			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No users found.")
				return nil
			}

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

			t := cmdutil.NewTable("ID", "NAME", "EMAIL", "STATE", "KIND", "POSITION").Rows(rows...)

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
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().BoolVar(&flagActive, "active", false, "Exclude users whose contract has ended")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
