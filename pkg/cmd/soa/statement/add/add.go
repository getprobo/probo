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

package add

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateApplicabilityStatementInput!) {
  createApplicabilityStatement(input: $input) {
    applicabilityStatementEdge {
      node {
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
}
`

type createResponse struct {
	CreateApplicabilityStatement struct {
		ApplicabilityStatementEdge struct {
			Node struct {
				ID            string `json:"id"`
				Applicability bool   `json:"applicability"`
				Justification string `json:"justification"`
				Control       struct {
					ID           string `json:"id"`
					SectionTitle string `json:"sectionTitle"`
					Name         string `json:"name"`
				} `json:"control"`
			} `json:"node"`
		} `json:"applicabilityStatementEdge"`
	} `json:"createApplicabilityStatement"`
}

func NewCmdAdd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagSoA           string
		flagControl       string
		flagApplicable    bool
		flagNotApplicable bool
		flagJustification string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an applicability statement to a SoA",
		Example: `  # Add a control as applicable
  prb soa statement add --soa SOA_ID --control CTRL_ID --applicable

  # Add a control as not applicable with justification
  prb soa statement add --soa SOA_ID --control CTRL_ID --not-applicable --justification "Not in scope"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagApplicable && flagNotApplicable {
				return fmt.Errorf("cannot set both --applicable and --not-applicable")
			}
			if !flagApplicable && !flagNotApplicable {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("either --applicable or --not-applicable is required")
				}

				var choice string
				err := huh.NewSelect[string]().
					Title("Is this control applicable?").
					Options(
						huh.NewOption("Applicable", "applicable"),
						huh.NewOption("Not applicable", "not-applicable"),
					).
					Value(&choice).
					Run()
				if err != nil {
					return err
				}
				flagApplicable = choice == "applicable"
			}

			if !flagApplicable && flagJustification == "" && f.IOStreams.IsInteractive() {
				err := huh.NewText().
					Title("Justification for not applicable").
					Value(&flagJustification).
					Run()
				if err != nil {
					return err
				}
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
				"statementOfApplicabilityId": flagSoA,
				"controlId":                  flagControl,
				"applicability":              flagApplicable,
			}

			if flagJustification != "" {
				input["justification"] = flagJustification
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

			s := resp.CreateApplicabilityStatement.ApplicabilityStatementEdge.Node
			applicable := "not applicable"
			if s.Applicability {
				applicable = "applicable"
			}
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Added statement %s: control %s (%s) marked as %s\n",
				s.ID,
				s.Control.SectionTitle,
				s.Control.Name,
				applicable,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagSoA, "soa", "", "Statement of applicability ID (required)")
	cmd.Flags().StringVar(&flagControl, "control", "", "Control ID (required)")
	cmd.Flags().BoolVar(&flagApplicable, "applicable", false, "Mark control as applicable")
	cmd.Flags().BoolVar(&flagNotApplicable, "not-applicable", false, "Mark control as not applicable")
	cmd.Flags().StringVar(&flagJustification, "justification", "", "Justification for the applicability decision")

	_ = cmd.MarkFlagRequired("soa")
	_ = cmd.MarkFlagRequired("control")

	return cmd
}
