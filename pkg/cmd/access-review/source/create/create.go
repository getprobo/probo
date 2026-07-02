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
mutation($input: CreateAccessReviewSourceInput!) {
  createAccessReviewSource(input: $input) {
    accessReviewSourceEdge {
      node {
        id
        name
      }
    }
  }
}
`

type createResponse struct {
	CreateAccessReviewSource struct {
		AccessReviewSourceEdge struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		} `json:"accessReviewSourceEdge"`
	} `json:"createAccessReviewSource"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagCSVFile     string
		flagConnectorID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an access source",
		Example: `  # Create an access source from a CSV file
  prb access-review source create --name "Okta Users" --csv-file users.csv

  # Create an access source with a connector
  prb access-review source create --name "GitHub" --connector-id <connector-id>`,
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

			if flagCSVFile != "" && flagConnectorID != "" {
				return fmt.Errorf("cannot specify both --csv-file and --connector-id")
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
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			s := resp.CreateAccessReviewSource.AccessReviewSourceEdge.Node
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Created access source %s\n", s.ID)
			_, _ = fmt.Fprintf(out, "Name: %s\n", s.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Access source name (required)")
	cmd.Flags().StringVar(&flagCSVFile, "csv-file", "", "Path to CSV file with access data")
	cmd.Flags().StringVar(&flagConnectorID, "connector-id", "", "Connector ID to use as data source")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
