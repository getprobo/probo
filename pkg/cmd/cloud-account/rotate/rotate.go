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

package rotate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const rotateMutation = `
mutation($input: RotateCloudAccountCredentialsInput!) {
  rotateCloudAccountCredentials(input: $input) {
    cloudAccount {
      id
      status
    }
    verifyStatus
    lastProbeError
  }
}
`

type rotateResponse struct {
	RotateCloudAccountCredentials struct {
		CloudAccount struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"cloudAccount"`
		VerifyStatus   string  `json:"verifyStatus"`
		LastProbeError *string `json:"lastProbeError"`
	} `json:"rotateCloudAccountCredentials"`
}

func NewCmdRotate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProvider       string
		flagCredentialKind string
		flagAWSRoleARN     string
		flagAWSExternalID  string
		flagAzureTenantID  string
		flagAzureClientID  string
	)

	cmd := &cobra.Command{
		Use:   "rotate <id>",
		Short: "Rotate non-secret cloud-account metadata (AWS role / external_id, Azure tenant / client)",
		Long: `Rotate the non-secret AWS / Azure metadata on an existing cloud
account.

Secret credential bodies (GCP service-account JSON, Azure
client_secret) MUST be uploaded out-of-band via
/api/console/v1/cloud-accounts/credentials/upload, NOT via this
command. After upload the row's status flips to PENDING_VERIFICATION;
run 'prb cloud-account verify <id>' to drive the next probe.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("provider", flagProvider, []string{"AWS", "GCP", "AZURE"}); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("credential-kind", flagCredentialKind, []string{"AWS_ASSUME_ROLE", "GCP_SERVICE_ACCOUNT_KEY", "AZURE_CLIENT_SECRET"}); err != nil {
				return err
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

			input := map[string]any{
				"cloudAccountId": args[0],
				"provider":       flagProvider,
				"credentialKind": flagCredentialKind,
			}
			if flagAWSRoleARN != "" {
				input["awsRoleArn"] = flagAWSRoleARN
			}
			if flagAWSExternalID != "" {
				input["awsExternalId"] = flagAWSExternalID
			}
			if flagAzureTenantID != "" {
				input["azureTenantId"] = flagAzureTenantID
			}
			if flagAzureClientID != "" {
				input["azureClientId"] = flagAzureClientID
			}

			data, err := client.Do(
				rotateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp rotateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Rotated credentials for cloud account %s\n", resp.RotateCloudAccountCredentials.CloudAccount.ID)
			_, _ = fmt.Fprintf(out, "Verify Status: %s\n", resp.RotateCloudAccountCredentials.VerifyStatus)
			if resp.RotateCloudAccountCredentials.LastProbeError != nil {
				_, _ = fmt.Fprintf(out, "Last Probe Error: %s\n", *resp.RotateCloudAccountCredentials.LastProbeError)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProvider, "provider", "", "Provider (AWS, GCP, AZURE) (must match the existing row) (required)")
	cmd.Flags().StringVar(&flagCredentialKind, "credential-kind", "", "Credential kind (must match the existing row) (required)")
	cmd.Flags().StringVar(&flagAWSRoleARN, "aws-role-arn", "", "New AWS role ARN (AWS only)")
	cmd.Flags().StringVar(&flagAWSExternalID, "aws-external-id", "", "New AWS external_id (AWS only)")
	cmd.Flags().StringVar(&flagAzureTenantID, "azure-tenant-id", "", "New Azure tenant id (Azure only)")
	cmd.Flags().StringVar(&flagAzureClientID, "azure-client-id", "", "New Azure client id (Azure only)")

	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("credential-kind")

	return cmd
}
