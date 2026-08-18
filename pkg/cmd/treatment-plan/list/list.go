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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listByOrgQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            treatment
            inherentRiskScore
            residualRiskScore
            netRiskScore
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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder) {
  node(id: $id) {
    __typename
    ... on Risk {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            treatment
            inherentRiskScore
            residualRiskScore
            netRiskScore
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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder) {
  node(id: $id) {
    __typename
    ... on RiskAnalysis {
      treatmentPlans(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            treatment
            inherentRiskScore
            residualRiskScore
            netRiskScore
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
	InherentRiskScore int    `json:"inherentRiskScore"`
	ResidualRiskScore int    `json:"residualRiskScore"`
	NetRiskScore      int    `json:"netRiskScore"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagRisk         string
		flagRiskAnalysis string
		flagLimit        int
		flagOrderBy      string
		flagOrderDir     string
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			query := listByOrgQuery
			parentID := flagOrg
			expectedType := "Organization"
			missingParent := "organization is required; pass --org or set a default with 'prb auth login'"

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

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT", "TREATMENT", "INHERENT_RISK_SCORE", "RESIDUAL_RISK_SCORE"},
				); err != nil {
					return err
				}

				if err := cmdutil.ValidateEnum(
					"order-direction",
					flagOrderDir,
					[]string{"ASC", "DESC"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			plans, totalCount, err := api.Paginate(
				client,
				query,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[treatmentPlan], error) {
					var resp struct {
						Node *struct {
							Typename       string                        `json:"__typename"`
							TreatmentPlans api.Connection[treatmentPlan] `json:"treatmentPlans"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil || resp.Node.Typename != expectedType {
						return nil, fmt.Errorf("%s %s not found", expectedType, parentID)
					}

					return &resp.Node.TreatmentPlans, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, plans)
			}

			if len(plans) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No treatment plans found.")
				return nil
			}

			rows := make([][]string, 0, len(plans))
			for _, p := range plans {
				rows = append(rows, []string{
					p.ID,
					p.Treatment,
					fmt.Sprintf("%d", p.InherentRiskScore),
					fmt.Sprintf("%d", p.ResidualRiskScore),
					fmt.Sprintf("%d", p.NetRiskScore),
				})
			}

			t := cmdutil.NewTable("ID", "TREATMENT", "INHERENT", "RESIDUAL", "NET").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(plans) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d treatment plans\n",
					len(plans),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagRisk, "risk", "", "List plans for a risk ID")
	cmd.Flags().StringVar(&flagRiskAnalysis, "risk-analysis", "", "List plans for a risk analysis ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of treatment plans to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, TREATMENT, INHERENT_RISK_SCORE, RESIDUAL_RISK_SCORE)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
