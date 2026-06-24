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

package regeneratetoken

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const regenerateMutation = `
mutation($input: RegenerateSCIMTokenInput!) {
  regenerateSCIMToken(input: $input) {
    scimConfiguration {
      id
    }
    token
  }
}
`

type regenerateResponse struct {
	RegenerateSCIMToken struct {
		ScimConfiguration struct {
			ID string `json:"id"`
		} `json:"scimConfiguration"`
		Token string `json:"token"`
	} `json:"regenerateSCIMToken"`
}

func NewCmdRegenerateToken(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg string
		flagYes bool
	)

	cmd := &cobra.Command{
		Use:   "regenerate-token <scim-configuration-id>",
		Short: "Regenerate the SCIM bearer token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot regenerate SCIM token: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				err := huh.NewConfirm().
					Title("Regenerate SCIM token? The current token will be invalidated.").
					Value(&confirmed).
					Run()
				if err != nil {
					return err
				}

				if !confirmed {
					return nil
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			data, err := client.Do(
				regenerateMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId":      flagOrg,
						"scimConfigurationId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp regenerateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Regenerated SCIM token for configuration %s\n", resp.RegenerateSCIMToken.ScimConfiguration.ID)
			_, _ = fmt.Fprintf(out, "\nSCIM Bearer Token (save this — it will not be shown again):\n%s\n", resp.RegenerateSCIMToken.Token)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
