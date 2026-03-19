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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateFindingInput!) {
  createFinding(input: $input) {
    findingEdge {
      node {
        id
        referenceId
        kind
        status
        priority
      }
    }
  }
}
`

type createResponse struct {
	CreateFinding struct {
		FindingEdge struct {
			Node struct {
				ID          string `json:"id"`
				ReferenceID string `json:"referenceId"`
				Kind        string `json:"kind"`
				Status      string `json:"status"`
				Priority    string `json:"priority"`
			} `json:"node"`
		} `json:"findingEdge"`
	} `json:"createFinding"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrganization     string
		flagKind             string
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
		Use:   "create",
		Short: "Create a new finding",
		Example: `  # Create a finding
  prb finding create --organization ORG_ID --kind MINOR_NONCONFORMITY --owner-id OWNER_ID --status OPEN --priority HIGH`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("kind", flagKind, []string{"MINOR_NONCONFORMITY", "MAJOR_NONCONFORMITY", "OBSERVATION", "EXCEPTION"}); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("status", flagStatus, []string{"OPEN", "IN_PROGRESS", "CLOSED", "RISK_ACCEPTED", "MITIGATED", "FALSE_POSITIVE"}); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("priority", flagPriority, []string{"LOW", "MEDIUM", "HIGH"}); err != nil {
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

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

			input := map[string]any{
				"organizationId": flagOrganization,
				"kind":           flagKind,
				"status":         flagStatus,
				"priority":       flagPriority,
			}

			if flagOwnerID != "" {
				input["ownerId"] = flagOwnerID
			}
			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if flagSource != "" {
				input["source"] = flagSource
			}
			if flagIdentifiedOn != "" {
				input["identifiedOn"] = flagIdentifiedOn
			}
			if flagRootCause != "" {
				input["rootCause"] = flagRootCause
			}
			if flagCorrectiveAction != "" {
				input["correctiveAction"] = flagCorrectiveAction
			}
			if flagDueDate != "" {
				input["dueDate"] = flagDueDate
			}
			if flagRiskID != "" {
				input["riskId"] = flagRiskID
			}
			if flagEffectivenessChk != "" {
				input["effectivenessCheck"] = flagEffectivenessChk
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

			n := resp.CreateFinding.FindingEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created finding %s (%s)\n",
				n.ID,
				n.ReferenceID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrganization, "organization", "", "Organization ID (required)")
	cmd.Flags().StringVar(&flagKind, "kind", "", "Finding kind: MINOR_NONCONFORMITY, MAJOR_NONCONFORMITY, OBSERVATION, EXCEPTION (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Finding description")
	cmd.Flags().StringVar(&flagSource, "source", "", "Finding source")
	cmd.Flags().StringVar(&flagIdentifiedOn, "identified-on", "", "Date identified (RFC3339)")
	cmd.Flags().StringVar(&flagRootCause, "root-cause", "", "Root cause")
	cmd.Flags().StringVar(&flagCorrectiveAction, "corrective-action", "", "Corrective action")
	cmd.Flags().StringVar(&flagOwnerID, "owner-id", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date (RFC3339)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Status: OPEN, IN_PROGRESS, CLOSED, RISK_ACCEPTED, MITIGATED, FALSE_POSITIVE (required)")
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Priority: LOW, MEDIUM, HIGH (required)")
	cmd.Flags().StringVar(&flagRiskID, "risk-id", "", "Associated risk ID")
	cmd.Flags().StringVar(&flagEffectivenessChk, "effectiveness-check", "", "Effectiveness check")

	_ = cmd.MarkFlagRequired("organization")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("priority")

	return cmd
}
