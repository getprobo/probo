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
		flagDataSubjectCategory  string
		flagPersonalDataCategory string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new processing activity",
		Example: `  # Create a processing activity interactively
  prb processing-activity create

  # Create a processing activity non-interactively
  prb pa create --name "Customer onboarding" --role CONTROLLER --lawful-basis CONSENT`,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
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
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagPurpose != "" {
				input["purpose"] = flagPurpose
			}
			if flagRole != "" {
				input["role"] = flagRole
			}
			if flagLawfulBasis != "" {
				input["lawfulBasis"] = flagLawfulBasis
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

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Processing activity name (required)")
	cmd.Flags().StringVar(&flagPurpose, "purpose", "", "Purpose of processing")
	cmd.Flags().StringVar(&flagRole, "role", "", "Role: CONTROLLER, PROCESSOR")
	cmd.Flags().StringVar(&flagLawfulBasis, "lawful-basis", "", "Lawful basis: LEGITIMATE_INTEREST, CONSENT, CONTRACTUAL_NECESSITY, LEGAL_OBLIGATION, VITAL_INTERESTS, PUBLIC_TASK")
	cmd.Flags().StringVar(&flagDataSubjectCategory, "data-subject-category", "", "Data subject category")
	cmd.Flags().StringVar(&flagPersonalDataCategory, "personal-data-category", "", "Personal data category")

	return cmd
}
