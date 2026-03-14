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
