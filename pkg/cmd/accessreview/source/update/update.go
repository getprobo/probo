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
mutation($input: UpdateAccessSourceInput!) {
  updateAccessSource(input: $input) {
    accessSource {
      id
      name
    }
  }
}
`

type updateResponse struct {
	UpdateAccessSource struct {
		AccessSource struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"accessSource"`
	} `json:"updateAccessSource"`
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
			)

			input := map[string]any{
				"accessSourceId": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if flagCSVFile != "" {
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

			s := resp.UpdateAccessSource.AccessSource

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
