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
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: InternalControlOrder, $filter: InternalControlFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      internalControls(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            name
            code
            category
            implementationStatus
            state
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

type internalControl struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Code                 *string `json:"code"`
	Category             string  `json:"category"`
	ImplementationStatus string  `json:"implementationStatus"`
	State                string  `json:"state"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                    string
		flagLimit                  int
		flagAfter                  string
		flagOrderBy                string
		flagOrderDir               string
		flagFilter                 string
		flagImplementationStatuses []string
		flagOwner                  string
		flagNeedsAttention         bool
		flagOutput                 *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List internal controls in an organization",
		Aliases: []string{"ls"},
		Example: `  # List internal controls in the default organization
  prb internal-control list

  # Filter internal controls by name
  prb internal-control list --filter "encryption"

  # List internal controls sorted by name
  prb internal-control ls --order-by NAME --json

  # List controls that need attention, owned by one profile
  prb internal-control list --needs-attention --owner <profile-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "NAME", "NEXT_DUE"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}
			if flagFilter != "" {
				filter["query"] = flagFilter
			}
			if len(flagImplementationStatuses) > 0 {
				if err := cmdutil.ValidateEnum(
					"implementation-status",
					flagImplementationStatuses[0],
					[]string{"NOT_IMPLEMENTED", "IN_PROGRESS", "IMPLEMENTED", "OPERATING"},
				); err != nil {
					return err
				}
				for _, status := range flagImplementationStatuses[1:] {
					if err := cmdutil.ValidateEnum(
						"implementation-status",
						status,
						[]string{"NOT_IMPLEMENTED", "IN_PROGRESS", "IMPLEMENTED", "OPERATING"},
					); err != nil {
						return err
					}
				}
				filter["implementationStatuses"] = flagImplementationStatuses
			}
			if flagOwner != "" {
				filter["ownerId"] = flagOwner
			}
			if flagNeedsAttention {
				filter["needsAttention"] = true
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			internalControls, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[internalControl]("internalControls", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				internalControls,
				totalCount,
				cmdutil.ListView[internalControl]{
					Empty:      "No internal controls found.",
					Noun:       "internal controls",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "CODE", "NAME", "CATEGORY", "STATUS", "STATE"},
					Row: func(m internalControl) ([]string, error) {
						code := ""
						if m.Code != nil {
							code = *m.Code
						}

						return []string{
							m.ID,
							code,
							m.Name,
							m.Category,
							m.ImplementationStatus,
							m.State,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of internal controls to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME, NEXT_DUE)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVarP(&flagFilter, "filter", "q", "", "Filter internal controls by name or code")
	cmd.Flags().StringArrayVar(&flagImplementationStatuses, "implementation-status", nil, "Filter by implementation status (repeatable)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Filter by owner profile ID")
	cmd.Flags().BoolVar(&flagNeedsAttention, "needs-attention", false, "Only controls that are not implemented or overdue")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
