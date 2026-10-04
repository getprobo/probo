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
	"go.probo.inc/probo/pkg/cmd/measure/fields"
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

			operatingMode := flagOperatingMode
			operatingEvent := flagOperatingEvent

			if err := preserveOperatingFrequency(cmd, client, args[0], &operatingMode, &operatingEvent); err != nil {
				return err
			}

			if err := fields.SetMeasureFields(cmd, input, fields.MeasureFieldFlags{
				Code:                 flagCode,
				ControlType:          flagControlType,
				Nature:               flagNature,
				OperatingMode:        operatingMode,
				OperatingFrequency:   flagOperatingFrequency,
				OperatingEvent:       operatingEvent,
				EvidenceCadence:      flagEvidenceCadence,
				TestingCadence:       flagTestingCadence,
				ImplementationStatus: flagImplementationStatus,
				OwnerID:              flagOwnerID,
				ReviewerID:           flagReviewerID,
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
	fields.AddMeasureFieldFlags(
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

const operatingFrequencyQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Measure {
      operatingFrequency {
        mode
        event
      }
    }
  }
}
`

type storedOperatingFrequency struct {
	Mode  string
	Event string
}

func preserveOperatingFrequency(
	cmd *cobra.Command,
	client *api.Client,
	id string,
	mode *string,
	event *string,
) error {
	sendingFrequency := cmd.Flags().Changed("operating-mode") ||
		cmd.Flags().Changed("operating-frequency") ||
		cmd.Flags().Changed("operating-event")
	if !sendingFrequency {
		return nil
	}

	needsMode := !cmd.Flags().Changed("operating-mode")
	needsEvent := !cmd.Flags().Changed("operating-event") && (needsMode || *mode == "EVENT")

	if !needsMode && !needsEvent {
		return nil
	}

	current, err := currentOperatingFrequency(client, id)
	if err != nil {
		return err
	}

	if needsMode {
		*mode = current.Mode
	}

	if needsEvent && *mode == "EVENT" {
		*event = current.Event
	}

	return nil
}

func currentOperatingFrequency(client *api.Client, id string) (storedOperatingFrequency, error) {
	data, err := client.Do(operatingFrequencyQuery, map[string]any{"id": id})
	if err != nil {
		return storedOperatingFrequency{}, err
	}

	var resp struct {
		Node *struct {
			Typename           string `json:"__typename"`
			OperatingFrequency *struct {
				Mode  string  `json:"mode"`
				Event *string `json:"event"`
			} `json:"operatingFrequency"`
		} `json:"node"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return storedOperatingFrequency{}, fmt.Errorf("cannot parse response: %w", err)
	}

	if resp.Node == nil || resp.Node.Typename != "Measure" {
		return storedOperatingFrequency{}, fmt.Errorf("measure %s not found", id)
	}

	if resp.Node.OperatingFrequency == nil {
		return storedOperatingFrequency{}, nil
	}

	current := storedOperatingFrequency{Mode: resp.Node.OperatingFrequency.Mode}
	if resp.Node.OperatingFrequency.Event != nil {
		current.Event = *resp.Node.OperatingFrequency.Event
	}

	return current, nil
}
