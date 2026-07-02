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
mutation($input: UpdateApplicabilityStatementInput!) {
  updateApplicabilityStatement(input: $input) {
    applicabilityStatement {
      id
      applicability
      justification
      control {
        id
        sectionTitle
        name
      }
    }
  }
}
`

type updateResponse struct {
	UpdateApplicabilityStatement struct {
		ApplicabilityStatement struct {
			ID            string `json:"id"`
			Applicability bool   `json:"applicability"`
			Justification string `json:"justification"`
			Control       struct {
				ID           string `json:"id"`
				SectionTitle string `json:"sectionTitle"`
				Name         string `json:"name"`
			} `json:"control"`
		} `json:"applicabilityStatement"`
	} `json:"updateApplicabilityStatement"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagApplicable    bool
		flagNotApplicable bool
		flagJustification string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an applicability statement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagApplicable && flagNotApplicable {
				return fmt.Errorf("cannot set both --applicable and --not-applicable")
			}

			if !flagApplicable && !flagNotApplicable && !cmd.Flags().Changed("justification") {
				return fmt.Errorf("at least one of --applicable, --not-applicable, or --justification is required")
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"applicabilityStatementId": args[0],
			}

			if cmd.Flags().Changed("applicable") || cmd.Flags().Changed("not-applicable") {
				input["applicability"] = flagApplicable
			}

			if cmd.Flags().Changed("justification") {
				input["justification"] = flagJustification
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

			s := resp.UpdateApplicabilityStatement.ApplicabilityStatement

			applicable := "not applicable"
			if s.Applicability {
				applicable = "applicable"
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated statement %s: control %s (%s) marked as %s\n",
				s.ID,
				s.Control.SectionTitle,
				s.Control.Name,
				applicable,
			)

			return nil
		},
	}

	cmd.Flags().BoolVar(&flagApplicable, "applicable", false, "Mark control as applicable")
	cmd.Flags().BoolVar(&flagNotApplicable, "not-applicable", false, "Mark control as not applicable")
	cmd.Flags().StringVar(&flagJustification, "justification", "", "Justification for the applicability decision")

	return cmd
}
