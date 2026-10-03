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
mutation($input: CreateMeasureInput!) {
  createMeasure(input: $input) {
    measureEdge {
      node {
        id
        name
        category
        code
        implementationStatus
        state
      }
    }
  }
}
`

type createResponse struct {
	CreateMeasure struct {
		MeasureEdge struct {
			Node struct {
				ID                   string  `json:"id"`
				Name                 string  `json:"name"`
				Category             string  `json:"category"`
				Code                 *string `json:"code"`
				ImplementationStatus string  `json:"implementationStatus"`
				State                string  `json:"state"`
			} `json:"node"`
		} `json:"measureEdge"`
	} `json:"createMeasure"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                  string
		flagName                 string
		flagCategory             string
		flagDescription          string
		flagCode                 string
		flagControlType          string
		flagNature               string
		flagOperatingFrequency   string
		flagEvidenceCadence      string
		flagTestingCadence       string
		flagImplementationStatus string
		flagOwnerID              string
		flagReviewerID           string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an internal control",
		Example: `  # Create a measure interactively
  prb measure create

  # Create a measure non-interactively
  prb measure create --name "Enable encryption at rest" --category "Security"`,
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
						Title("Measure name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagCategory == "" {
					err := huh.NewInput().
						Title("Measure category").
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

			if err := setMeasureFields(cmd, input, measureFieldFlags{
				code:                 flagCode,
				controlType:          flagControlType,
				nature:               flagNature,
				operatingFrequency:   flagOperatingFrequency,
				evidenceCadence:      flagEvidenceCadence,
				testingCadence:       flagTestingCadence,
				implementationStatus: flagImplementationStatus,
				ownerID:              flagOwnerID,
				reviewerID:           flagReviewerID,
			}); err != nil {
				return err
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

			m := resp.CreateMeasure.MeasureEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created measure %s (%s)\n",
				m.ID,
				m.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Measure name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Measure category (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Measure description")
	addMeasureFieldFlags(
		cmd,
		&flagCode,
		&flagControlType,
		&flagNature,
		&flagOperatingFrequency,
		&flagEvidenceCadence,
		&flagTestingCadence,
		&flagImplementationStatus,
		&flagOwnerID,
		&flagReviewerID,
	)

	return cmd
}

type measureFieldFlags struct {
	code                 string
	controlType          string
	nature               string
	operatingFrequency   string
	evidenceCadence      string
	testingCadence       string
	implementationStatus string
	ownerID              string
	reviewerID           string
}

func addMeasureFieldFlags(
	cmd *cobra.Command,
	code *string,
	controlType *string,
	nature *string,
	operatingFrequency *string,
	evidenceCadence *string,
	testingCadence *string,
	implementationStatus *string,
	ownerID *string,
	reviewerID *string,
) {
	cmd.Flags().StringVar(code, "code", "", "Stable reference, for example IC-ACCESS-01")
	cmd.Flags().StringVar(controlType, "control-type", "", "Control type: PREVENTIVE, DETECTIVE, CORRECTIVE")
	cmd.Flags().StringVar(nature, "nature", "", "Nature: MANUAL")
	cmd.Flags().StringVar(operatingFrequency, "operating-frequency", "", "How often the control runs")
	cmd.Flags().StringVar(evidenceCadence, "evidence-cadence", "", "How often evidence is collected")
	cmd.Flags().StringVar(testingCadence, "testing-cadence", "", "How often effectiveness is tested")
	cmd.Flags().StringVar(implementationStatus, "implementation-status", "", "Status: NOT_IMPLEMENTED, IN_PROGRESS, IMPLEMENTED, OPERATING")
	cmd.Flags().StringVar(ownerID, "owner-id", "", "Owner profile ID")
	cmd.Flags().StringVar(reviewerID, "reviewer-id", "", "Reviewer profile ID")
}

func setMeasureFields(cmd *cobra.Command, input map[string]any, flags measureFieldFlags) error {
	cadences := []string{
		"CONTINUOUS",
		"DAILY",
		"WEEKLY",
		"MONTHLY",
		"QUARTERLY",
		"SEMIANNUALLY",
		"ANNUALLY",
		"AD_HOC",
	}

	if cmd.Flags().Changed("code") {
		input["code"] = flags.code
	}

	if cmd.Flags().Changed("control-type") {
		if err := cmdutil.ValidateEnum("control-type", flags.controlType, []string{"PREVENTIVE", "DETECTIVE", "CORRECTIVE"}); err != nil {
			return err
		}

		input["controlType"] = flags.controlType
	}

	if cmd.Flags().Changed("nature") {
		if err := cmdutil.ValidateEnum("nature", flags.nature, []string{"MANUAL"}); err != nil {
			return err
		}

		input["nature"] = flags.nature
	}

	if cmd.Flags().Changed("operating-frequency") {
		if err := cmdutil.ValidateEnum("operating-frequency", flags.operatingFrequency, cadences); err != nil {
			return err
		}

		input["operatingFrequency"] = flags.operatingFrequency
	}

	if cmd.Flags().Changed("evidence-cadence") {
		if err := cmdutil.ValidateEnum("evidence-cadence", flags.evidenceCadence, cadences); err != nil {
			return err
		}

		input["evidenceCadence"] = flags.evidenceCadence
	}

	if cmd.Flags().Changed("testing-cadence") {
		if err := cmdutil.ValidateEnum("testing-cadence", flags.testingCadence, cadences); err != nil {
			return err
		}

		input["testingCadence"] = flags.testingCadence
	}

	if cmd.Flags().Changed("implementation-status") {
		if err := cmdutil.ValidateEnum(
			"implementation-status",
			flags.implementationStatus,
			[]string{"NOT_IMPLEMENTED", "IN_PROGRESS", "IMPLEMENTED", "OPERATING"},
		); err != nil {
			return err
		}

		input["implementationStatus"] = flags.implementationStatus
	}

	if cmd.Flags().Changed("owner-id") {
		input["ownerId"] = flags.ownerID
	}

	if cmd.Flags().Changed("reviewer-id") {
		input["reviewerId"] = flags.reviewerID
	}

	return nil
}
