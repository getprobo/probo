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
mutation($input: CreateAiSystemInput!) {
  createAiSystem(input: $input) {
    aiSystemEdge {
      node {
        id
        name
        status
      }
    }
  }
}
`

type createResponse struct {
	CreateAiSystem struct {
		AiSystemEdge struct {
			Node struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"node"`
		} `json:"aiSystemEdge"`
	} `json:"createAiSystem"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                     string
		flagName                    string
		flagVersion                 string
		flagCompanyRoles            []string
		flagStatus                  string
		flagOwner                   string
		flagSource                  string
		flagPurpose                 string
		flagIntendedUseCases        string
		flagAutonomyLevel           string
		flagHumanOversightMechanism string
		flagRiskClassification      string
		flagKeyStakeholders         string
		flagDataSourcesAndType      string
		flagDeploymentDate          string
		flagLastReviewDate          string
		flagNextReviewDate          string
		flagNotes                   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new AI system",
		Example: `  # Create an AI system interactively
  prb ai-system create

  # Create non-interactively
  prb ai-system create --name "Support Chatbot" --status ACTIVE`,
		Args: cobra.NoArgs,
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
						Title("Name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagStatus == "" {
					err := huh.NewSelect[string]().
						Title("Status").
						Options(
							huh.NewOption("Active", "ACTIVE"),
							huh.NewOption("In development", "IN_DEVELOPMENT"),
							huh.NewOption("Decommissioned", "DECOMMISSIONED"),
						).
						Value(&flagStatus).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagStatus == "" {
				return fmt.Errorf("status is required; pass --status or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"status",
				flagStatus,
				[]string{"ACTIVE", "IN_DEVELOPMENT", "DECOMMISSIONED"},
			); err != nil {
				return err
			}

			if flagRiskClassification != "" {
				if err := cmdutil.ValidateEnum(
					"risk-classification",
					flagRiskClassification,
					[]string{"HIGH_RISK", "LIMITED", "MINIMAL", "GPAI"},
				); err != nil {
					return err
				}
			}

			for _, role := range flagCompanyRoles {
				if err := cmdutil.ValidateEnum(
					"company-role",
					role,
					[]string{"PROVIDER", "DEPLOYER", "USER", "DEVELOPER"},
				); err != nil {
					return err
				}
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"status":         flagStatus,
			}

			if flagVersion != "" {
				input["version"] = flagVersion
			}

			if len(flagCompanyRoles) > 0 {
				input["companyRoles"] = flagCompanyRoles
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			if flagSource != "" {
				input["source"] = flagSource
			}

			if flagPurpose != "" {
				input["purpose"] = flagPurpose
			}

			if flagIntendedUseCases != "" {
				input["intendedUseCases"] = flagIntendedUseCases
			}

			if flagAutonomyLevel != "" {
				input["autonomyLevel"] = flagAutonomyLevel
			}

			if flagHumanOversightMechanism != "" {
				input["humanOversightMechanism"] = flagHumanOversightMechanism
			}

			if flagRiskClassification != "" {
				input["riskClassification"] = flagRiskClassification
			}

			if flagKeyStakeholders != "" {
				input["keyStakeholders"] = flagKeyStakeholders
			}

			if flagDataSourcesAndType != "" {
				input["dataSourcesAndType"] = flagDataSourcesAndType
			}

			if flagDeploymentDate != "" {
				input["deploymentDate"] = flagDeploymentDate
			}

			if flagLastReviewDate != "" {
				input["lastReviewDate"] = flagLastReviewDate
			}

			if flagNextReviewDate != "" {
				input["nextReviewDate"] = flagNextReviewDate
			}

			if flagNotes != "" {
				input["notes"] = flagNotes
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

			system := resp.CreateAiSystem.AiSystemEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created AI system %s (%s) %s\n",
				system.Name,
				system.Status,
				system.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "System name (required)")
	cmd.Flags().StringVar(&flagVersion, "version", "", "System version")
	cmd.Flags().StringArrayVar(&flagCompanyRoles, "company-role", nil, "Company role: PROVIDER, DEPLOYER, USER, DEVELOPER (repeatable)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Status: ACTIVE, IN_DEVELOPMENT, DECOMMISSIONED (required)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagSource, "source", "", "Source")
	cmd.Flags().StringVar(&flagPurpose, "purpose", "", "Purpose / business goals")
	cmd.Flags().StringVar(&flagIntendedUseCases, "intended-use-cases", "", "Intended use cases")
	cmd.Flags().StringVar(&flagAutonomyLevel, "autonomy-level", "", "Autonomy level")
	cmd.Flags().StringVar(&flagHumanOversightMechanism, "human-oversight-mechanism", "", "Human oversight mechanism")
	cmd.Flags().StringVar(&flagRiskClassification, "risk-classification", "", "Risk classification: HIGH_RISK, LIMITED, MINIMAL, GPAI")
	cmd.Flags().StringVar(&flagKeyStakeholders, "key-stakeholders", "", "Key stakeholders")
	cmd.Flags().StringVar(&flagDataSourcesAndType, "data-sources-and-type", "", "Data sources and type")
	cmd.Flags().StringVar(&flagDeploymentDate, "deployment-date", "", "Deployment date (RFC3339)")
	cmd.Flags().StringVar(&flagLastReviewDate, "last-review-date", "", "Last review date (RFC3339)")
	cmd.Flags().StringVar(&flagNextReviewDate, "next-review-date", "", "Next review date (RFC3339)")
	cmd.Flags().StringVar(&flagNotes, "notes", "", "Notes")

	return cmd
}
