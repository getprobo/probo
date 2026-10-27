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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: AiSystemOrder, $filter: AiSystemFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      aiSystems(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            name
            version
            companyRoles
            status
            riskClassification
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

type aiSystem struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Version            *string  `json:"version"`
	CompanyRoles       []string `json:"companyRoles"`
	Status             string   `json:"status"`
	RiskClassification *string  `json:"riskClassification"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                string
		flagLimit              int
		flagAfter              string
		flagOrderBy            string
		flagOrderDir           string
		flagStatus             string
		flagRiskClassification string
		flagOutput             *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List AI systems in an organization",
		Aliases: []string{"ls"},
		Example: `  # List AI systems in the default organization
  prb ai-system list

  # List active systems only
  prb ai-system ls --status ACTIVE --json`,
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
					[]string{"CREATED_AT", "NAME", "STATUS"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}

			if flagStatus != "" {
				if err := cmdutil.ValidateEnum(
					"status",
					flagStatus,
					[]string{"ACTIVE", "IN_DEVELOPMENT", "DECOMMISSIONED"},
				); err != nil {
					return err
				}

				filter["status"] = flagStatus
			}

			if flagRiskClassification != "" {
				if err := cmdutil.ValidateEnum(
					"risk-classification",
					flagRiskClassification,
					[]string{"HIGH_RISK", "LIMITED", "MINIMAL", "GPAI"},
				); err != nil {
					return err
				}

				filter["riskClassification"] = flagRiskClassification
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			systems, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[aiSystem]("aiSystems", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				systems,
				totalCount,
				cmdutil.ListView[aiSystem]{
					Empty:      "No AI systems found.",
					Noun:       "AI systems",
					NextCursor: nextCursor,
					Headers: []string{
						"ID",
						"NAME",
						"VERSION",
						"COMPANY ROLES",
						"STATUS",
						"RISK CLASSIFICATION",
					},
					Row: func(system aiSystem) ([]string, error) {
						version := ""
						if system.Version != nil {
							version = *system.Version
						}

						risk := ""
						if system.RiskClassification != nil {
							risk = *system.RiskClassification
						}

						return []string{
							system.ID,
							system.Name,
							version,
							fmt.Sprintf("%v", system.CompanyRoles),
							system.Status,
							risk,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of AI systems to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(
		&flagOrderBy,
		"order-by",
		"",
		"Order by field (CREATED_AT, NAME, STATUS)",
	)
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (ACTIVE, IN_DEVELOPMENT, DECOMMISSIONED)")
	cmd.Flags().StringVar(&flagRiskClassification, "risk-classification", "", "Filter by risk classification (HIGH_RISK, LIMITED, MINIMAL, GPAI)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
