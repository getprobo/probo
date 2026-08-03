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

package get

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const getQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on Organization {
      malaysiaPDPAProfile {
        organizationId
        totalDataSubjects
        sensitiveDataSubjects
        regularSystematicMonitoring
        dpoRequired
        dpoRequirementReasons
        assessedByProfileId
        assessedAt
        dpoProfileId
        dpoAppointedAt
        commissionerNotificationDueAt
        commissionerNotifiedAt
        commissionerNotificationReference
        createdAt
        updatedAt
      }
    }
  }
}
`

type getResponse struct {
	Node *struct {
		MalaysiaPDPAProfile *struct {
			OrganizationID                    string   `json:"organizationId"`
			TotalDataSubjects                 int64    `json:"totalDataSubjects"`
			SensitiveDataSubjects             int64    `json:"sensitiveDataSubjects"`
			RegularSystematicMonitoring       bool     `json:"regularSystematicMonitoring"`
			DPORequired                       bool     `json:"dpoRequired"`
			DPORequirementReasons             []string `json:"dpoRequirementReasons"`
			AssessedByProfileID               *string  `json:"assessedByProfileId"`
			AssessedAt                        *string  `json:"assessedAt"`
			DPOProfileID                      *string  `json:"dpoProfileId"`
			DPOAppointedAt                    *string  `json:"dpoAppointedAt"`
			CommissionerNotificationDueAt     *string  `json:"commissionerNotificationDueAt"`
			CommissionerNotifiedAt            *string  `json:"commissionerNotifiedAt"`
			CommissionerNotificationReference *string  `json:"commissionerNotificationReference"`
			CreatedAt                         *string  `json:"createdAt"`
			UpdatedAt                         *string  `json:"updatedAt"`
		} `json:"malaysiaPDPAProfile"`
	} `json:"node"`
}

func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get the Malaysia PDPA profile",
		Example: `  prb malaysia-pdpa get --org <org-id>`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
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

			data, err := client.Do(getQuery, map[string]any{"id": organizationID})
			if err != nil {
				return err
			}

			var response getResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if response.Node == nil || response.Node.MalaysiaPDPAProfile == nil {
				return fmt.Errorf("organization %s not found", organizationID)
			}

			profile := response.Node.MalaysiaPDPAProfile
			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, profile)
			}

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(36)
			out := f.IOStreams.Out

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Malaysia PDPA profile"))
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Total data subjects:"), profile.TotalDataSubjects)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Sensitive/financial data subjects:"), profile.SensitiveDataSubjects)
			_, _ = fmt.Fprintf(out, "%s%t\n", label.Render("Regular systematic monitoring:"), profile.RegularSystematicMonitoring)
			_, _ = fmt.Fprintf(out, "%s%t\n", label.Render("DPO required:"), profile.DPORequired)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("DPO reasons:"), strings.Join(profile.DPORequirementReasons, ", "))

			if profile.AssessedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Assessed at:"), cmdutil.FormatTime(*profile.AssessedAt))
			}

			if profile.DPOProfileID != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("DPO profile ID:"), *profile.DPOProfileID)
			}

			if profile.CommissionerNotificationDueAt != nil {
				_, _ = fmt.Fprintf(
					out,
					"%s%s\n",
					label.Render("Commissioner notification due:"),
					cmdutil.FormatTime(*profile.CommissionerNotificationDueAt),
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
