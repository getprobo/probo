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
	"go.probo.inc/probo/pkg/timespan"
)

const updateMutation = `
mutation($input: UpdateMeasureInput!) {
  updateMeasure(input: $input) {
    measure {
      id
      name
      category
      code
      implementationStatus
      state
    }
  }
}
`

type updateResponse struct {
	UpdateMeasure struct {
		Measure struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			State    string `json:"state"`
		} `json:"measure"`
	} `json:"updateMeasure"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName                 string
		flagDescription          string
		flagCategory             string
		flagState                string
		flagCode                 string
		flagControlType          string
		flagNature               string
		flagOperatingMode        string
		flagOperatingFrequency   string
		flagOperatingEvent       string
		flagEvidenceCadence      string
		flagTestingCadence       string
		flagImplementationStatus string
		flagOwnerID              string
		flagReviewerID           string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an internal control",
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

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}

			if cmd.Flags().Changed("category") {
				input["category"] = flagCategory
			}

			if cmd.Flags().Changed("state") {
				if err := cmdutil.ValidateEnum("state", flagState, []string{"NOT_STARTED", "IN_PROGRESS", "NOT_APPLICABLE", "IMPLEMENTED"}); err != nil {
					return err
				}

				input["state"] = flagState
			}

			if err := setMeasureFields(cmd, input, measureFieldFlags{
				code:                 flagCode,
				controlType:          flagControlType,
				nature:               flagNature,
				operatingMode:        flagOperatingMode,
				operatingFrequency:   flagOperatingFrequency,
				operatingEvent:       flagOperatingEvent,
				evidenceCadence:      flagEvidenceCadence,
				testingCadence:       flagTestingCadence,
				implementationStatus: flagImplementationStatus,
				ownerID:              flagOwnerID,
				reviewerID:           flagReviewerID,
			}); err != nil {
				return err
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

			m := resp.UpdateMeasure.Measure
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated measure %s (%s)\n",
				m.ID,
				m.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Measure name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Measure description")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Measure category")
	cmd.Flags().StringVar(&flagState, "state", "", "Measure state: NOT_STARTED, IN_PROGRESS, NOT_APPLICABLE, IMPLEMENTED")
	addMeasureFieldFlags(
		cmd,
		&flagCode,
		&flagControlType,
		&flagNature,
		&flagOperatingMode,
		&flagOperatingFrequency,
		&flagOperatingEvent,
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
	operatingMode        string
	operatingFrequency   string
	operatingEvent       string
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
	operatingMode *string,
	operatingFrequency *string,
	operatingEvent *string,
	evidenceCadence *string,
	testingCadence *string,
	implementationStatus *string,
	ownerID *string,
	reviewerID *string,
) {
	cmd.Flags().StringVar(code, "code", "", "Stable reference, for example IC-ACCESS-01")
	cmd.Flags().StringVar(controlType, "control-type", "", "Control type: PREVENTIVE, DETECTIVE, CORRECTIVE")
	cmd.Flags().StringVar(nature, "nature", "", "Nature: MANUAL")
	cmd.Flags().StringVar(operatingMode, "operating-mode", "", "Operating mode: CONTINUOUS, EVENT, or PERIODIC. Empty clears it")
	cmd.Flags().StringVar(operatingFrequency, "operating-frequency", "", "ISO-8601 duration for PERIODIC mode, for example P3M")
	cmd.Flags().StringVar(operatingEvent, "operating-event", "", "Event that runs the control, for example when someone leaves")
	cmd.Flags().StringVar(evidenceCadence, "evidence-cadence", "", "How often evidence is collected, as an ISO-8601 duration such as P1M")
	cmd.Flags().StringVar(testingCadence, "testing-cadence", "", "How often effectiveness is tested, as an ISO-8601 duration such as P3M")
	cmd.Flags().StringVar(implementationStatus, "implementation-status", "", "Status: NOT_IMPLEMENTED, IN_PROGRESS, IMPLEMENTED, OPERATING")
	cmd.Flags().StringVar(ownerID, "owner-id", "", "Owner profile ID")
	cmd.Flags().StringVar(reviewerID, "reviewer-id", "", "Reviewer profile ID")
}

func setOperatingFrequency(cmd *cobra.Command, input map[string]any, mode, interval, event string) error {
	if !cmd.Flags().Changed("operating-mode") &&
		!cmd.Flags().Changed("operating-frequency") &&
		!cmd.Flags().Changed("operating-event") {
		return nil
	}

	if mode == "" {
		if interval != "" || event != "" {
			return fmt.Errorf("operating-mode is required")
		}

		input["operatingFrequency"] = nil

		return nil
	}

	if err := cmdutil.ValidateEnum("operating-mode", mode, []string{"CONTINUOUS", "EVENT", "PERIODIC"}); err != nil {
		return err
	}

	value := map[string]any{"mode": mode}

	switch mode {
	case "PERIODIC":
		if interval == "" {
			return fmt.Errorf("operating-frequency is required for PERIODIC")
		}

		span, err := timespan.Parse(interval)
		if err != nil {
			return fmt.Errorf("operating-frequency: %w", err)
		}

		value["interval"] = span.String()
	case "EVENT":
		if event != "" {
			value["event"] = event
		}
	}

	input["operatingFrequency"] = value

	return nil
}

func setDuration(cmd *cobra.Command, input map[string]any, flag, key, raw string) error {
	if !cmd.Flags().Changed(flag) {
		return nil
	}

	if raw == "" {
		input[key] = nil
		return nil
	}

	span, err := timespan.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", flag, err)
	}

	input[key] = span.String()

	return nil
}

func setMeasureFields(cmd *cobra.Command, input map[string]any, flags measureFieldFlags) error {
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

	if err := setOperatingFrequency(cmd, input, flags.operatingMode, flags.operatingFrequency, flags.operatingEvent); err != nil {
		return err
	}

	if err := setDuration(cmd, input, "evidence-cadence", "evidenceCadence", flags.evidenceCadence); err != nil {
		return err
	}

	if err := setDuration(cmd, input, "testing-cadence", "testingCadence", flags.testingCadence); err != nil {
		return err
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
