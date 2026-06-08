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
mutation($input: UpdateRiskAssessmentBoundaryInput!) {
  updateRiskAssessmentBoundary(input: $input) {
    riskAssessmentBoundary {
      id
      parentBoundaryId
      name
      createdAt
      updatedAt
    }
  }
}
`

type updateResponse struct {
	UpdateRiskAssessmentBoundary struct {
		RiskAssessmentBoundary struct {
			ID               string  `json:"id"`
			ParentBoundaryId *string `json:"parentBoundaryId"`
			Name             string  `json:"name"`
			CreatedAt        string  `json:"createdAt"`
			UpdatedAt        string  `json:"updatedAt"`
		} `json:"riskAssessmentBoundary"`
	} `json:"updateRiskAssessmentBoundary"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagParentId    string
		flagClearParent bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a risk assessment boundary",
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

			if flagClearParent && cmd.Flags().Changed("parent-id") {
				return fmt.Errorf("cannot use --parent-id and --clear-parent together")
			}

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("parent-id") {
				input["parentBoundaryId"] = flagParentId
			}

			if flagClearParent {
				input["parentBoundaryId"] = nil
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

			r := resp.UpdateRiskAssessmentBoundary.RiskAssessmentBoundary
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated risk assessment boundary %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Boundary name")
	cmd.Flags().StringVar(&flagParentId, "parent-id", "", "Parent boundary ID")
	cmd.Flags().BoolVar(&flagClearParent, "clear-parent", false, "Remove the parent boundary (make it top-level)")

	return cmd
}
