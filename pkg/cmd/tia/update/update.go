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
		flagDataSubjects                       string
		flagLegalMechanism                     string
		flagTransfer                           string
		flagLocalLawRisk                       string
		flagSupplementaryMeasures              string
		flagMalaysiaBasis                      string
		flagMalaysiaDestinationCountry         string
		flagMalaysiaRecipientThirdParty        string
		flagMalaysiaReceiverRegistrationNumber string
		flagMalaysiaReceiverContact            string
		flagMalaysiaTransferPurpose            string
		flagMalaysiaPersonalDataCategories     string
		flagMalaysiaSafeguards                 string
		flagMalaysiaApprovalStatus             string
		flagMalaysiaApprovalNotes              string
		flagMalaysiaReviewEvidence             string
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

			malaysiaFlags := []string{
				"malaysia-basis",
				"malaysia-destination-country",
				"malaysia-recipient-third-party",
				"malaysia-receiver-registration-number",
				"malaysia-receiver-contact",
				"malaysia-transfer-purpose",
				"malaysia-personal-data-categories",
				"malaysia-safeguards",
				"malaysia-approval-status",
				"malaysia-approval-notes",
				"malaysia-review-evidence",
			}
			malaysiaChanged := false
			for _, name := range malaysiaFlags {
				malaysiaChanged = malaysiaChanged || cmd.Flags().Changed(name)
			}
			if malaysiaChanged {
				required := map[string]string{
					"malaysia-basis":                    flagMalaysiaBasis,
					"malaysia-destination-country":      flagMalaysiaDestinationCountry,
					"malaysia-recipient-third-party":    flagMalaysiaRecipientThirdParty,
					"malaysia-receiver-contact":         flagMalaysiaReceiverContact,
					"malaysia-transfer-purpose":         flagMalaysiaTransferPurpose,
					"malaysia-personal-data-categories": flagMalaysiaPersonalDataCategories,
					"malaysia-safeguards":               flagMalaysiaSafeguards,
					"malaysia-approval-status":          flagMalaysiaApprovalStatus,
				}
				for name, value := range required {
					if value == "" {
						return fmt.Errorf("--%s is required for a Malaysia PDPA transfer record", name)
					}
				}

				input["malaysiaPDPA"] = map[string]any{
					"basis":                      flagMalaysiaBasis,
					"destinationCountry":         flagMalaysiaDestinationCountry,
					"recipientThirdPartyId":      flagMalaysiaRecipientThirdParty,
					"receiverRegistrationNumber": flagMalaysiaReceiverRegistrationNumber,
					"receiverContact":            flagMalaysiaReceiverContact,
					"transferPurpose":            flagMalaysiaTransferPurpose,
					"personalDataCategories":     flagMalaysiaPersonalDataCategories,
					"safeguards":                 flagMalaysiaSafeguards,
					"approvalStatus":             flagMalaysiaApprovalStatus,
					"approvalNotes":              flagMalaysiaApprovalNotes,
					"reviewEvidence":             flagMalaysiaReviewEvidence,
				}
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
	cmd.Flags().StringVar(&flagMalaysiaBasis, "malaysia-basis", "", "Malaysia PDPA section 129 transfer basis")
	cmd.Flags().StringVar(&flagMalaysiaDestinationCountry, "malaysia-destination-country", "", "Foreign destination country code")
	cmd.Flags().StringVar(&flagMalaysiaRecipientThirdParty, "malaysia-recipient-third-party", "", "Recipient third-party ID")
	cmd.Flags().StringVar(&flagMalaysiaReceiverRegistrationNumber, "malaysia-receiver-registration-number", "", "Receiver company registration number")
	cmd.Flags().StringVar(&flagMalaysiaReceiverContact, "malaysia-receiver-contact", "", "Receiver DPO or responsible contact")
	cmd.Flags().StringVar(&flagMalaysiaTransferPurpose, "malaysia-transfer-purpose", "", "Purpose of the cross-border transfer")
	cmd.Flags().StringVar(&flagMalaysiaPersonalDataCategories, "malaysia-personal-data-categories", "", "Personal data categories transferred")
	cmd.Flags().StringVar(&flagMalaysiaSafeguards, "malaysia-safeguards", "", "Transfer safeguards and evidence")
	cmd.Flags().StringVar(&flagMalaysiaApprovalStatus, "malaysia-approval-status", "", "Approval status: PENDING, APPROVED, REJECTED")
	cmd.Flags().StringVar(&flagMalaysiaApprovalNotes, "malaysia-approval-notes", "", "Approval or rejection notes")
	cmd.Flags().StringVar(&flagMalaysiaReviewEvidence, "malaysia-review-evidence", "", "Review evidence required for approval")

	return cmd
}
