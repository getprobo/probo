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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateTransferImpactAssessmentInput!) {
  updateTransferImpactAssessment(input: $input) {
    transferImpactAssessment {
      id
      dataSubjects
      legalMechanism
    }
  }
}
`

type updateResponse struct {
	UpdateTransferImpactAssessment struct {
		TransferImpactAssessment struct {
			ID             string `json:"id"`
			DataSubjects   string `json:"dataSubjects"`
			LegalMechanism string `json:"legalMechanism"`
		} `json:"transferImpactAssessment"`
	} `json:"updateTransferImpactAssessment"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDataSubjects          string
		flagLegalMechanism        string
		flagTransfer              string
		flagLocalLawRisk          string
		flagSupplementaryMeasures string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transfer impact assessment",
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

			if cmd.Flags().Changed("data-subjects") {
				input["dataSubjects"] = flagDataSubjects
			}
			if cmd.Flags().Changed("legal-mechanism") {
				input["legalMechanism"] = flagLegalMechanism
			}
			if cmd.Flags().Changed("transfer") {
				input["transfer"] = flagTransfer
			}
			if cmd.Flags().Changed("local-law-risk") {
				input["localLawRisk"] = flagLocalLawRisk
			}
			if cmd.Flags().Changed("supplementary-measures") {
				input["supplementaryMeasures"] = flagSupplementaryMeasures
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

			r := resp.UpdateTransferImpactAssessment.TransferImpactAssessment
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated transfer impact assessment %s\n",
				r.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDataSubjects, "data-subjects", "", "Data subjects")
	cmd.Flags().StringVar(&flagLegalMechanism, "legal-mechanism", "", "Legal mechanism")
	cmd.Flags().StringVar(&flagTransfer, "transfer", "", "Transfer")
	cmd.Flags().StringVar(&flagLocalLawRisk, "local-law-risk", "", "Local law risk")
	cmd.Flags().StringVar(&flagSupplementaryMeasures, "supplementary-measures", "", "Supplementary measures")

	return cmd
}
