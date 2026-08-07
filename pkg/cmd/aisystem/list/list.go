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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
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
					[]string{
						"CREATED_AT",
						"NAME",
						"STATUS",
						"RISK_CLASSIFICATION",
						"DEPLOYMENT_DATE",
						"LAST_REVIEW_DATE",
						"NEXT_REVIEW_DATE",
					},
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

			systems, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[aiSystem], error) {
					var resp struct {
						Node *struct {
							Typename  string                   `json:"__typename"`
							AiSystems api.Connection[aiSystem] `json:"aiSystems"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}

					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.AiSystems, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, systems)
			}

			if len(systems) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No AI systems found.")
				return nil
			}

			rows := make([][]string, 0, len(systems))
			for _, system := range systems {
				version := ""
				if system.Version != nil {
					version = *system.Version
				}

				risk := ""
				if system.RiskClassification != nil {
					risk = *system.RiskClassification
				}

				rows = append(
					rows,
					[]string{
						system.ID,
						system.Name,
						version,
						fmt.Sprintf("%v", system.CompanyRoles),
						system.Status,
						risk,
					},
				)
			}

			t := cmdutil.NewTable(
				"ID",
				"NAME",
				"VERSION",
				"COMPANY ROLES",
				"STATUS",
				"RISK CLASSIFICATION",
			).Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(systems) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d AI systems\n",
					len(systems),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of AI systems to list")
	cmd.Flags().StringVar(
		&flagOrderBy,
		"order-by",
		"",
		"Order by field (CREATED_AT, NAME, STATUS, RISK_CLASSIFICATION, DEPLOYMENT_DATE, LAST_REVIEW_DATE, NEXT_REVIEW_DATE)",
	)
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (ACTIVE, IN_DEVELOPMENT, DECOMMISSIONED)")
	cmd.Flags().StringVar(&flagRiskClassification, "risk-classification", "", "Filter by risk classification (HIGH_RISK, LIMITED, MINIMAL, GPAI)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
