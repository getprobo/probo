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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: FindingOrder, $filter: FindingFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      findings(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            referenceId
            kind
            status
            priority
            dueDate
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

type finding struct {
	ID          string  `json:"id"`
	ReferenceID string  `json:"referenceId"`
	Kind        string  `json:"kind"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"dueDate"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagKind     string
		flagAudit    string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List findings in an organization",
		Aliases: []string{"ls"},
		Example: `  # List findings in an organization
  prb finding list --org <organization-id>

  # Filter by kind and output as JSON
  prb finding ls --org <organization-id> --kind MINOR_NONCONFORMITY --json

  # List findings linked to an audit
  prb finding ls --org <organization-id> --audit <audit-id>`,
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
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "REFERENCE_ID", "IDENTIFIED_ON", "DUE_DATE", "STATUS", "PRIORITY", "KIND"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}

			if flagKind != "" {
				if err := cmdutil.ValidateEnum("kind", flagKind, []string{"MINOR_NONCONFORMITY", "MAJOR_NONCONFORMITY", "OBSERVATION", "EXCEPTION"}); err != nil {
					return err
				}

				filter["kind"] = flagKind
			}

			if flagAudit != "" {
				filter["auditId"] = flagAudit
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			findings, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[finding]("findings", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				findings,
				totalCount,
				cmdutil.ListView[finding]{
					Empty:      "No findings found.",
					Noun:       "findings",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "REFERENCE", "KIND", "STATUS", "PRIORITY", "DUE DATE"},
					Row: func(fi finding) ([]string, error) {
						dueDate := "-"
						if fi.DueDate != nil {
							dueDate = cmdutil.FormatTime(*fi.DueDate)
						}

						return []string{
							fi.ID,
							fi.ReferenceID,
							fi.Kind,
							fi.Status,
							fi.Priority,
							dueDate,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of findings to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, REFERENCE_ID, IDENTIFIED_ON, DUE_DATE, STATUS, PRIORITY, KIND)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagKind, "kind", "", "Filter by kind (MINOR_NONCONFORMITY, MAJOR_NONCONFORMITY, OBSERVATION, EXCEPTION)")
	cmd.Flags().StringVar(&flagAudit, "audit", "", "Filter by audit ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
