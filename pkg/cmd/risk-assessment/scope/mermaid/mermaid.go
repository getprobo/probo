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
