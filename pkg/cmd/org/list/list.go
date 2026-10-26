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
            legalName
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
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		LegalName *string `json:"legalName"`
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

			client, _, err := f.GraphQLClient(cmdutil.ConnectEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"filter": map[string]any{
					"states": []string{"ACTIVE"},
				},
			}

			profiles, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.ViewerConnection[profile]("profiles"),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				profiles,
				totalCount,
				cmdutil.ListView[profile]{
					Empty:   "No organizations found.",
					Noun:    "organizations",
					Headers: []string{"ID", "NAME", "LEGAL NAME", "ROLE"},
					Row: func(p profile) ([]string, error) {
						orgID := ""
						orgName := ""
						legalName := ""

						if p.Organization != nil {
							orgID = p.Organization.ID
							orgName = p.Organization.Name

							if p.Organization.LegalName != nil {
								legalName = *p.Organization.LegalName
							}
						}

						role := ""
						if p.Membership != nil {
							role = p.Membership.Role
						}

						return []string{
							orgID,
							orgName,
							legalName,
							role,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of organizations to list")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
