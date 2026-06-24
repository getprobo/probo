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
      notImplementedJustification
      maturityLevel
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
			NotImplementedJustification *string `json:"notImplementedJustification"`
			MaturityLevel               string  `json:"maturityLevel"`
		} `json:"control"`
	} `json:"updateControl"`
}

var maturityLevelValues = []string{
	"NONE",
	"INITIAL",
	"MANAGED",
	"DEFINED",
	"QUANTITATIVELY_MANAGED",
	"OPTIMIZING",
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagSectionTitle                string
		flagName                        string
		flagDescription                 string
		flagBestPractice                bool
		flagMaturityLevel               string
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

			if cmd.Flags().Changed("maturity-level") {
				if err := cmdutil.ValidateEnum("maturity-level", flagMaturityLevel, maturityLevelValues); err != nil {
					return err
				}

				input["maturityLevel"] = flagMaturityLevel
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
	cmd.Flags().StringVar(&flagMaturityLevel, "maturity-level", "", "CMMI maturity level (NONE, INITIAL, MANAGED, DEFINED, QUANTITATIVELY_MANAGED, OPTIMIZING)")
	cmd.Flags().StringVar(&flagNotImplementedJustification, "not-implemented-justification", "", "Justification when maturity level is NONE")

	return cmd
}
