// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
mutation($input: CreateDataProtectionImpactAssessmentInput!) {
  createDataProtectionImpactAssessment(input: $input) {
    dataProtectionImpactAssessmentEdge {
      node {
        id
        description
        residualRisk
      }
    }
  }
}
`

type createResponse struct {
	CreateDataProtectionImpactAssessment struct {
		DataProtectionImpactAssessmentEdge struct {
			Node struct {
				ID           string `json:"id"`
				Description  string `json:"description"`
				ResidualRisk string `json:"residualRisk"`
			} `json:"node"`
		} `json:"dataProtectionImpactAssessmentEdge"`
	} `json:"createDataProtectionImpactAssessment"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProcessingActivity          string
		flagDescription                 string
		flagNecessityAndProportionality string
		flagPotentialRisk               string
		flagMitigations                 string
		flagResidualRisk                string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new data protection impact assessment",
		Example: `  # Create a DPIA
  prb dpia create --processing-activity <id> --description "Assessment for HR processing"

  # Create a DPIA with all fields
  prb dpia create --processing-activity <id> --description "Assessment" --necessity "Required by law" --potential-risk "Data leak" --mitigations "Encryption" --residual-risk LOW`,
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

			if flagProcessingActivity == "" {
				return fmt.Errorf("processing activity is required; pass --processing-activity")
			}

			if f.IOStreams.IsInteractive() && flagResidualRisk == "" {
				err := huh.NewSelect[string]().
					Title("Residual risk").
					Options(
						huh.NewOption("Low", "LOW"),
						huh.NewOption("Medium", "MEDIUM"),
						huh.NewOption("High", "HIGH"),
					).
					Value(&flagResidualRisk).
					Run()
				if err != nil {
					return err
				}
			}

			input := map[string]any{
				"processingActivityId": flagProcessingActivity,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if flagNecessityAndProportionality != "" {
				input["necessityAndProportionality"] = flagNecessityAndProportionality
			}
			if flagPotentialRisk != "" {
				input["potentialRisk"] = flagPotentialRisk
			}
			if flagMitigations != "" {
				input["mitigations"] = flagMitigations
			}
			if flagResidualRisk != "" {
				input["residualRisk"] = flagResidualRisk
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

			r := resp.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessmentEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created data protection impact assessment %s\n",
				r.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProcessingActivity, "processing-activity", "", "Processing activity ID (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Description")
	cmd.Flags().StringVar(&flagNecessityAndProportionality, "necessity", "", "Necessity and proportionality")
	cmd.Flags().StringVar(&flagPotentialRisk, "potential-risk", "", "Potential risk")
	cmd.Flags().StringVar(&flagMitigations, "mitigations", "", "Mitigations")
	cmd.Flags().StringVar(&flagResidualRisk, "residual-risk", "", "Residual risk: LOW, MEDIUM, HIGH")

	_ = cmd.MarkFlagRequired("processing-activity")

	return cmd
}
