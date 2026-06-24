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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateSCIMConfigurationInput!) {
  createSCIMConfiguration(input: $input) {
    scimConfiguration {
      id
      endpointUrl
    }
    scimBridge {
      id
      state
      type
    }
    token
  }
}
`

type createResponse struct {
	CreateSCIMConfiguration struct {
		ScimConfiguration struct {
			ID          string `json:"id"`
			EndpointURL string `json:"endpointUrl"`
		} `json:"scimConfiguration"`
		ScimBridge *struct {
			ID    string `json:"id"`
			State string `json:"state"`
			Type  string `json:"type"`
		} `json:"scimBridge"`
		Token string `json:"token"`
	} `json:"createSCIMConfiguration"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagConnectorID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a SCIM configuration",
		Example: `  # Create a SCIM configuration
  prb scim create

  # Create with a connector to also set up a SCIM bridge
  prb scim create --connector-id <connector-id>`,
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

			input := map[string]any{
				"organizationId": flagOrg,
			}

			if flagConnectorID != "" {
				input["connectorId"] = flagConnectorID
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

			out := f.IOStreams.Out
			sc := resp.CreateSCIMConfiguration

			_, _ = fmt.Fprintf(out, "Created SCIM configuration %s\n", sc.ScimConfiguration.ID)
			_, _ = fmt.Fprintf(out, "Endpoint URL: %s\n", sc.ScimConfiguration.EndpointURL)

			if sc.ScimBridge != nil {
				_, _ = fmt.Fprintf(out, "Bridge: %s (%s, %s)\n", sc.ScimBridge.ID, sc.ScimBridge.Type, sc.ScimBridge.State)
			}

			_, _ = fmt.Fprintf(out, "\nSCIM Bearer Token (save this — it will not be shown again):\n%s\n", sc.Token)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagConnectorID, "connector-id", "", "Connector ID to create a SCIM bridge")

	return cmd
}
