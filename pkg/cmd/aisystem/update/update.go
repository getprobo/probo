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
mutation($input: UpdateAiSystemInput!) {
  updateAiSystem(input: $input) {
    aiSystem {
      id
      name
      status
    }
  }
}
`

type updateResponse struct {
	UpdateAiSystem struct {
		AiSystem struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"aiSystem"`
	} `json:"updateAiSystem"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
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
		Use:   "update <id>",
		Short: "Update an AI system",
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

			if flagName != "" {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("version") {
				input["version"] = nilString(flagVersion)
			}

			if len(flagCompanyRoles) > 0 {
				for _, role := range flagCompanyRoles {
					if err := cmdutil.ValidateEnum(
						"company-role",
						role,
						[]string{"PROVIDER", "DEPLOYER", "USER", "DEVELOPER"},
					); err != nil {
						return err
					}
				}

				input["companyRoles"] = flagCompanyRoles
			}

			if flagStatus != "" {
				if err := cmdutil.ValidateEnum(
					"status",
					flagStatus,
					[]string{"ACTIVE", "IN_DEVELOPMENT", "DECOMMISSIONED"},
				); err != nil {
					return err
				}

				input["status"] = flagStatus
			}

			if cmd.Flags().Changed("owner") {
				input["ownerId"] = nilString(flagOwner)
			}

			if cmd.Flags().Changed("source") {
				input["source"] = nilString(flagSource)
			}

			if cmd.Flags().Changed("purpose") {
				input["purpose"] = nilString(flagPurpose)
			}

			if cmd.Flags().Changed("intended-use-cases") {
				input["intendedUseCases"] = nilString(flagIntendedUseCases)
			}

			if cmd.Flags().Changed("autonomy-level") {
				input["autonomyLevel"] = nilString(flagAutonomyLevel)
			}

			if cmd.Flags().Changed("human-oversight-mechanism") {
				input["humanOversightMechanism"] = nilString(flagHumanOversightMechanism)
			}

			if cmd.Flags().Changed("risk-classification") {
				if flagRiskClassification != "" {
					if err := cmdutil.ValidateEnum(
						"risk-classification",
						flagRiskClassification,
						[]string{"HIGH_RISK", "LIMITED", "MINIMAL", "GPAI"},
					); err != nil {
						return err
					}
				}

				input["riskClassification"] = nilString(flagRiskClassification)
			}

			if cmd.Flags().Changed("key-stakeholders") {
				input["keyStakeholders"] = nilString(flagKeyStakeholders)
			}

			if cmd.Flags().Changed("data-sources-and-type") {
				input["dataSourcesAndType"] = nilString(flagDataSourcesAndType)
			}

			if cmd.Flags().Changed("deployment-date") {
				input["deploymentDate"] = nilString(flagDeploymentDate)
			}

			if cmd.Flags().Changed("last-review-date") {
				input["lastReviewDate"] = nilString(flagLastReviewDate)
			}

			if cmd.Flags().Changed("next-review-date") {
				input["nextReviewDate"] = nilString(flagNextReviewDate)
			}

			if cmd.Flags().Changed("notes") {
				input["notes"] = nilString(flagNotes)
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

			system := resp.UpdateAiSystem.AiSystem
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated AI system %s (%s)\n",
				system.Name,
				system.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "System name")
	cmd.Flags().StringVar(&flagVersion, "version", "", "System version")
	cmd.Flags().StringArrayVar(&flagCompanyRoles, "company-role", nil, "Company role (repeatable)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Status")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagSource, "source", "", "Source")
	cmd.Flags().StringVar(&flagPurpose, "purpose", "", "Purpose")
	cmd.Flags().StringVar(&flagIntendedUseCases, "intended-use-cases", "", "Intended use cases")
	cmd.Flags().StringVar(&flagAutonomyLevel, "autonomy-level", "", "Autonomy level")
	cmd.Flags().StringVar(&flagHumanOversightMechanism, "human-oversight-mechanism", "", "Human oversight mechanism")
	cmd.Flags().StringVar(&flagRiskClassification, "risk-classification", "", "Risk classification")
	cmd.Flags().StringVar(&flagKeyStakeholders, "key-stakeholders", "", "Key stakeholders")
	cmd.Flags().StringVar(&flagDataSourcesAndType, "data-sources-and-type", "", "Data sources and type")
	cmd.Flags().StringVar(&flagDeploymentDate, "deployment-date", "", "Deployment date")
	cmd.Flags().StringVar(&flagLastReviewDate, "last-review-date", "", "Last review date")
	cmd.Flags().StringVar(&flagNextReviewDate, "next-review-date", "", "Next review date")
	cmd.Flags().StringVar(&flagNotes, "notes", "", "Notes")

	return cmd
}

func nilString(value string) any {
	if value == "" {
		return nil
	}

	return value
}
