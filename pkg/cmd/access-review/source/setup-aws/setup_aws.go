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

package setupaws

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const setupQuery = `
query($organizationId: ID!) {
  awsConnectorSetup(organizationId: $organizationId) {
    issuer
    audience
    subject
    suggestedRoleName
    terraformSnippet
    cloudFormationQuickCreateURL
  }
}
`

type setupResponse struct {
	AWSConnectorSetup struct {
		Issuer                       string `json:"issuer"`
		Audience                     string `json:"audience"`
		Subject                      string `json:"subject"`
		SuggestedRoleName            string `json:"suggestedRoleName"`
		TerraformSnippet             string `json:"terraformSnippet"`
		CloudFormationQuickCreateURL string `json:"cloudFormationQuickCreateURL"`
	} `json:"awsConnectorSetup"`
}

func NewCmdSetupAWS(f *cmdutil.Factory) *cobra.Command {
	var flagOrg string

	cmd := &cobra.Command{
		Use:   "setup-aws",
		Short: "Show AWS access source setup values",
		Args:  cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			data, err := client.Do(setupQuery, map[string]any{"organizationId": flagOrg})
			if err != nil {
				return err
			}

			var resp setupResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			return cmdutil.PrintJSON(f.IOStreams.Out, resp.AWSConnectorSetup)
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")

	return cmd
}
