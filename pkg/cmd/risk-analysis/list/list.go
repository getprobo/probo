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
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskAnalysisOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      riskAnalyses(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            name
            description
            period {
              start
              end
            }
            matrixSize {
              rows
              cols
            }
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

type riskAnalysis struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Period      *struct {
		Start *string `json:"start"`
		End   *string `json:"end"`
	} `json:"period"`
	MatrixSize struct {
		Rows int `json:"rows"`
		Cols int `json:"cols"`
	} `json:"matrixSize"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
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
		Short:   "List risk analyses in an organization",
		Aliases: []string{"ls"},
		Example: `  # List risk analyses in the default organization
  prb risk-analysis list

  # List risk analyses sorted by name
  prb risk-analysis ls --order-by NAME --json`,
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
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "NAME"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			riskAnalyses, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[riskAnalysis]("riskAnalyses", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				riskAnalyses,
				totalCount,
				cmdutil.ListView[riskAnalysis]{
					Empty:      "No risk analyses found.",
					Noun:       "risk analyses",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "NAME", "DESCRIPTION", "PERIOD START", "PERIOD END", "MATRIX SIZE", "CREATED AT"},
					Row: func(r riskAnalysis) ([]string, error) {
						desc := ""

						if r.Description != nil {
							formatted, err := cmdutil.FormatRichText(*r.Description)
							if err != nil {
								return nil, fmt.Errorf("cannot format description: %w", err)
							}

							// Hard breaks render as "\\\n"; strip that marker before
							// flattening remaining newlines into a single table cell.
							desc = strings.ReplaceAll(formatted, "\\\n", " ")
							desc = strings.ReplaceAll(desc, "\n", " ")
						}

						periodStart := ""
						periodEnd := ""

						if r.Period != nil {
							if r.Period.Start != nil {
								periodStart = cmdutil.FormatTime(*r.Period.Start)
							}

							if r.Period.End != nil {
								periodEnd = cmdutil.FormatTime(*r.Period.End)
							}
						}

						return []string{
							r.ID,
							r.Name,
							desc,
							periodStart,
							periodEnd,
							fmt.Sprintf("%d×%d", r.MatrixSize.Rows, r.MatrixSize.Cols),
							cmdutil.FormatTime(r.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of risk analyses to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
