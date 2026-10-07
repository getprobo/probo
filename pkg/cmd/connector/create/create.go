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
	"go.probo.inc/probo/pkg/cmd/connector/catalog"
)

const (
	protocolAPIKey            = "api-key"
	protocolClientCredentials = "client-credentials"

	createAPIKeyMutation = `
mutation($input: CreateAPIKeyConnectorInput!) {
  createAPIKeyConnector(input: $input) {
    connector {
      id
      name
      provider
      protocol
      connectionStatus
    }
  }
}
`

	createClientCredentialsMutation = `
mutation($input: CreateClientCredentialsConnectorInput!) {
  createClientCredentialsConnector(input: $input) {
    connector {
      id
      name
      provider
      protocol
      connectionStatus
    }
  }
}
`
)

type createdConnector struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Provider         string `json:"provider"`
	Protocol         string `json:"protocol"`
	ConnectionStatus string `json:"connectionStatus"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagProtocol     string
		flagProvider     string
		flagName         string
		flagAPIKey       string
		flagClientID     string
		flagClientSecret string
		flagTokenURL     string
		flagScope        string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API-key or client-credentials connector",
		Long:  "Creates a connector for a provider that needs no extra setting. A provider that needs a base URL, slug, or region is connected in the console. Use prb connector providers to list the providers this command accepts.",
		Example: `  # Connect a provider with an API key
  prb connector create --protocol api-key --provider NOTION --name production \
    --api-key <key>

  # Connect a provider with client credentials
  prb connector create --protocol client-credentials --provider OVHCLOUD \
    --name production --client-id <id> --client-secret <secret>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if flagName == "" {
				return fmt.Errorf("--name is required")
			}

			if flagProvider == "" {
				return fmt.Errorf("--provider is required")
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			providers, err := catalog.List(client)
			if err != nil {
				return err
			}

			provider, ok := catalog.Find(providers, flagProvider)
			if !ok {
				return fmt.Errorf("unknown connector provider %s", flagProvider)
			}

			var (
				mutation string
				input    map[string]any
				payload  string
			)

			switch flagProtocol {
			case protocolAPIKey:
				if !provider.APIKeySupported || provider.APIKeyManaged || provider.InstallSupported {
					return fmt.Errorf("%s is connected in the console, not with an API key", provider.DisplayName)
				}

				if len(provider.APIKeyExtraSettings) > 0 {
					return fmt.Errorf("%s needs an extra setting; connect it in the console", provider.DisplayName)
				}

				if flagAPIKey == "" {
					return fmt.Errorf("--api-key is required")
				}

				mutation = createAPIKeyMutation
				payload = "createAPIKeyConnector"
				input = map[string]any{
					"organizationId": flagOrg,
					"name":           flagName,
					"provider":       flagProvider,
					"apiKey":         flagAPIKey,
				}
			case protocolClientCredentials:
				if !provider.ClientCredentialsSupported {
					return fmt.Errorf("%s is connected in the console, not with client credentials", provider.DisplayName)
				}

				if len(provider.ClientCredentialsExtraSettings) > 0 {
					return fmt.Errorf("%s needs an extra setting; connect it in the console", provider.DisplayName)
				}

				if flagClientID == "" || flagClientSecret == "" {
					return fmt.Errorf("--client-id and --client-secret are required")
				}

				mutation = createClientCredentialsMutation
				payload = "createClientCredentialsConnector"

				input = map[string]any{
					"organizationId": flagOrg,
					"name":           flagName,
					"provider":       flagProvider,
					"clientId":       flagClientID,
					"clientSecret":   flagClientSecret,
				}
				if flagTokenURL != "" {
					input["tokenUrl"] = flagTokenURL
				}

				if flagScope != "" {
					input["scope"] = flagScope
				}
			default:
				return fmt.Errorf("--protocol must be %s or %s", protocolAPIKey, protocolClientCredentials)
			}

			data, err := client.Do(mutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var raw map[string]struct {
				Connector createdConnector `json:"connector"`
			}
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			created, ok := raw[payload]
			if !ok || created.Connector.ID == "" {
				return fmt.Errorf("cannot parse response: connector was not created")
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, created.Connector)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created connector %s (%s, %s)\n",
				created.Connector.ID,
				created.Connector.Provider,
				created.Connector.ConnectionStatus,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagProtocol, "protocol", "", "api-key or client-credentials")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Connector provider")
	cmd.Flags().StringVar(&flagName, "name", "", "Name that distinguishes this credential")
	cmd.Flags().StringVar(&flagAPIKey, "api-key", "", "Customer API key")
	cmd.Flags().StringVar(&flagClientID, "client-id", "", "OAuth client ID")
	cmd.Flags().StringVar(&flagClientSecret, "client-secret", "", "OAuth client secret")
	cmd.Flags().StringVar(&flagTokenURL, "token-url", "", "Token endpoint. Leave empty when the provider pins one.")
	cmd.Flags().StringVar(&flagScope, "scope", "", "OAuth scope. The provider's own scopes win when it declares them.")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("protocol")
	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
