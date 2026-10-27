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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: AuditOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      audits(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            name
            firm
            state
            validity {
              start
              end
            }
            auditDates {
              start
              end
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
}
`

type audit struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Firm     string `json:"firm"`
	State    string `json:"state"`
	Validity *struct {
		Start *string `json:"start"`
		End   *string `json:"end"`
	} `json:"validity"`
	AuditDates *struct {
		Start *string `json:"start"`
		End   *string `json:"end"`
	} `json:"auditDates"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List audits in an organization",
		Aliases: []string{"ls"},
		Example: `  # List audits in the default organization
  prb audit list

  # List audits sorted by state
  prb audit ls --order-by STATE --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum("order-direction", flagOrderDir, []string{"ASC", "DESC"}); err != nil {
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
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "VALID_FROM", "VALID_UNTIL", "AUDIT_START_DATE", "AUDIT_END_DATE", "STATE"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			audits, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[audit]("audits", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				audits,
				totalCount,
				cmdutil.ListView[audit]{
					Empty:      "No audits found.",
					Noun:       "audits",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "NAME", "FIRM", "STATE", "VALID FROM", "VALID UNTIL", "AUDIT START", "AUDIT END"},
					Row: func(a audit) ([]string, error) {
						validFrom, validUntil := "", ""

						if a.Validity != nil {
							if a.Validity.Start != nil {
								validFrom = *a.Validity.Start
							}

							if a.Validity.End != nil {
								validUntil = *a.Validity.End
							}
						}

						auditStartDate, auditEndDate := "", ""

						if a.AuditDates != nil {
							if a.AuditDates.Start != nil {
								auditStartDate = *a.AuditDates.Start
							}

							if a.AuditDates.End != nil {
								auditEndDate = *a.AuditDates.End
							}
						}

						return []string{
							a.ID,
							a.Name,
							a.Firm,
							a.State,
							validFrom,
							validUntil,
							auditStartDate,
							auditEndDate,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of audits to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, VALID_FROM, VALID_UNTIL, AUDIT_START_DATE, AUDIT_END_DATE, STATE)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
