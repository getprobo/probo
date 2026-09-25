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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateTreatmentPlanInput!) {
  updateTreatmentPlan(input: $input) {
    treatmentPlan {
      id
      treatment
      inherentRiskScore
      residualRiskScore
      netRiskScore
    }
  }
}
`

type updateResponse struct {
	UpdateTreatmentPlan struct {
		TreatmentPlan struct {
			ID                string `json:"id"`
			Treatment         string `json:"treatment"`
			InherentRiskScore int    `json:"inherentRiskScore"`
			ResidualRiskScore int    `json:"residualRiskScore"`
			NetRiskScore      int    `json:"netRiskScore"`
		} `json:"treatmentPlan"`
	} `json:"updateTreatmentPlan"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTreatment          string
		flagOwner              string
		flagInherentLikelihood int
		flagInherentImpact     int
		flagResidualLikelihood int
		flagResidualImpact     int
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a treatment plan",
		Args:  cobra.ExactArgs(1),
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

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("treatment") {
				input["treatment"] = flagTreatment
			}

			if cmd.Flags().Changed("inherent-likelihood") {
				input["inherentLikelihood"] = flagInherentLikelihood
			}

			if cmd.Flags().Changed("inherent-impact") {
				input["inherentImpact"] = flagInherentImpact
			}

			if cmd.Flags().Changed("residual-likelihood") {
				input["residualLikelihood"] = flagResidualLikelihood
			}

			if cmd.Flags().Changed("residual-impact") {
				input["residualImpact"] = flagResidualImpact
			}

			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					return fmt.Errorf("owner is required; pass a profile ID")
				}

				input["ownerId"] = flagOwner
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.UpdateTreatmentPlan.TreatmentPlan
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated treatment plan %s\n",
				p.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTreatment, "treatment", "", "Treatment: MITIGATED, ACCEPTED, AVOIDED, TRANSFERRED")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().IntVar(&flagInherentLikelihood, "inherent-likelihood", 0, "Inherent likelihood")
	cmd.Flags().IntVar(&flagInherentImpact, "inherent-impact", 0, "Inherent impact")
	cmd.Flags().IntVar(&flagResidualLikelihood, "residual-likelihood", 0, "Residual likelihood")
	cmd.Flags().IntVar(&flagResidualImpact, "residual-impact", 0, "Residual impact")

	return cmd
}
