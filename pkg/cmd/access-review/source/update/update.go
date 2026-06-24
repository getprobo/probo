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

package update

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateAccessReviewSourceInput!) {
  updateAccessReviewSource(input: $input) {
    accessReviewSource {
      id
      name
    }
  }
}
`

type updateResponse struct {
	UpdateAccessReviewSource struct {
		AccessReviewSource struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"accessReviewSource"`
	} `json:"updateAccessReviewSource"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagCSVFile     string
		flagConnectorID string
		flagOutput      *string
	)

	cmd := &cobra.Command{
		Use:   "update <source-id>",
		Short: "Update an access source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
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
				"accessReviewSourceId": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("csv-file") {
				csvData, err := os.ReadFile(flagCSVFile)
				if err != nil {
					return fmt.Errorf("cannot read CSV file: %w", err)
				}

				input["csvData"] = string(csvData)
			}

			if cmd.Flags().Changed("connector-id") {
				input["connectorId"] = flagConnectorID
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			s := resp.UpdateAccessReviewSource.AccessReviewSource

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, s)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated access source %s\n", s.ID)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Name: %s\n", s.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Access source name")
	cmd.Flags().StringVar(&flagCSVFile, "csv-file", "", "Path to CSV file with access data")
	cmd.Flags().StringVar(&flagConnectorID, "connector-id", "", "Connector ID to use as data source")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
