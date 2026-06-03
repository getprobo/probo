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
mutation($input: CreateThirdPartyInput!) {
  createThirdParty(input: $input) {
    thirdPartyEdge {
      node {
        id
        name
        category
      }
    }
  }
}
`

type createResponse struct {
	CreateThirdParty struct {
		ThirdPartyEdge struct {
			Node struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Category string `json:"category"`
			} `json:"node"`
		} `json:"thirdPartyEdge"`
	} `json:"createThirdParty"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagCategory    string
		flagDescription string
		flagLegalName   string
		flagAddress     string
		flagWebsite     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new thirdParty",
		Example: `  # Create a third_party interactively
  prb third_party create

  # Create a third_party non-interactively
  prb third_party create --name "Acme Corp" --category CLOUD_PROVIDER`,
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
						Title("ThirdParty name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagCategory == "" {
					err := huh.NewSelect[string]().
						Title("ThirdParty category").
						Options(
							huh.NewOption("Analytics", "ANALYTICS"),
							huh.NewOption("Cloud Monitoring", "CLOUD_MONITORING"),
							huh.NewOption("Cloud Provider", "CLOUD_PROVIDER"),
							huh.NewOption("Collaboration", "COLLABORATION"),
							huh.NewOption("Customer Support", "CUSTOMER_SUPPORT"),
							huh.NewOption("Data Storage and Processing", "DATA_STORAGE_AND_PROCESSING"),
							huh.NewOption("Document Management", "DOCUMENT_MANAGEMENT"),
							huh.NewOption("Employee Management", "EMPLOYEE_MANAGEMENT"),
							huh.NewOption("Engineering", "ENGINEERING"),
							huh.NewOption("Finance", "FINANCE"),
							huh.NewOption("Identity Provider", "IDENTITY_PROVIDER"),
							huh.NewOption("IT", "IT"),
							huh.NewOption("Marketing", "MARKETING"),
							huh.NewOption("Office Operations", "OFFICE_OPERATIONS"),
							huh.NewOption("Other", "OTHER"),
							huh.NewOption("Password Management", "PASSWORD_MANAGEMENT"),
							huh.NewOption("Product and Design", "PRODUCT_AND_DESIGN"),
							huh.NewOption("Professional Services", "PROFESSIONAL_SERVICES"),
							huh.NewOption("Recruiting", "RECRUITING"),
							huh.NewOption("Sales", "SALES"),
							huh.NewOption("Security", "SECURITY"),
							huh.NewOption("Version Control", "VERSION_CONTROL"),
						).
						Value(&flagCategory).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagCategory == "" {
				return fmt.Errorf("category is required; pass --category or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"category":       flagCategory,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagLegalName != "" {
				input["legalName"] = flagLegalName
			}

			if flagAddress != "" {
				input["headquarterAddress"] = flagAddress
			}

			if flagWebsite != "" {
				input["websiteUrl"] = flagWebsite
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

			v := resp.CreateThirdParty.ThirdPartyEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created thirdParty %s (%s)\n",
				v.ID,
				v.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "ThirdParty name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "ThirdParty category (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "ThirdParty description")
	cmd.Flags().StringVar(&flagLegalName, "legal-name", "", "Legal name")
	cmd.Flags().StringVar(&flagAddress, "address", "", "Headquarter address")
	cmd.Flags().StringVar(&flagWebsite, "website", "", "Website URL")

	return cmd
}
