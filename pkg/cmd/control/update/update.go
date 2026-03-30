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
mutation($input: UpdateControlInput!) {
  updateControl(input: $input) {
    control {
      id
      sectionTitle
      name
      description
      bestPractice
      implemented
      notImplementedJustification
    }
  }
}
`

type updateResponse struct {
	UpdateControl struct {
		Control struct {
			ID                          string  `json:"id"`
			SectionTitle                string  `json:"sectionTitle"`
			Name                        string  `json:"name"`
			Description                 *string `json:"description"`
			BestPractice                bool    `json:"bestPractice"`
			Implemented                 string  `json:"implemented"`
			NotImplementedJustification *string `json:"notImplementedJustification"`
		} `json:"control"`
	} `json:"updateControl"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagSectionTitle                string
		flagName                        string
		flagDescription                 string
		flagBestPractice                bool
		flagNotImplemented              bool
		flagNotImplementedJustification string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a control",
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

			if cmd.Flags().Changed("section-title") {
				input["sectionTitle"] = flagSectionTitle
			}
			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("description") {
				if flagDescription == "" {
					input["description"] = nil
				} else {
					input["description"] = flagDescription
				}
			}
			if cmd.Flags().Changed("best-practice") {
				input["bestPractice"] = flagBestPractice
			}
			if cmd.Flags().Changed("not-implemented") {
				if flagNotImplemented {
					input["implemented"] = "NOT_IMPLEMENTED"
				} else {
					input["implemented"] = "IMPLEMENTED"
				}
			}
			if cmd.Flags().Changed("not-implemented-justification") {
				if flagNotImplementedJustification == "" {
					input["notImplementedJustification"] = nil
				} else {
					input["notImplementedJustification"] = flagNotImplementedJustification
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

			c := resp.UpdateControl.Control
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated control %s (%s)\n",
				c.ID,
				c.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagSectionTitle, "section-title", "", "Section title")
	cmd.Flags().StringVar(&flagName, "name", "", "Control name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Control description")
	cmd.Flags().BoolVar(&flagBestPractice, "best-practice", false, "Mark as best practice")
	cmd.Flags().BoolVar(&flagNotImplemented, "not-implemented", false, "Mark as not implemented")
	cmd.Flags().StringVar(&flagNotImplementedJustification, "not-implemented-justification", "", "Justification for non-implementation")

	return cmd
}
