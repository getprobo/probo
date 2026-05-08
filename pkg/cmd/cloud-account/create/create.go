// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateCloudAccountInput!) {
  createCloudAccount(input: $input) {
    cloudAccount {
      id
      label
      provider
      status
    }
    verifyStatus
    lastProbeError
  }
}
`

type createResponse struct {
	CreateCloudAccount struct {
		CloudAccount struct {
			ID       string `json:"id"`
			Label    string `json:"label"`
			Provider string `json:"provider"`
			Status   string `json:"status"`
		} `json:"cloudAccount"`
		VerifyStatus   string  `json:"verifyStatus"`
		LastProbeError *string `json:"lastProbeError"`
	} `json:"createCloudAccount"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                 string
		flagLabel               string
		flagProvider            string
		flagCredentialKind      string
		flagScopeKind           string
		flagScopeIdentifier     string
		flagModules             []string
		flagAWSRoleARN          string
		flagAWSExternalID       string
		flagGCPProjectID        string
		flagGCPOrganizationID   string
		flagAzureTenantID       string
		flagAzureClientID       string
		flagAzureSubscriptionID string
		flagAzureMgmtGroupID    string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cloud account and synchronously verify it",
		Long: `Create a new cloud account row in PENDING_VERIFICATION status, then
synchronously probe it.

For AWS callers: run 'prb cloud-account install-assets --provider AWS
--scope-kind AWS_ACCOUNT --scope-identifier <account-id> --modules
ACCESS_REVIEW --aws-region <region> --output json' FIRST to obtain the
external_id Probo persisted on the future row. Pass that same value to
--aws-external-id below; do NOT invent your own external_id.

Secret credential bodies (GCP service-account JSON, Azure
client_secret) are NOT accepted as flags here. They must be uploaded
out-of-band via the dedicated
/api/console/v1/cloud-accounts/credentials/upload endpoint.`,
		Example: `  # AWS, post-install-assets
  prb cloud-account create \
    --provider AWS \
    --credential-kind AWS_ASSUME_ROLE \
    --scope-kind AWS_ACCOUNT \
    --scope-identifier 123456789012 \
    --label 'Prod AWS' \
    --modules ACCESS_REVIEW \
    --aws-role-arn arn:aws:iam::123456789012:role/Probo-Auditor \
    --aws-external-id <value-from-install-assets>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("provider", flagProvider, []string{"AWS", "GCP", "AZURE"}); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("credential-kind", flagCredentialKind, []string{"AWS_ASSUME_ROLE", "GCP_SERVICE_ACCOUNT_KEY", "AZURE_CLIENT_SECRET"}); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("scope-kind", flagScopeKind, []string{"AWS_ACCOUNT", "GCP_PROJECT", "GCP_ORGANIZATION", "AZURE_SUBSCRIPTION", "AZURE_MANAGEMENT_GROUP", "AZURE_TENANT"}); err != nil {
				return err
			}
			for _, m := range flagModules {
				if err := cmdutil.ValidateEnum("modules", m, []string{"ACCESS_REVIEW"}); err != nil {
					return err
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
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("cannot determine organization, use --org or 'prb auth login'")
			}

			input := map[string]any{
				"organizationId":      flagOrg,
				"label":               flagLabel,
				"provider":            flagProvider,
				"credentialKind":      flagCredentialKind,
				"scopeKind":           flagScopeKind,
				"scopeIdentifier":     flagScopeIdentifier,
				"enabledAuditModules": flagModules,
			}

			if flagAWSRoleARN != "" {
				input["awsRoleArn"] = flagAWSRoleARN
			}
			if flagAWSExternalID != "" {
				input["awsExternalId"] = flagAWSExternalID
			}
			if flagGCPProjectID != "" {
				input["gcpProjectId"] = flagGCPProjectID
			}
			if flagGCPOrganizationID != "" {
				input["gcpOrganizationId"] = flagGCPOrganizationID
			}
			if flagAzureTenantID != "" {
				input["azureTenantId"] = flagAzureTenantID
			}
			if flagAzureClientID != "" {
				input["azureClientId"] = flagAzureClientID
			}
			if flagAzureSubscriptionID != "" {
				input["azureSubscriptionId"] = flagAzureSubscriptionID
			}
			if flagAzureMgmtGroupID != "" {
				input["azureManagementGroupId"] = flagAzureMgmtGroupID
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

			a := resp.CreateCloudAccount.CloudAccount
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Created cloud account %s\n", a.ID)
			_, _ = fmt.Fprintf(out, "Label: %s\n", a.Label)
			_, _ = fmt.Fprintf(out, "Provider: %s\n", a.Provider)
			_, _ = fmt.Fprintf(out, "Status: %s\n", a.Status)
			_, _ = fmt.Fprintf(out, "Verify Status: %s\n", resp.CreateCloudAccount.VerifyStatus)
			if resp.CreateCloudAccount.LastProbeError != nil {
				_, _ = fmt.Fprintf(out, "Last Probe Error: %s\n", *resp.CreateCloudAccount.LastProbeError)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagLabel, "label", "", "Customer-supplied label (required)")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Provider (AWS, GCP, AZURE) (required)")
	cmd.Flags().StringVar(&flagCredentialKind, "credential-kind", "", "Credential kind (AWS_ASSUME_ROLE, GCP_SERVICE_ACCOUNT_KEY, AZURE_CLIENT_SECRET) (required)")
	cmd.Flags().StringVar(&flagScopeKind, "scope-kind", "", "Scope kind (required)")
	cmd.Flags().StringVar(&flagScopeIdentifier, "scope-identifier", "", "Scope identifier (required)")
	cmd.Flags().StringSliceVar(&flagModules, "modules", nil, "Audit modules to enable (ACCESS_REVIEW) (required)")

	cmd.Flags().StringVar(&flagAWSRoleARN, "aws-role-arn", "", "AWS role ARN to assume (AWS only)")
	cmd.Flags().StringVar(&flagAWSExternalID, "aws-external-id", "", "AWS external_id (AWS only) -- MUST come from 'prb cloud-account install-assets --output json'")

	cmd.Flags().StringVar(&flagGCPProjectID, "gcp-project-id", "", "GCP project id (GCP only)")
	cmd.Flags().StringVar(&flagGCPOrganizationID, "gcp-organization-id", "", "GCP organization id (GCP only)")

	cmd.Flags().StringVar(&flagAzureTenantID, "azure-tenant-id", "", "Azure tenant id (Azure only)")
	cmd.Flags().StringVar(&flagAzureClientID, "azure-client-id", "", "Azure client id (Azure only)")
	cmd.Flags().StringVar(&flagAzureSubscriptionID, "azure-subscription-id", "", "Azure subscription id (Azure only)")
	cmd.Flags().StringVar(&flagAzureMgmtGroupID, "azure-management-group-id", "", "Azure management group id (Azure only)")

	_ = cmd.MarkFlagRequired("label")
	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("credential-kind")
	_ = cmd.MarkFlagRequired("scope-kind")
	_ = cmd.MarkFlagRequired("scope-identifier")
	_ = cmd.MarkFlagRequired("modules")

	return cmd
}
