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
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listByOrgQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder, $filter: TreatmentPlanFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            treatment
            category
            inherentRiskScore
            residualRiskScore
            netRiskScore
            risk {
              category
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

const listByRiskQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder, $filter: TreatmentPlanFilter) {
  node(id: $id) {
    __typename
    ... on Risk {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            treatment
            category
            inherentRiskScore
            residualRiskScore
            netRiskScore
            risk {
              category
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

const listByAnalysisQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder, $filter: TreatmentPlanFilter, $asOf: Datetime) {
  node(id: $id) {
    __typename
    ... on RiskAnalysis {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy, filter: $filter, asOf: $asOf) {
        totalCount
        edges {
          node {
            id
            treatment
            category
            inherentRiskScore
            residualRiskScore
            netRiskScore
            risk {
              category
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

type treatmentPlan struct {
	ID                string `json:"id"`
	Treatment         string `json:"treatment"`
	Category          string `json:"category"`
	InherentRiskScore int    `json:"inherentRiskScore"`
	ResidualRiskScore int    `json:"residualRiskScore"`
	NetRiskScore      int    `json:"netRiskScore"`
	Risk              struct {
		Category string `json:"category"`
	} `json:"risk"`
}

func (p treatmentPlan) tableCategory() string {
	if p.Category != "" {
		return p.Category
	}

	return p.Risk.Category
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagRisk         string
		flagRiskAnalysis string
		flagLimit        int
		flagAfter        string
		flagOrderBy      string
		flagOrderDir     string
		flagScoreType    string
		flagLikelihood   int
		flagImpact       int
		flagAsOf         string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List treatment plans",
		Aliases: []string{"ls"},
		Example: `  # List treatment plans in the default organization
  prb treatment-plan list

  # List treatment plans for a risk analysis
  prb treatment-plan ls --risk-analysis <id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			if flagRisk != "" && flagRiskAnalysis != "" {
				return fmt.Errorf("pass only one of --risk and --risk-analysis")
			}

			if flagAsOf != "" && flagRiskAnalysis == "" {
				return fmt.Errorf("--as-of requires --risk-analysis")
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			query := listByOrgQuery
			parentID := flagOrg
			expectedType := "Organization"
			missingParent := cmdutil.OrganizationRequiredMessage

			if flagRisk != "" {
				query = listByRiskQuery
				parentID = flagRisk
				expectedType = "Risk"
				missingParent = "risk is required; pass --risk"
			} else if flagRiskAnalysis != "" {
				query = listByAnalysisQuery
				parentID = flagRiskAnalysis
				expectedType = "RiskAnalysis"
				missingParent = "risk analysis is required; pass --risk-analysis"
			} else if parentID == "" {
				parentID = hc.Organization
			}

			if parentID == "" {
				return fmt.Errorf("%s", missingParent)
			}

			variables := map[string]any{
				"id": parentID,
			}

			if flagAsOf != "" {
				if _, err := time.Parse(time.RFC3339, flagAsOf); err != nil {
					return fmt.Errorf(
						"--as-of must be RFC3339 (e.g. 2026-01-15T23:59:59Z): %w",
						err,
					)
				}

				variables["asOf"] = flagAsOf
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT", "TREATMENT", "CATEGORY", "INHERENT_RISK_SCORE", "RESIDUAL_RISK_SCORE"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			hasCell := flagScoreType != "" || flagLikelihood != 0 || flagImpact != 0
			if hasCell {
				if flagScoreType == "" || flagLikelihood == 0 || flagImpact == 0 {
					return fmt.Errorf("--score-type, --likelihood, and --impact must be set together")
				}

				if err := cmdutil.ValidateEnum(
					"score-type",
					flagScoreType,
					[]string{"INHERENT", "NET", "RESIDUAL"},
				); err != nil {
					return err
				}

				variables["filter"] = map[string]any{
					"scoreType":  flagScoreType,
					"likelihood": flagLikelihood,
					"impact":     flagImpact,
				}
			}

			plans, totalCount, nextCursor, err := api.Paginate(
				client,
				query,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnectionOrMissing[treatmentPlan]("treatmentPlans", expectedType, fmt.Sprintf("%s %s not found", expectedType, parentID)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				plans,
				totalCount,
				cmdutil.ListView[treatmentPlan]{
					Empty:      "No treatment plans found.",
					Noun:       "treatment plans",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TREATMENT", "CATEGORY", "INHERENT", "RESIDUAL", "NET"},
					Row: func(p treatmentPlan) ([]string, error) {
						return []string{
							p.ID,
							p.Treatment,
							p.tableCategory(),
							fmt.Sprintf("%d", p.InherentRiskScore),
							fmt.Sprintf("%d", p.ResidualRiskScore),
							fmt.Sprintf("%d", p.NetRiskScore),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().StringVar(&flagRisk, "risk", "", "List plans for a risk ID")
	cmd.Flags().StringVar(&flagRiskAnalysis, "risk-analysis", "", "List plans for a risk analysis ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of treatment plans to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, TREATMENT, CATEGORY, INHERENT_RISK_SCORE, RESIDUAL_RISK_SCORE)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagScoreType, "score-type", "", "Filter by matrix score type (INHERENT, NET, RESIDUAL)")
	cmd.Flags().IntVar(&flagLikelihood, "likelihood", 0, "Filter by likelihood (with --score-type and --impact)")
	cmd.Flags().IntVar(&flagImpact, "impact", 0, "Filter by impact (with --score-type and --likelihood)")
	cmd.Flags().StringVar(
		&flagAsOf,
		"as-of",
		"",
		"Reconstruct plans as of this RFC3339 instant (requires --risk-analysis; omit for live tables)",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
