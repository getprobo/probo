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
mutation($input: CreateObligationInput!) {
  createObligation(input: $input) {
    obligationEdge {
      node {
        id
        area
        source
        status
        type
      }
    }
  }
}
`

type createResponse struct {
	CreateObligation struct {
		ObligationEdge struct {
			Node struct {
				ID     string `json:"id"`
				Area   string `json:"area"`
				Source string `json:"source"`
				Status string `json:"status"`
				Type   string `json:"type"`
			} `json:"node"`
		} `json:"obligationEdge"`
	} `json:"createObligation"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                    string
		flagArea                   string
		flagSource                 string
		flagStatus                 string
		flagType                   string
		flagRequirement            string
		flagActionsToBeImplemented string
		flagRegulator              string
		flagOwner                  string
		flagLastReviewDate         string
		flagDueDate                string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new obligation",
		Example: `  # Create an obligation interactively
  prb obligation create

  # Create an obligation non-interactively
  prb obligation create --area "Data Protection" --source "GDPR Article 5" --status NON_COMPLIANT --type LEGAL`,
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
				if flagArea == "" {
					err := huh.NewInput().
						Title("Obligation area").
						Value(&flagArea).
						Run()
					if err != nil {
						return err
					}
				}

				if flagSource == "" {
					err := huh.NewInput().
						Title("Obligation source").
						Value(&flagSource).
						Run()
					if err != nil {
						return err
					}
				}

				if flagStatus == "" {
					err := huh.NewSelect[string]().
						Title("Obligation status").
						Options(
							huh.NewOption("Non-Compliant", "NON_COMPLIANT"),
							huh.NewOption("Partially Compliant", "PARTIALLY_COMPLIANT"),
							huh.NewOption("Compliant", "COMPLIANT"),
						).
						Value(&flagStatus).
						Run()
					if err != nil {
						return err
					}
				}

				if flagType == "" {
					err := huh.NewSelect[string]().
						Title("Obligation type").
						Options(
							huh.NewOption("Legal", "LEGAL"),
							huh.NewOption("Contractual", "CONTRACTUAL"),
						).
						Value(&flagType).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagArea == "" {
				return fmt.Errorf("area is required; pass --area or run interactively")
			}
			if flagSource == "" {
				return fmt.Errorf("source is required; pass --source or run interactively")
			}
			if flagStatus == "" {
				return fmt.Errorf("status is required; pass --status or run interactively")
			}
			if flagType == "" {
				return fmt.Errorf("type is required; pass --type or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"area":           flagArea,
				"source":         flagSource,
				"status":         flagStatus,
				"type":           flagType,
			}

			if flagRequirement != "" {
				input["requirement"] = flagRequirement
			}
			if flagActionsToBeImplemented != "" {
				input["actionsToBeImplemented"] = flagActionsToBeImplemented
			}
			if flagRegulator != "" {
				input["regulator"] = flagRegulator
			}
			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}
			if flagLastReviewDate != "" {
				input["lastReviewDate"] = flagLastReviewDate
			}
			if flagDueDate != "" {
				input["dueDate"] = flagDueDate
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

			o := resp.CreateObligation.ObligationEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created obligation %s\n",
				o.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagArea, "area", "", "Obligation area (required)")
	cmd.Flags().StringVar(&flagSource, "source", "", "Obligation source (required)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Obligation status: NON_COMPLIANT, PARTIALLY_COMPLIANT, COMPLIANT (required)")
	cmd.Flags().StringVar(&flagType, "type", "", "Obligation type: LEGAL, CONTRACTUAL (required)")
	cmd.Flags().StringVar(&flagRequirement, "requirement", "", "Obligation requirement")
	cmd.Flags().StringVar(&flagActionsToBeImplemented, "actions-to-be-implemented", "", "Actions to be implemented")
	cmd.Flags().StringVar(&flagRegulator, "regulator", "", "Regulator")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagLastReviewDate, "last-review-date", "", "Last review date (ISO 8601)")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date (ISO 8601)")

	return cmd
}
