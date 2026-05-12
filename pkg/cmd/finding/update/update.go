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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateFindingInput!) {
  updateFinding(input: $input) {
    finding {
      id
      referenceId
      kind
      status
      priority
    }
  }
}
`

type updateResponse struct {
	UpdateFinding struct {
		Finding struct {
			ID          string `json:"id"`
			ReferenceID string `json:"referenceId"`
			Kind        string `json:"kind"`
			Status      string `json:"status"`
			Priority    string `json:"priority"`
		} `json:"finding"`
	} `json:"updateFinding"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDescription      string
		flagSource           string
		flagIdentifiedOn     string
		flagRootCause        string
		flagCorrectiveAction string
		flagOwnerID          string
		flagDueDate          string
		flagStatus           string
		flagPriority         string
		flagRiskID           string
		flagEffectivenessChk string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a finding",
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

			if cmd.Flags().Changed("description") {
				if flagDescription == "" {
					input["description"] = nil
				} else {
					input["description"] = flagDescription
				}
			}
			if cmd.Flags().Changed("source") {
				if flagSource == "" {
					input["source"] = nil
				} else {
					input["source"] = flagSource
				}
			}
			if cmd.Flags().Changed("identified-on") {
				if flagIdentifiedOn == "" {
					input["identifiedOn"] = nil
				} else {
					input["identifiedOn"] = flagIdentifiedOn
				}
			}
			if cmd.Flags().Changed("root-cause") {
				if flagRootCause == "" {
					input["rootCause"] = nil
				} else {
					input["rootCause"] = flagRootCause
				}
			}
			if cmd.Flags().Changed("corrective-action") {
				if flagCorrectiveAction == "" {
					input["correctiveAction"] = nil
				} else {
					input["correctiveAction"] = flagCorrectiveAction
				}
			}
			if cmd.Flags().Changed("owner-id") {
				input["ownerId"] = flagOwnerID
			}
			if cmd.Flags().Changed("due-date") {
				if flagDueDate == "" {
					input["dueDate"] = nil
				} else {
					input["dueDate"] = flagDueDate
				}
			}
			if cmd.Flags().Changed("status") {
				if err := cmdutil.ValidateEnum("status", flagStatus, []string{"OPEN", "IN_PROGRESS", "CLOSED", "RISK_ACCEPTED", "MITIGATED", "FALSE_POSITIVE"}); err != nil {
					return err
				}
				input["status"] = flagStatus
			}
			if cmd.Flags().Changed("priority") {
				if err := cmdutil.ValidateEnum("priority", flagPriority, []string{"LOW", "MEDIUM", "HIGH"}); err != nil {
					return err
				}
				input["priority"] = flagPriority
			}
			if cmd.Flags().Changed("risk-id") {
				if flagRiskID == "" {
					input["riskId"] = nil
				} else {
					input["riskId"] = flagRiskID
				}
			}
			if cmd.Flags().Changed("effectiveness-check") {
				if flagEffectivenessChk == "" {
					input["effectivenessCheck"] = nil
				} else {
					input["effectivenessCheck"] = flagEffectivenessChk
				}
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

			fi := resp.UpdateFinding.Finding
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated finding %s (%s)\n",
				fi.ID,
				fi.ReferenceID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDescription, "description", "", "Finding description")
	cmd.Flags().StringVar(&flagSource, "source", "", "Finding source")
	cmd.Flags().StringVar(&flagIdentifiedOn, "identified-on", "", "Date identified (RFC3339)")
	cmd.Flags().StringVar(&flagRootCause, "root-cause", "", "Root cause")
	cmd.Flags().StringVar(&flagCorrectiveAction, "corrective-action", "", "Corrective action")
	cmd.Flags().StringVar(&flagOwnerID, "owner-id", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date (RFC3339)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Status: OPEN, IN_PROGRESS, CLOSED, RISK_ACCEPTED, MITIGATED, FALSE_POSITIVE")
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Priority: LOW, MEDIUM, HIGH")
	cmd.Flags().StringVar(&flagRiskID, "risk-id", "", "Associated risk ID")
	cmd.Flags().StringVar(&flagEffectivenessChk, "effectiveness-check", "", "Effectiveness check")

	return cmd
}
