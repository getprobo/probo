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

package probe

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const (
	sourceQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on AccessReviewSource {
      id
      connectorId
    }
  }
}
`

	probeMutation = `
mutation($input: ProbeConnectorInput!) {
  probeConnector(input: $input) {
    ok
  }
}
`
)

type (
	sourceResponse struct {
		Node *struct {
			Typename    string  `json:"__typename"`
			ID          string  `json:"id"`
			ConnectorID *string `json:"connectorId"`
		} `json:"node"`
	}

	probeResponse struct {
		ProbeConnector struct {
			Ok bool `json:"ok"`
		} `json:"probeConnector"`
	}
)

func NewCmdProbe(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "probe <id>",
		Short: "Probe an access source connector",
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

			sourceID := args[0]

			data, err := client.Do(sourceQuery, map[string]any{"id": sourceID})
			if err != nil {
				return err
			}

			var sourceResp sourceResponse
			if err := json.Unmarshal(data, &sourceResp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if sourceResp.Node == nil {
				return fmt.Errorf("access source %s not found", sourceID)
			}

			if sourceResp.Node.Typename != "AccessReviewSource" {
				return fmt.Errorf("expected AccessReviewSource node, got %s", sourceResp.Node.Typename)
			}

			if sourceResp.Node.ConnectorID == nil || *sourceResp.Node.ConnectorID == "" {
				return fmt.Errorf("access source %s has no connector", sourceID)
			}

			data, err = client.Do(
				probeMutation,
				map[string]any{"input": map[string]any{"connectorId": *sourceResp.Node.ConnectorID}},
			)
			if err != nil {
				return err
			}

			var resp probeResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse probe response: %w", err)
			}

			return cmdutil.PrintJSON(f.IOStreams.Out, resp.ProbeConnector)
		},
	}

	return cmd
}
