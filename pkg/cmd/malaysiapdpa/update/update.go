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
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const (
	getProfileQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on Organization {
      malaysiaPDPAProfile {
        dpoProfileId
        dpoAppointedAt
        commissionerNotifiedAt
        commissionerNotificationReference
      }
    }
  }
}
`

	updateMutation = `
mutation($input: UpdateMalaysiaPDPAProfileInput!) {
  updateMalaysiaPDPAProfile(input: $input) {
    malaysiaPDPAProfile {
      organizationId
      dpoRequired
      dpoRequirementReasons
      assessedAt
    }
  }
}
`
)

type (
	currentProfileResponse struct {
		Node *struct {
			MalaysiaPDPAProfile *struct {
				DPOProfileID                      *string    `json:"dpoProfileId"`
				DPOAppointedAt                    *time.Time `json:"dpoAppointedAt"`
				CommissionerNotifiedAt            *time.Time `json:"commissionerNotifiedAt"`
				CommissionerNotificationReference *string    `json:"commissionerNotificationReference"`
			} `json:"malaysiaPDPAProfile"`
		} `json:"node"`
	}

	updateResponse struct {
		UpdateMalaysiaPDPAProfile struct {
			MalaysiaPDPAProfile struct {
				OrganizationID        string   `json:"organizationId"`
				DPORequired           bool     `json:"dpoRequired"`
				DPORequirementReasons []string `json:"dpoRequirementReasons"`
				AssessedAt            string   `json:"assessedAt"`
			} `json:"malaysiaPDPAProfile"`
		} `json:"updateMalaysiaPDPAProfile"`
	}
)

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                               string
		flagTotalDataSubjects                 int64
		flagSensitiveDataSubjects             int64
		flagRegularSystematicMonitoring       bool
		flagDPOProfileID                      string
		flagDPOAppointedAt                    string
		flagCommissionerNotifiedAt            string
		flagCommissionerNotificationReference string
		flagClearDPO                          bool
		flagClearCommissionerNotification     bool
	)

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update the Malaysia PDPA profile",
		Example: `  prb malaysia-pdpa update --org <org-id> --total-data-subjects 25000 --sensitive-data-subjects 2000 --regular-systematic-monitoring=false`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{
				"total-data-subjects",
				"sensitive-data-subjects",
				"regular-systematic-monitoring",
			} {
				if !cmd.Flags().Changed(name) {
					return fmt.Errorf("--%s is required", name)
				}
			}

			if flagTotalDataSubjects < 0 || flagSensitiveDataSubjects < 0 {
				return fmt.Errorf("data-subject counts cannot be negative")
			}

			if flagSensitiveDataSubjects > flagTotalDataSubjects {
				return fmt.Errorf("--sensitive-data-subjects cannot exceed --total-data-subjects")
			}

			if cmd.Flags().Changed("dpo-profile-id") != cmd.Flags().Changed("dpo-appointed-at") {
				return fmt.Errorf("--dpo-profile-id and --dpo-appointed-at must be provided together")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			organizationID := flagOrg
			if organizationID == "" {
				organizationID = hc.Organization
			}

			if organizationID == "" {
				return fmt.Errorf("organization ID is required: pass --org or run `prb auth login`")
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			currentData, err := client.Do(getProfileQuery, map[string]any{"id": organizationID})
			if err != nil {
				return err
			}

			var current currentProfileResponse
			if err := json.Unmarshal(currentData, &current); err != nil {
				return fmt.Errorf("cannot parse current profile: %w", err)
			}

			if current.Node == nil || current.Node.MalaysiaPDPAProfile == nil {
				return fmt.Errorf("organization %s not found", organizationID)
			}

			profile := current.Node.MalaysiaPDPAProfile
			input := map[string]any{
				"organizationId":              organizationID,
				"totalDataSubjects":           flagTotalDataSubjects,
				"sensitiveDataSubjects":       flagSensitiveDataSubjects,
				"regularSystematicMonitoring": flagRegularSystematicMonitoring,
			}

			if !flagClearDPO {
				if profile.DPOProfileID != nil {
					input["dpoProfileId"] = *profile.DPOProfileID
				}
				if profile.DPOAppointedAt != nil {
					input["dpoAppointedAt"] = profile.DPOAppointedAt.Format(time.RFC3339)
				}
				if profile.CommissionerNotifiedAt != nil && !flagClearCommissionerNotification {
					input["commissionerNotifiedAt"] = profile.CommissionerNotifiedAt.Format(time.RFC3339)
				}
				if profile.CommissionerNotificationReference != nil && !flagClearCommissionerNotification {
					input["commissionerNotificationReference"] = *profile.CommissionerNotificationReference
				}
			}

			if cmd.Flags().Changed("dpo-profile-id") {
				appointedAt, err := time.Parse(time.RFC3339, flagDPOAppointedAt)
				if err != nil {
					return fmt.Errorf("invalid --dpo-appointed-at: use RFC3339 format: %w", err)
				}

				input["dpoProfileId"] = flagDPOProfileID
				input["dpoAppointedAt"] = appointedAt.Format(time.RFC3339)
			}

			if cmd.Flags().Changed("commissioner-notified-at") {
				notifiedAt, err := time.Parse(time.RFC3339, flagCommissionerNotifiedAt)
				if err != nil {
					return fmt.Errorf("invalid --commissioner-notified-at: use RFC3339 format: %w", err)
				}

				input["commissionerNotifiedAt"] = notifiedAt.Format(time.RFC3339)
			}

			if cmd.Flags().Changed("commissioner-notification-reference") {
				input["commissionerNotificationReference"] = flagCommissionerNotificationReference
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var response updateResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			updated := response.UpdateMalaysiaPDPAProfile.MalaysiaPDPAProfile
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated Malaysia PDPA profile for organization %s (DPO required: %t)\n",
				updated.OrganizationID,
				updated.DPORequired,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().Int64Var(&flagTotalDataSubjects, "total-data-subjects", 0, "Estimated total number of data subjects")
	cmd.Flags().Int64Var(&flagSensitiveDataSubjects, "sensitive-data-subjects", 0, "Estimated number of sensitive or financial data subjects")
	cmd.Flags().BoolVar(&flagRegularSystematicMonitoring, "regular-systematic-monitoring", false, "Whether processing includes regular and systematic monitoring")
	cmd.Flags().StringVar(&flagDPOProfileID, "dpo-profile-id", "", "Appointed DPO membership profile ID")
	cmd.Flags().StringVar(&flagDPOAppointedAt, "dpo-appointed-at", "", "DPO appointment time in RFC3339 format")
	cmd.Flags().StringVar(&flagCommissionerNotifiedAt, "commissioner-notified-at", "", "Commissioner notification time in RFC3339 format")
	cmd.Flags().StringVar(&flagCommissionerNotificationReference, "commissioner-notification-reference", "", "Commissioner notification evidence or reference")
	cmd.Flags().BoolVar(&flagClearDPO, "clear-dpo", false, "Clear the DPO appointment and Commissioner notification")
	cmd.Flags().BoolVar(&flagClearCommissionerNotification, "clear-commissioner-notification", false, "Clear the Commissioner notification while preserving the DPO appointment")
	cmd.MarkFlagsMutuallyExclusive("clear-dpo", "dpo-profile-id")
	cmd.MarkFlagsMutuallyExclusive("clear-dpo", "dpo-appointed-at")
	cmd.MarkFlagsMutuallyExclusive("clear-dpo", "commissioner-notified-at")
	cmd.MarkFlagsMutuallyExclusive("clear-dpo", "commissioner-notification-reference")
	cmd.MarkFlagsMutuallyExclusive("clear-commissioner-notification", "commissioner-notified-at")
	cmd.MarkFlagsMutuallyExclusive("clear-commissioner-notification", "commissioner-notification-reference")

	return cmd
}
