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
mutation($input: CreateProcessingActivityInput!) {
  createProcessingActivity(input: $input) {
    processingActivityEdge {
      node {
        id
        name
        role
        lawfulBasis
      }
    }
  }
}
`

type createResponse struct {
	CreateProcessingActivity struct {
		ProcessingActivityEdge struct {
			Node struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Role        string `json:"role"`
				LawfulBasis string `json:"lawfulBasis"`
			} `json:"node"`
		} `json:"processingActivityEdge"`
	} `json:"createProcessingActivity"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                  string
		flagName                 string
		flagPurpose              string
		flagRole                 string
		flagLawfulBasis          string
		flagSpecialOrCriminal    string
		flagInternational        bool
		flagDPIA                 string
		flagTIA                  string
		flagDataSubjectCategory  string
		flagPersonalDataCategory string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new processing activity",
		Example: `  # Create a processing activity interactively
  prb processing-activity create

  # Create a processing activity non-interactively
  prb pa create --name "Customer onboarding" --role CONTROLLER --lawful-basis CONSENT --special-or-criminal-data NO --data-protection-impact-assessment NOT_NEEDED --transfer-impact-assessment NOT_NEEDED`,
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

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					err := huh.NewInput().
						Title("Processing activity name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagRole == "" {
					err := huh.NewSelect[string]().
						Title("Role").
						Options(
							huh.NewOption("Controller", "CONTROLLER"),
							huh.NewOption("Processor", "PROCESSOR"),
						).
						Value(&flagRole).
						Run()
					if err != nil {
						return err
					}
				}

				if flagLawfulBasis == "" {
					err := huh.NewSelect[string]().
						Title("Lawful basis").
						Options(
							huh.NewOption("Legitimate interest", "LEGITIMATE_INTEREST"),
							huh.NewOption("Consent", "CONSENT"),
							huh.NewOption("Contractual necessity", "CONTRACTUAL_NECESSITY"),
							huh.NewOption("Legal obligation", "LEGAL_OBLIGATION"),
							huh.NewOption("Vital interests", "VITAL_INTERESTS"),
							huh.NewOption("Public task", "PUBLIC_TASK"),
						).
						Value(&flagLawfulBasis).
						Run()
					if err != nil {
						return err
					}
				}

				if flagSpecialOrCriminal == "" {
					err := huh.NewSelect[string]().
						Title("Special or criminal data").
						Options(
							huh.NewOption("No", "NO"),
							huh.NewOption("Yes", "YES"),
							huh.NewOption("Possible", "POSSIBLE"),
						).
						Value(&flagSpecialOrCriminal).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDPIA == "" {
					err := huh.NewSelect[string]().
						Title("Data protection impact assessment").
						Options(
							huh.NewOption("Not needed", "NOT_NEEDED"),
							huh.NewOption("Needed", "NEEDED"),
						).
						Value(&flagDPIA).
						Run()
					if err != nil {
						return err
					}
				}

				if flagTIA == "" {
					err := huh.NewSelect[string]().
						Title("Transfer impact assessment").
						Options(
							huh.NewOption("Not needed", "NOT_NEEDED"),
							huh.NewOption("Needed", "NEEDED"),
						).
						Value(&flagTIA).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagRole == "" {
				return fmt.Errorf("role is required; pass --role or run interactively")
			}

			if err := cmdutil.ValidateEnum("role", flagRole, []string{"CONTROLLER", "PROCESSOR"}); err != nil {
				return err
			}

			if flagLawfulBasis == "" {
				return fmt.Errorf("lawful basis is required; pass --lawful-basis or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"lawful-basis",
				flagLawfulBasis,
				[]string{
					"LEGITIMATE_INTEREST",
					"CONSENT",
					"CONTRACTUAL_NECESSITY",
					"LEGAL_OBLIGATION",
					"VITAL_INTERESTS",
					"PUBLIC_TASK",
				},
			); err != nil {
				return err
			}

			if flagSpecialOrCriminal == "" {
				return fmt.Errorf("special or criminal data is required; pass --special-or-criminal-data or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"special-or-criminal-data",
				flagSpecialOrCriminal,
				[]string{"YES", "NO", "POSSIBLE"},
			); err != nil {
				return err
			}

			if flagDPIA == "" {
				return fmt.Errorf("data protection impact assessment is required; pass --data-protection-impact-assessment or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"data-protection-impact-assessment",
				flagDPIA,
				[]string{"NEEDED", "NOT_NEEDED"},
			); err != nil {
				return err
			}

			if flagTIA == "" {
				return fmt.Errorf("transfer impact assessment is required; pass --transfer-impact-assessment or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"transfer-impact-assessment",
				flagTIA,
				[]string{"NEEDED", "NOT_NEEDED"},
			); err != nil {
				return err
			}

			input := map[string]any{
				"organizationId":                       flagOrg,
				"name":                                 flagName,
				"role":                                 flagRole,
				"lawfulBasis":                          flagLawfulBasis,
				"specialOrCriminalData":                flagSpecialOrCriminal,
				"internationalTransfers":               flagInternational,
				"dataProtectionImpactAssessmentNeeded": flagDPIA,
				"transferImpactAssessmentNeeded":       flagTIA,
			}

			if flagPurpose != "" {
				input["purpose"] = flagPurpose
			}

			if flagDataSubjectCategory != "" {
				input["dataSubjectCategory"] = flagDataSubjectCategory
			}

			if flagPersonalDataCategory != "" {
				input["personalDataCategory"] = flagPersonalDataCategory
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

			a := resp.CreateProcessingActivity.ProcessingActivityEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created processing activity %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Processing activity name (required)")
	cmd.Flags().StringVar(&flagPurpose, "purpose", "", "Purpose of processing")
	cmd.Flags().StringVar(&flagRole, "role", "", "Role: CONTROLLER, PROCESSOR (required)")
	cmd.Flags().StringVar(&flagLawfulBasis, "lawful-basis", "", "Lawful basis: LEGITIMATE_INTEREST, CONSENT, CONTRACTUAL_NECESSITY, LEGAL_OBLIGATION, VITAL_INTERESTS, PUBLIC_TASK (required)")
	cmd.Flags().StringVar(&flagSpecialOrCriminal, "special-or-criminal-data", "", "Special or criminal data: YES, NO, POSSIBLE (required)")
	cmd.Flags().BoolVar(&flagInternational, "international-transfers", false, "Processing involves international transfers")
	cmd.Flags().StringVar(&flagDPIA, "data-protection-impact-assessment", "", "Data protection impact assessment: NEEDED, NOT_NEEDED (required)")
	cmd.Flags().StringVar(&flagTIA, "transfer-impact-assessment", "", "Transfer impact assessment: NEEDED, NOT_NEEDED (required)")
	cmd.Flags().StringVar(&flagDataSubjectCategory, "data-subject-category", "", "Data subject category")
	cmd.Flags().StringVar(&flagPersonalDataCategory, "personal-data-category", "", "Personal data category")

	return cmd
}
