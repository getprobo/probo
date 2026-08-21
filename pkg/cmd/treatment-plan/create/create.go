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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateTreatmentPlanInput!) {
  createTreatmentPlan(input: $input) {
    treatmentPlanEdge {
      node {
        id
        treatment
        inherentLikelihood
        inherentImpact
        inherentRiskScore
        residualLikelihood
        residualImpact
        residualRiskScore
      }
    }
  }
}
`

type createResponse struct {
	CreateTreatmentPlan struct {
		TreatmentPlanEdge struct {
			Node struct {
				ID                 string `json:"id"`
				Treatment          string `json:"treatment"`
				InherentLikelihood int    `json:"inherentLikelihood"`
				InherentImpact     int    `json:"inherentImpact"`
				InherentRiskScore  int    `json:"inherentRiskScore"`
				ResidualLikelihood int    `json:"residualLikelihood"`
				ResidualImpact     int    `json:"residualImpact"`
				ResidualRiskScore  int    `json:"residualRiskScore"`
			} `json:"node"`
		} `json:"treatmentPlanEdge"`
	} `json:"createTreatmentPlan"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagRisk               string
		flagRiskAnalysis       string
		flagTreatment          string
		flagOwner              string
		flagInherentLikelihood int
		flagInherentImpact     int
		flagResidualLikelihood int
		flagResidualImpact     int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new treatment plan",
		Example: `  # Create a treatment plan interactively
  prb treatment-plan create --inherent-likelihood 3 --inherent-impact 4

  # Create a treatment plan non-interactively
  prb treatment-plan create --risk <id> --risk-analysis <id> --treatment MITIGATED --owner <id> --inherent-likelihood 3 --inherent-impact 4`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if f.IOStreams.IsInteractive() {
				if flagRisk == "" {
					err := huh.NewInput().
						Title("Risk ID").
						Value(&flagRisk).
						Run()
					if err != nil {
						return err
					}
				}

				if flagRiskAnalysis == "" {
					err := huh.NewInput().
						Title("Risk analysis ID").
						Value(&flagRiskAnalysis).
						Run()
					if err != nil {
						return err
					}
				}

				if flagTreatment == "" {
					err := huh.NewSelect[string]().
						Title("Treatment").
						Options(
							huh.NewOption("Mitigated", "MITIGATED"),
							huh.NewOption("Accepted", "ACCEPTED"),
							huh.NewOption("Avoided", "AVOIDED"),
							huh.NewOption("Transferred", "TRANSFERRED"),
						).
						Value(&flagTreatment).
						Run()
					if err != nil {
						return err
					}
				}

				if flagOwner == "" {
					err := huh.NewInput().
						Title("Owner profile ID").
						Value(&flagOwner).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagRisk == "" {
				return fmt.Errorf("risk is required; pass --risk or run interactively")
			}

			if flagRiskAnalysis == "" {
				return fmt.Errorf("risk analysis is required; pass --risk-analysis or run interactively")
			}

			if flagTreatment == "" {
				return fmt.Errorf("treatment is required; pass --treatment or run interactively")
			}

			if flagOwner == "" {
				return fmt.Errorf("owner is required; pass --owner or run interactively")
			}

			input := map[string]any{
				"riskId":             flagRisk,
				"riskAnalysisId":     flagRiskAnalysis,
				"treatment":          flagTreatment,
				"ownerId":            flagOwner,
				"inherentLikelihood": flagInherentLikelihood,
				"inherentImpact":     flagInherentImpact,
			}

			if cmd.Flags().Changed("residual-likelihood") {
				input["residualLikelihood"] = flagResidualLikelihood
			}

			if cmd.Flags().Changed("residual-impact") {
				input["residualImpact"] = flagResidualImpact
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.CreateTreatmentPlan.TreatmentPlanEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created treatment plan %s\n",
				p.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagRisk, "risk", "", "Risk ID (required)")
	cmd.Flags().StringVar(&flagRiskAnalysis, "risk-analysis", "", "Risk analysis ID (required)")
	cmd.Flags().StringVar(&flagTreatment, "treatment", "", "Treatment: MITIGATED, ACCEPTED, AVOIDED, TRANSFERRED (required)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (required)")
	cmd.Flags().IntVar(&flagInherentLikelihood, "inherent-likelihood", 0, "Inherent likelihood (required)")
	cmd.Flags().IntVar(&flagInherentImpact, "inherent-impact", 0, "Inherent impact (required)")
	cmd.Flags().IntVar(&flagResidualLikelihood, "residual-likelihood", 0, "Residual likelihood")
	cmd.Flags().IntVar(&flagResidualImpact, "residual-impact", 0, "Residual impact")

	_ = cmd.MarkFlagRequired("inherent-likelihood")
	_ = cmd.MarkFlagRequired("inherent-impact")

	return cmd
}
