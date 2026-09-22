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

package reactivate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const reactivateMutation = `
mutation($input: ReactivateSCIMBridgeInput!) {
  reactivateSCIMBridge(input: $input) {
    scimBridge {
      id
      state
    }
  }
}
`

type reactivateResponse struct {
	ReactivateSCIMBridge struct {
		ScimBridge struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"scimBridge"`
	} `json:"reactivateSCIMBridge"`
}

func NewCmdReactivate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactivate <id>",
		Short: "Reactivate a SCIM bridge so it syncs again",
		Long: `Schedule the next sync for now, clear consecutive errors, and set a
failed, disabled, or stuck bridge back to active.`,
		Example: `  prb scim bridge reactivate <id>`,
		Args:    cobra.ExactArgs(1),
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			data, err := client.Do(
				reactivateMutation,
				map[string]any{
					"input": map[string]any{
						"scimBridgeId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp reactivateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			bridge := resp.ReactivateSCIMBridge.ScimBridge

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Reactivated SCIM bridge %s (state %s)\n",
				bridge.ID,
				bridge.State,
			)

			return nil
		},
	}

	return cmd
}
