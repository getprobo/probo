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
mutation($input: UpdateObligationInput!) {
  updateObligation(input: $input) {
    obligation {
      id
      area
      source
      status
      type
    }
  }
}
`

type updateResponse struct {
	UpdateObligation struct {
		Obligation struct {
			ID     string `json:"id"`
			Area   string `json:"area"`
			Source string `json:"source"`
			Status string `json:"status"`
			Type   string `json:"type"`
		} `json:"obligation"`
	} `json:"updateObligation"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
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
		Use:   "update <id>",
		Short: "Update an obligation",
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

			if cmd.Flags().Changed("area") {
				input["area"] = flagArea
			}
			if cmd.Flags().Changed("source") {
				input["source"] = flagSource
			}
			if cmd.Flags().Changed("status") {
				input["status"] = flagStatus
			}
			if cmd.Flags().Changed("type") {
				input["type"] = flagType
			}
			if cmd.Flags().Changed("requirement") {
				input["requirement"] = flagRequirement
			}
			if cmd.Flags().Changed("actions-to-be-implemented") {
				input["actionsToBeImplemented"] = flagActionsToBeImplemented
			}
			if cmd.Flags().Changed("regulator") {
				input["regulator"] = flagRegulator
			}
			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
			}
			if cmd.Flags().Changed("last-review-date") {
				input["lastReviewDate"] = flagLastReviewDate
			}
			if cmd.Flags().Changed("due-date") {
				input["dueDate"] = flagDueDate
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

			o := resp.UpdateObligation.Obligation
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated obligation %s\n",
				o.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagArea, "area", "", "Obligation area")
	cmd.Flags().StringVar(&flagSource, "source", "", "Obligation source")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Obligation status: NON_COMPLIANT, PARTIALLY_COMPLIANT, COMPLIANT")
	cmd.Flags().StringVar(&flagType, "type", "", "Obligation type: LEGAL, CONTRACTUAL")
	cmd.Flags().StringVar(&flagRequirement, "requirement", "", "Obligation requirement")
	cmd.Flags().StringVar(&flagActionsToBeImplemented, "actions-to-be-implemented", "", "Actions to be implemented")
	cmd.Flags().StringVar(&flagRegulator, "regulator", "", "Regulator")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagLastReviewDate, "last-review-date", "", "Last review date (ISO 8601)")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date (ISO 8601)")

	return cmd
}
