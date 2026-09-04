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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const connectionStatusConnected = "CONNECTED"

const (
	createMutation = `
mutation($input: CreateAccessReviewSourceInput!) {
  createAccessReviewSource(input: $input) {
    created
    accessReviewSourceEdge {
      node {
        id
        name
      }
    }
  }
}
`

	createWorkloadIdentityConnectorMutation = `
mutation($input: CreateWorkloadIdentityConnectorInput!) {
  createWorkloadIdentityConnector(input: $input) {
    connector {
      id
      connectionStatus
    }
  }
}
`

	deleteConnectorMutation = `
mutation($input: DeleteConnectorInput!) {
  deleteConnector(input: $input) {
    deletedConnectorId
  }
}
`
)

type (
	createResponse struct {
		CreateAccessReviewSource struct {
			Created                bool `json:"created"`
			AccessReviewSourceEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"accessReviewSourceEdge"`
		} `json:"createAccessReviewSource"`
	}

	createWorkloadIdentityConnectorResponse struct {
		CreateWorkloadIdentityConnector struct {
			Connector struct {
				ID               string `json:"id"`
				ConnectionStatus string `json:"connectionStatus"`
			} `json:"connector"`
		} `json:"createWorkloadIdentityConnector"`
	}
)

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                         string
		flagName                        string
		flagCSVFile                     string
		flagConnectorID                 string
		flagRoleARN                     string
		flagGCPWorkloadIdentityProvider string
		flagGCPServiceAccountEmail      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an access source",
		Example: `  # Create an access source from a CSV file
  prb access-review source create --name "Okta Users" --csv-file users.csv

  # Create an access source with a connector
  prb access-review source create --name "GitHub" --connector-id <connector-id>

  # Create an AWS workload-identity access source
  prb access-review source create --name "AWS prod" --aws-role-arn arn:aws:iam::123456789012:role/ProboAudit

  # Create a GCP workload-identity access source
  prb access-review source create --name "GCP prod" \
    --gcp-workload-identity-provider projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo \
    --gcp-service-account-email probo-audit@my-project.iam.gserviceaccount.com`,
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

			var createdConnectorID string

			if flagRoleARN != "" {
				connectorID, status, err := createAWSConnector(client, flagOrg, flagRoleARN)
				if err != nil {
					return err
				}

				createdConnectorID = connectorID
				flagConnectorID = connectorID

				if status != connectionStatusConnected {
					return abandonCreatedConnector(
						client,
						createdConnectorID,
						fmt.Errorf("connector is %s", status),
					)
				}
			}

			if flagGCPWorkloadIdentityProvider != "" {
				connectorID, status, err := createGCPConnector(
					client,
					flagOrg,
					flagGCPWorkloadIdentityProvider,
					flagGCPServiceAccountEmail,
				)
				if err != nil {
					return err
				}

				createdConnectorID = connectorID
				flagConnectorID = connectorID

				if status != connectionStatusConnected {
					return abandonCreatedConnector(
						client,
						createdConnectorID,
						fmt.Errorf("connector is %s", status),
					)
				}
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagCSVFile != "" {
				csvData, err := os.ReadFile(flagCSVFile)
				if err != nil {
					return fmt.Errorf("cannot read CSV file: %w", err)
				}

				input["csvData"] = string(csvData)
			}

			if flagConnectorID != "" {
				input["connectorId"] = flagConnectorID
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return abandonCreatedConnector(client, createdConnectorID, err)
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return abandonCreatedConnector(
					client,
					createdConnectorID,
					fmt.Errorf("cannot parse response: %w", err),
				)
			}

			s := resp.CreateAccessReviewSource.AccessReviewSourceEdge.Node
			out := f.IOStreams.Out

			if resp.CreateAccessReviewSource.Created {
				_, _ = fmt.Fprintf(out, "Created access source %s\n", s.ID)
			} else {
				_, _ = fmt.Fprintf(out, "Access source %s already exists for this connector\n", s.ID)
			}

			_, _ = fmt.Fprintf(out, "Name: %s\n", s.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Access source name (required)")
	cmd.Flags().StringVar(&flagCSVFile, "csv-file", "", "Path to CSV file with access data")
	cmd.Flags().StringVar(&flagConnectorID, "connector-id", "", "Connector ID to use as data source")
	cmd.Flags().StringVar(&flagRoleARN, "aws-role-arn", "", "IAM role ARN")
	cmd.Flags().StringVar(
		&flagGCPWorkloadIdentityProvider,
		"gcp-workload-identity-provider",
		"",
		"GCP workload identity provider resource, including the S3NS IAM host when applicable",
	)
	cmd.Flags().StringVar(
		&flagGCPServiceAccountEmail,
		"gcp-service-account-email",
		"",
		"GCP service account email to impersonate, including the universe-specific suffix",
	)

	_ = cmd.MarkFlagRequired("name")
	cmd.MarkFlagsMutuallyExclusive(
		"csv-file",
		"connector-id",
		"aws-role-arn",
		"gcp-workload-identity-provider",
	)
	cmd.MarkFlagsRequiredTogether(
		"gcp-workload-identity-provider",
		"gcp-service-account-email",
	)

	return cmd
}

func createAWSConnector(
	client *api.Client,
	orgID string,
	roleARN string,
) (string, string, error) {
	return createWorkloadIdentityConnector(
		client,
		map[string]any{
			"organizationId": orgID,
			"provider":       "AWS",
			"awsRoleArn":     roleARN,
		},
	)
}

func createGCPConnector(
	client *api.Client,
	orgID string,
	providerResource string,
	serviceAccountEmail string,
) (string, string, error) {
	return createWorkloadIdentityConnector(
		client,
		map[string]any{
			"organizationId":              orgID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": providerResource,
			"gcpServiceAccountEmail":      serviceAccountEmail,
		},
	)
}

func createWorkloadIdentityConnector(
	client *api.Client,
	input map[string]any,
) (string, string, error) {
	data, err := client.Do(
		createWorkloadIdentityConnectorMutation,
		map[string]any{"input": input},
	)
	if err != nil {
		return "", "", err
	}

	var resp createWorkloadIdentityConnectorResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", "", fmt.Errorf("cannot parse response: %w", err)
	}

	cnnctr := resp.CreateWorkloadIdentityConnector.Connector

	return cnnctr.ID, cnnctr.ConnectionStatus, nil
}

func abandonCreatedConnector(client *api.Client, connectorID string, cause error) error {
	if connectorID == "" {
		return cause
	}

	_, err := client.Do(
		deleteConnectorMutation,
		map[string]any{
			"input": map[string]any{
				"connectorId": connectorID,
			},
		},
	)
	if err != nil {
		return errors.Join(
			cause,
			fmt.Errorf("cannot delete leftover connector %s: %w", connectorID, err),
		)
	}

	return cause
}
