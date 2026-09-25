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

package connect

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const (
	connectMutation = `
mutation($input: CreateWorkloadIdentityConnectorInput!) {
  createWorkloadIdentityConnector(input: $input) {
    connector {
      id
      provider
      protocol
      connectionStatus
    }
  }
}
`

	connectOrganizationMutation = `
mutation($input: CreateOrganizationConnectorInput!) {
  createOrganizationConnector(input: $input) {
    connector {
      id
      provider
      protocol
      connectionStatus
    }
    discoveredAccounts {
      externalAccountId
      name
    }
  }
}
`
)

type (
	connectedConnector struct {
		ID               string `json:"id"`
		Provider         string `json:"provider"`
		Protocol         string `json:"protocol"`
		ConnectionStatus string `json:"connectionStatus"`
	}

	discoveredAccount struct {
		ExternalAccountID string `json:"externalAccountId"`
		Name              string `json:"name"`
	}

	connectResult struct {
		Connector          connectedConnector  `json:"connector"`
		DiscoveredAccounts []discoveredAccount `json:"discoveredAccounts,omitempty"`
	}
)

func NewCmdConnect(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                         string
		flagOrganization                bool
		flagProvider                    string
		flagRoleARN                     string
		flagMemberRoleName              string
		flagGCPWorkloadIdentityProvider string
		flagGCPServiceAccountEmail      string
		flagGCPParent                   string
		flagAzureTenantID               string
		flagAzureClientID               string
		flagAzureSubscriptionID         string
		flagAzureEnvironment            string
		flagOutput                      *string
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Create a workload-identity connector",
		Example: `  # Connect an AWS account
  prb connector connect --provider AWS \
    --aws-role-arn arn:aws:iam::123456789012:role/ProboAudit

  # Connect an AWS organization
  prb connector connect --organization --provider AWS \
    --aws-role-arn arn:aws:iam::111111111111:role/ProboAudit

  # Connect a GCP organization
  prb connector connect --organization --provider GCP \
    --gcp-workload-identity-provider projects/123/locations/global/workloadIdentityPools/probo/providers/probo \
    --gcp-service-account-email probo-audit@my-project.iam.gserviceaccount.com \
    --gcp-parent organizations/123456789`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if flagProvider == "" {
				return fmt.Errorf("--provider is required")
			}

			if flagOrganization && flagAzureSubscriptionID != "" {
				return fmt.Errorf("--azure-subscription-id cannot be used with --organization")
			}

			if !flagOrganization && (flagMemberRoleName != "" || flagGCPParent != "") {
				return fmt.Errorf("--organization is required to set --aws-member-role-name or --gcp-parent")
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

			input := connectorInput(
				flagOrg,
				flagProvider,
				flagRoleARN,
				flagGCPWorkloadIdentityProvider,
				flagGCPServiceAccountEmail,
				flagAzureTenantID,
				flagAzureClientID,
				flagAzureEnvironment,
			)

			mutation := connectMutation
			if flagOrganization {
				mutation = connectOrganizationMutation

				setInput(input, "awsMemberRoleName", flagMemberRoleName)
				setInput(input, "gcpParent", flagGCPParent)
			} else {
				setInput(input, "azureSubscriptionId", flagAzureSubscriptionID)
			}

			data, err := client.Do(mutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			result, err := parseConnectResult(data, flagOrganization)
			if err != nil {
				return err
			}

			return printConnectResult(f, flagOutput, flagOrganization, result)
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().BoolVar(&flagOrganization, "organization", false, "Connect an organization and list its accounts")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Connector provider (AWS, GCP, AZURE)")
	cmd.Flags().StringVar(&flagRoleARN, "aws-role-arn", "", "IAM role ARN")
	cmd.Flags().StringVar(&flagMemberRoleName, "aws-member-role-name", "", "IAM role assumed in member accounts")
	cmd.Flags().StringVar(
		&flagGCPWorkloadIdentityProvider,
		"gcp-workload-identity-provider",
		"",
		"GCP workload identity provider resource",
	)
	cmd.Flags().StringVar(
		&flagGCPServiceAccountEmail,
		"gcp-service-account-email",
		"",
		"GCP service account email to impersonate",
	)
	cmd.Flags().StringVar(&flagGCPParent, "gcp-parent", "", "Cloud Asset parent (organizations/{n} or folders/{n})")
	cmd.Flags().StringVar(&flagAzureTenantID, "azure-tenant-id", "", "Entra directory (tenant) ID")
	cmd.Flags().StringVar(&flagAzureClientID, "azure-client-id", "", "Entra application (client) ID")
	cmd.Flags().StringVar(&flagAzureSubscriptionID, "azure-subscription-id", "", "Azure subscription ID")
	cmd.Flags().StringVar(&flagAzureEnvironment, "azure-environment", "", "Azure environment (AZURE_PUBLIC, AZURE_GOVERNMENT, ...)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func connectorInput(
	orgID string,
	provider string,
	roleARN string,
	gcpProvider string,
	gcpServiceAccountEmail string,
	azureTenantID string,
	azureClientID string,
	azureEnvironment string,
) map[string]any {
	input := map[string]any{
		"organizationId": orgID,
		"provider":       provider,
	}

	setInput(input, "awsRoleArn", roleARN)
	setInput(input, "gcpWorkloadIdentityProvider", gcpProvider)
	setInput(input, "gcpServiceAccountEmail", gcpServiceAccountEmail)
	setInput(input, "azureTenantId", azureTenantID)
	setInput(input, "azureClientId", azureClientID)
	setInput(input, "azureEnvironment", azureEnvironment)

	return input
}

func setInput(input map[string]any, key string, value string) {
	if value == "" {
		return
	}

	input[key] = value
}

func parseConnectResult(data []byte, organization bool) (connectResult, error) {
	if organization {
		var resp struct {
			CreateOrganizationConnector connectResult `json:"createOrganizationConnector"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return connectResult{}, fmt.Errorf("cannot parse response: %w", err)
		}

		return resp.CreateOrganizationConnector, nil
	}

	var resp struct {
		CreateWorkloadIdentityConnector connectResult `json:"createWorkloadIdentityConnector"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return connectResult{}, fmt.Errorf("cannot parse response: %w", err)
	}

	return resp.CreateWorkloadIdentityConnector, nil
}

func printConnectResult(
	f *cmdutil.Factory,
	flagOutput *string,
	organization bool,
	result connectResult,
) error {
	if *flagOutput == cmdutil.OutputJSON {
		return cmdutil.PrintJSON(f.IOStreams.Out, result)
	}

	out := f.IOStreams.Out
	_, _ = fmt.Fprintf(out, "Created connector %s\n", result.Connector.ID)
	_, _ = fmt.Fprintf(out, "Provider: %s\n", result.Connector.Provider)

	_, _ = fmt.Fprintf(out, "Protocol: %s\n", result.Connector.Protocol)
	if result.Connector.ConnectionStatus != "" {
		_, _ = fmt.Fprintf(out, "Status: %s\n", result.Connector.ConnectionStatus)
	}

	if !organization {
		return nil
	}

	_, _ = fmt.Fprintf(out, "Discovered %d account(s)\n", len(result.DiscoveredAccounts))
	if len(result.DiscoveredAccounts) == 0 {
		return nil
	}

	rows := make([][]string, 0, len(result.DiscoveredAccounts))
	for _, account := range result.DiscoveredAccounts {
		rows = append(rows, []string{account.ExternalAccountID, account.Name})
	}

	t := cmdutil.NewTable("EXTERNAL ID", "NAME").Rows(rows...)
	_, _ = fmt.Fprintln(out, t)

	return nil
}
