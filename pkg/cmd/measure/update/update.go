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
mutation($input: UpdateMeasureInput!) {
  updateMeasure(input: $input) {
    measure {
      id
      name
      category
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
		flagName        string
		flagDescription string
		flagCategory    string
		flagState       string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a measure",
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

	return cmd
}
