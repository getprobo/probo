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
	"os"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateAccessReviewSourcesInput!) {
  createAccessReviewSources(input: $input) {
    results {
      created
      accessReviewSourceEdge {
        node {
          id
          name
        }
      }
    }
  }
}
`

type createResponse struct {
	CreateAccessReviewSources struct {
		Results []struct {
			Created                bool `json:"created"`
			AccessReviewSourceEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"accessReviewSourceEdge"`
		} `json:"results"`
	} `json:"createAccessReviewSources"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                string
		flagName               string
		flagCSVFile            string
		flagConnectorID        string
		flagConnectorAccountID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an access source",
		Long: `Connect the provider first. Create the connector with the generic form, or with Create AWS, GCP, or Azure Workload Identity Federation, then pass that connector id here. This operation does not collect credentials.

A CSV source is the path with no provider.`,
		Example: `  # Create an access source from a connector that already exists
  prb access-review source create --name "GitHub" --connector-id <connector-id>

  # Create an access source for one account on that connector
  prb access-review source create --name "GitHub" --connector-id <connector-id> \
    --connector-account-id <connector-account-id>

  # Create an access source from a CSV file
  prb access-review source create --name "Okta Users" --csv-file users.csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagConnectorID == "" && flagCSVFile == "" {
				return fmt.Errorf("connect the provider first, then pass --connector-id, or pass --csv-file for a source with no provider")
			}

			if flagConnectorAccountID != "" && flagConnectorID == "" {
				return fmt.Errorf("--connector-account-id requires --connector-id")
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

			source := map[string]any{
				"name": flagName,
			}

			if flagCSVFile != "" {
				csvData, err := os.ReadFile(flagCSVFile)
				if err != nil {
					return fmt.Errorf("cannot read CSV file: %w", err)
				}

				source["csvData"] = string(csvData)
			}

			if flagConnectorID != "" {
				source["connectorId"] = flagConnectorID
			}

			if flagConnectorAccountID != "" {
				source["connectorAccountId"] = flagConnectorAccountID
			}

			data, err := client.Do(
				createMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId": flagOrg,
						"sources":        []any{source},
					},
				},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if len(resp.CreateAccessReviewSources.Results) != 1 {
				return fmt.Errorf("cannot parse response: expected one access source")
			}

			result := resp.CreateAccessReviewSources.Results[0]
			s := result.AccessReviewSourceEdge.Node
			out := f.IOStreams.Out

			if result.Created {
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
	cmd.Flags().StringVar(&flagCSVFile, "csv-file", "", "Path to CSV file with access data. A CSV source has no provider.")
	cmd.Flags().StringVar(&flagConnectorID, "connector-id", "", "Connector ID. Connect the provider first; this command does not collect credentials.")
	cmd.Flags().StringVar(&flagConnectorAccountID, "connector-account-id", "", "Connector account ID (defaults to the account named by the connector, or its only account)")

	_ = cmd.MarkFlagRequired("name")
	cmd.MarkFlagsMutuallyExclusive("csv-file", "connector-id")

	return cmd
}
