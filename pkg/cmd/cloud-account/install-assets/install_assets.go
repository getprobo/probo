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

package installassets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const generateMutation = `
mutation($input: GenerateCloudAccountInstallAssetsInput!) {
  generateCloudAccountInstallAssets(input: $input) {
    assets {
      __typename
      ... on AWSInstallAssets {
        quickCreateURL
        externalId
        principalArn
        requiredActions
      }
      ... on GCPInstallAssets {
        setupScript
        requiredRoles
        requiredApis
      }
      ... on AzureInstallAssets {
        steps { title body code }
        requiredRbacRoles
        requiredGraphPermissions
      }
    }
  }
}
`

type assetsResponse struct {
	GenerateCloudAccountInstallAssets struct {
		Assets struct {
			Typename string `json:"__typename"`

			// AWS
			QuickCreateURL  string   `json:"quickCreateURL,omitempty"`
			ExternalID      string   `json:"externalId,omitempty"`
			PrincipalArn    string   `json:"principalArn,omitempty"`
			RequiredActions []string `json:"requiredActions,omitempty"`

			// GCP
			SetupScript   string   `json:"setupScript,omitempty"`
			RequiredRoles []string `json:"requiredRoles,omitempty"`
			RequiredApis  []string `json:"requiredApis,omitempty"`

			// Azure
			Steps []struct {
				Title string  `json:"title"`
				Body  string  `json:"body"`
				Code  *string `json:"code"`
			} `json:"steps,omitempty"`
			RequiredRbacRoles        []string `json:"requiredRbacRoles,omitempty"`
			RequiredGraphPermissions []string `json:"requiredGraphPermissions,omitempty"`
		} `json:"assets"`
	} `json:"generateCloudAccountInstallAssets"`
}

func NewCmdInstallAssets(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg             string
		flagProvider        string
		flagScopeKind       string
		flagScopeIdentifier string
		flagModules         []string
		flagAWSRegion       string
		flagOutput          *string
	)

	cmd := &cobra.Command{
		Use:   "install-assets",
		Short: "Generate per-provider install assets (CFN URL, gcloud script, Azure guide)",
		Long: `Generate the per-provider install assets a customer needs to wire
their cloud account into Probo.

For AWS this is a CloudFormation Quick-Create URL plus the persisted
external_id (use --output json to capture the external_id and feed it
into 'prb cloud-account create --aws-external-id ...').

For GCP this is a gcloud setup script.

For Azure this is a step-by-step install guide.

This is a mutating action for AWS (it persists the external_id on the
row) and requires generate-install-assets permission.`,
		Example: `  # Generate AWS Quick-Create URL and capture the external_id
  prb cloud-account install-assets \
    --provider AWS \
    --scope-kind AWS_ACCOUNT \
    --scope-identifier 123456789012 \
    --modules ACCESS_REVIEW \
    --aws-region us-east-1 \
    --output json

  # Generate GCP setup script for a project-scoped install
  prb cloud-account install-assets \
    --provider GCP \
    --scope-kind GCP_PROJECT \
    --scope-identifier my-project-123 \
    --modules ACCESS_REVIEW`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum("provider", flagProvider, []string{"AWS", "GCP", "AZURE"}); err != nil {
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
				"organizationId":  flagOrg,
				"provider":        flagProvider,
				"scopeKind":       flagScopeKind,
				"scopeIdentifier": flagScopeIdentifier,
				"modules":         flagModules,
			}
			if flagAWSRegion != "" {
				input["awsRegion"] = flagAWSRegion
			}

			data, err := client.Do(
				generateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp assetsResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			assets := resp.GenerateCloudAccountInstallAssets.Assets

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, assets)
			}

			out := f.IOStreams.Out
			switch assets.Typename {
			case "AWSInstallAssets":
				_, _ = fmt.Fprintln(out, "AWS Install Assets")
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintf(out, "Quick-Create URL:\n  %s\n\n", assets.QuickCreateURL)
				_, _ = fmt.Fprintf(out, "External ID (paste into the CFN stack):\n  %s\n\n", assets.ExternalID)
				_, _ = fmt.Fprintf(out, "Principal ARN:\n  %s\n\n", assets.PrincipalArn)
				_, _ = fmt.Fprintln(out, "Required IAM actions:")
				for _, a := range assets.RequiredActions {
					_, _ = fmt.Fprintf(out, "  - %s\n", a)
				}
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintln(out, "Next: 'prb cloud-account create --provider AWS --aws-role-arn ... --aws-external-id "+assets.ExternalID+" ...'")
			case "GCPInstallAssets":
				_, _ = fmt.Fprintln(out, "GCP Install Assets")
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintln(out, "Setup script:")
				_, _ = fmt.Fprintln(out, assets.SetupScript)
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintln(out, "Required roles: "+strings.Join(assets.RequiredRoles, ", "))
				_, _ = fmt.Fprintln(out, "Required APIs: "+strings.Join(assets.RequiredApis, ", "))
			case "AzureInstallAssets":
				_, _ = fmt.Fprintln(out, "Azure Install Assets")
				_, _ = fmt.Fprintln(out)
				for i, s := range assets.Steps {
					_, _ = fmt.Fprintf(out, "%d. %s\n", i+1, s.Title)
					_, _ = fmt.Fprintln(out, s.Body)
					if s.Code != nil && *s.Code != "" {
						_, _ = fmt.Fprintln(out, *s.Code)
					}
					_, _ = fmt.Fprintln(out)
				}
				_, _ = fmt.Fprintln(out, "Required RBAC roles: "+strings.Join(assets.RequiredRbacRoles, ", "))
				_, _ = fmt.Fprintln(out, "Required Graph permissions: "+strings.Join(assets.RequiredGraphPermissions, ", "))
			default:
				return fmt.Errorf("unexpected install assets payload type: %s", assets.Typename)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Provider (AWS, GCP, AZURE) (required)")
	cmd.Flags().StringVar(&flagScopeKind, "scope-kind", "", "Scope kind (required)")
	cmd.Flags().StringVar(&flagScopeIdentifier, "scope-identifier", "", "Scope identifier (AWS account id / GCP project or org id / Azure id) (required)")
	cmd.Flags().StringSliceVar(&flagModules, "modules", nil, "Audit modules to enable (ACCESS_REVIEW) (required)")
	cmd.Flags().StringVar(&flagAWSRegion, "aws-region", "", "AWS region for the CloudFormation Quick-Create URL (AWS only)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("scope-kind")
	_ = cmd.MarkFlagRequired("scope-identifier")
	_ = cmd.MarkFlagRequired("modules")

	return cmd
}
