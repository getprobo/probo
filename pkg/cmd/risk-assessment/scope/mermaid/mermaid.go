// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package mermaid

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const mermaidQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on RiskAssessmentScope {
      id
      name
      mermaidChart
    }
  }
}
`

type mermaidResponse struct {
	Node *struct {
		Typename     string `json:"__typename"`
		ID           string `json:"id"`
		Name         string `json:"name"`
		MermaidChart string `json:"mermaidChart"`
	} `json:"node"`
}

func NewCmdMermaid(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mermaid <id>",
		Short: "Get the Mermaid chart for a risk assessment scope",
		Example: `  # Print the Mermaid chart for a scope
  prb risk-assessment scope mermaid <id>

  # Output as JSON
  prb risk-assessment scope mermaid <id> --json`,
		Args: cobra.ExactArgs(1),
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

			data, err := client.Do(
				mermaidQuery,
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var resp mermaidResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("risk assessment scope %s not found", args[0])
			}

			if resp.Node.Typename != "RiskAssessmentScope" {
				return fmt.Errorf("expected RiskAssessmentScope node, got %s", resp.Node.Typename)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, resp.Node.MermaidChart)

			return nil
		},
	}

	return cmd
}
