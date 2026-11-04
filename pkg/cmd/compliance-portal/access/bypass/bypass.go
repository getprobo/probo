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

package bypass

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const bypassMutation = `
mutation($input: SetCompliancePortalAccessNDABypassInput!) {
  setCompliancePortalAccessNDABypass(input: $input) {
    compliancePortalAccess {
      id
      ndaBypassed
    }
  }
}
`

func NewCmdBypass(f *cmdutil.Factory) *cobra.Command {
	var bypassed bool

	cmd := &cobra.Command{
		Use:   "bypass <id>",
		Short: "Exempt a visitor from the portal NDA",
		Example: `  # Let this visitor open protected documents without signing
  prb compliance-portal access bypass <access-id>

  # Require the portal NDA again
  prb compliance-portal access bypass <access-id> --bypassed=false`,
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

			input := map[string]any{
				"id":       args[0],
				"bypassed": bypassed,
			}

			data, err := client.Do(bypassMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp struct {
				SetCompliancePortalAccessNDABypass struct {
					CompliancePortalAccess struct {
						ID          string `json:"id"`
						NDABypassed bool   `json:"ndaBypassed"`
					} `json:"compliancePortalAccess"`
				} `json:"setCompliancePortalAccessNDABypass"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			access := resp.SetCompliancePortalAccessNDABypass.CompliancePortalAccess
			if access.NDABypassed {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "Bypassed NDA for visitor access %s\n", access.ID)
				return nil
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "NDA required again for visitor access %s\n", access.ID)

			return nil
		},
	}

	cmd.Flags().BoolVar(&bypassed, "bypassed", true, "Exempt the visitor from signing the portal NDA")

	return cmd
}
