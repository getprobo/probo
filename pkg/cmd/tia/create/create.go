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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateTransferImpactAssessmentInput!) {
  createTransferImpactAssessment(input: $input) {
    transferImpactAssessmentEdge {
      node {
        id
        dataSubjects
        legalMechanism
      }
    }
  }
}
`

type createResponse struct {
	CreateTransferImpactAssessment struct {
		TransferImpactAssessmentEdge struct {
			Node struct {
				ID             string `json:"id"`
				DataSubjects   string `json:"dataSubjects"`
				LegalMechanism string `json:"legalMechanism"`
			} `json:"node"`
		} `json:"transferImpactAssessmentEdge"`
	} `json:"createTransferImpactAssessment"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProcessingActivity    string
		flagDataSubjects          string
		flagLegalMechanism        string
		flagTransfer              string
		flagLocalLawRisk          string
		flagSupplementaryMeasures string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new transfer impact assessment",
		Example: `  # Create a TIA
  prb tia create --processing-activity <id> --data-subjects "EU residents"

  # Create a TIA with all fields
  prb tia create --processing-activity <id> --data-subjects "EU residents" --legal-mechanism "SCCs" --transfer "US" --local-law-risk "FISA 702" --supplementary-measures "Encryption"`,
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

			input := map[string]any{
				"processingActivityId": flagProcessingActivity,
			}

			if flagDataSubjects != "" {
				input["dataSubjects"] = flagDataSubjects
			}

			if flagLegalMechanism != "" {
				input["legalMechanism"] = flagLegalMechanism
			}

			if flagTransfer != "" {
				input["transfer"] = flagTransfer
			}

			if flagLocalLawRisk != "" {
				input["localLawRisk"] = flagLocalLawRisk
			}

			if flagSupplementaryMeasures != "" {
				input["supplementaryMeasures"] = flagSupplementaryMeasures
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

			r := resp.CreateTransferImpactAssessment.TransferImpactAssessmentEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created transfer impact assessment %s\n",
				r.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProcessingActivity, "processing-activity", "", "Processing activity ID (required)")
	cmd.Flags().StringVar(&flagDataSubjects, "data-subjects", "", "Data subjects")
	cmd.Flags().StringVar(&flagLegalMechanism, "legal-mechanism", "", "Legal mechanism")
	cmd.Flags().StringVar(&flagTransfer, "transfer", "", "Transfer")
	cmd.Flags().StringVar(&flagLocalLawRisk, "local-law-risk", "", "Local law risk")
	cmd.Flags().StringVar(&flagSupplementaryMeasures, "supplementary-measures", "", "Supplementary measures")

	_ = cmd.MarkFlagRequired("processing-activity")

	return cmd
}
