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
	"strconv"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: CompliancePortalOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      compliancePortals(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            slug
            entityName
            active
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

type compliancePortal struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	EntityName string `json:"entityName"`
	Active     bool   `json:"active"`
	CreatedAt  string `json:"createdAt"`
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
		Short:   "List compliance portals",
		Aliases: []string{"ls"},
		Example: `  # List compliance portals for the default organization
  prb compliance-portal list

  # List compliance portals for a specific organization
  prb compliance-portal ls --org <organization-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
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
				api.NodeConnection[compliancePortal]("compliancePortals", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				portals,
				totalCount,
				cmdutil.ListView[compliancePortal]{
					Empty:   "No compliance portals found.",
					Noun:    "compliance portals",
					Headers: []string{"ID", "SLUG", "ENTITY NAME", "ACTIVE"},
					Row: func(portal compliancePortal) ([]string, error) {
						return []string{
							portal.ID,
							portal.Slug,
							portal.EntityName,
							strconv.FormatBool(portal.Active),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of compliance portals to list")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
