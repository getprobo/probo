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
mutation($input: CreateControlInput!) {
  createControl(input: $input) {
    controlEdge {
      node {
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
}
`

type createResponse struct {
	CreateControl struct {
		ControlEdge struct {
			Node struct {
				ID                          string  `json:"id"`
				SectionTitle                string  `json:"sectionTitle"`
				Name                        string  `json:"name"`
				Description                 *string `json:"description"`
				BestPractice                bool    `json:"bestPractice"`
				Implemented                 string  `json:"implemented"`
				NotImplementedJustification *string `json:"notImplementedJustification"`
			} `json:"node"`
		} `json:"controlEdge"`
	} `json:"createControl"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFramework                   string
		flagSectionTitle                string
		flagName                        string
		flagDescription                 string
		flagBestPractice                bool
		flagNotImplemented              bool
		flagNotImplementedJustification string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new control",
		Example: `  # Create a control
  prb control create --framework FW_ID --section-title "A.5" --name "Information security policies"`,
		Args: cobra.NoArgs,
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

			implemented := "IMPLEMENTED"
			if flagNotImplemented {
				implemented = "NOT_IMPLEMENTED"
			}

			input := map[string]any{
				"frameworkId":  flagFramework,
				"sectionTitle": flagSectionTitle,
				"name":         flagName,
				"bestPractice": flagBestPractice,
				"implemented":  implemented,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagNotImplemented && flagNotImplementedJustification != "" {
				input["notImplementedJustification"] = flagNotImplementedJustification
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

			c := resp.CreateControl.ControlEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created control %s (%s)\n",
				c.ID,
				c.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID (required)")
	cmd.Flags().StringVar(&flagSectionTitle, "section-title", "", "Section title (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Control name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Control description")
	cmd.Flags().BoolVar(&flagBestPractice, "best-practice", false, "Mark as best practice")
	cmd.Flags().BoolVar(&flagNotImplemented, "not-implemented", false, "Mark as not implemented")
	cmd.Flags().StringVar(&flagNotImplementedJustification, "not-implemented-justification", "", "Justification for non-implementation")

	_ = cmd.MarkFlagRequired("framework")
	_ = cmd.MarkFlagRequired("section-title")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
