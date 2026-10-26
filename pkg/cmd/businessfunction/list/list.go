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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: BusinessFunctionOrder, $filter: BusinessFunctionFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      businessFunctions(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            referenceId
            name
            classification
            mtdMinutes
            rtoMinutes
            rpoMinutes
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

type businessFunction struct {
	ID             string `json:"id"`
	ReferenceID    string `json:"referenceId"`
	Name           string `json:"name"`
	Classification string `json:"classification"`
	MTDMinutes     int    `json:"mtdMinutes"`
	RTOMinutes     int    `json:"rtoMinutes"`
	RPOMinutes     int    `json:"rpoMinutes"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg            string
		flagLimit          int
		flagOrderBy        string
		flagOrderDir       string
		flagClassification string
		flagOutput         *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List business functions in an organization",
		Aliases: []string{"ls"},
		Example: `  # List business functions in the default organization
  prb business-function list

  # List critical business functions only
  prb business-function ls --classification CRITICAL --json`,
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

			if err := cmdutil.ValidateEnum(
				"order-direction",
				flagOrderDir,
				[]string{"ASC", "DESC"},
			); err != nil {
				return err
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT", "REFERENCE_ID", "NAME", "CLASSIFICATION", "MTD_MINUTES", "RTO_MINUTES", "RPO_MINUTES"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}

			if flagClassification != "" {
				if err := cmdutil.ValidateEnum(
					"classification",
					flagClassification,
					[]string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"},
				); err != nil {
					return err
				}

				filter["classification"] = flagClassification
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			businessFunctions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeConnection[businessFunction]("businessFunctions", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				businessFunctions,
				totalCount,
				cmdutil.ListView[businessFunction]{
					Empty: "No business functions found.",
					Noun:  "business functions",
					Headers: []string{
						"ID",
						"REFERENCE ID",
						"NAME",
						"CLASSIFICATION",
						"MTD",
						"RTO",
						"RPO",
					},
					Row: func(bf businessFunction) ([]string, error) {
						return []string{
							bf.ID,
							bf.ReferenceID,
							bf.Name,
							bf.Classification,
							fmt.Sprintf("%d", bf.MTDMinutes),
							fmt.Sprintf("%d", bf.RTOMinutes),
							fmt.Sprintf("%d", bf.RPOMinutes),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of business functions to list")
	cmd.Flags().StringVar(
		&flagOrderBy,
		"order-by",
		"",
		"Order by field (CREATED_AT, REFERENCE_ID, NAME, CLASSIFICATION, MTD_MINUTES, RTO_MINUTES, RPO_MINUTES)",
	)
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Filter by classification (CRITICAL, IMPORTANT, SECONDARY, STANDARD)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
