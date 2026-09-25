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

package enableaccounts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const enableMutation = `
mutation($input: EnableConnectorAccountsInput!) {
  enableConnectorAccounts(input: $input) {
    connectorAccounts {
      id
      externalAccountId
      name
    }
  }
}
`

func NewCmdEnableAccounts(f *cmdutil.Factory) *cobra.Command {
	var (
		flagAccounts []string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "enable-accounts <connector-id>",
		Short: "Enable discovered connector accounts",
		Example: `  # Enable one account
  prb connector enable-accounts <connector-id> --account 123456789012=Production

  # Enable several accounts
  prb connector enable-accounts <connector-id> \
    --account 123456789012=Production \
    --account 111111111111=Staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if len(flagAccounts) == 0 {
				return fmt.Errorf("at least one --account is required")
			}

			accounts := make([]map[string]string, 0, len(flagAccounts))
			for _, raw := range flagAccounts {
				externalID, name, err := parseAccount(raw)
				if err != nil {
					return err
				}

				accounts = append(accounts, map[string]string{
					"externalAccountId": externalID,
					"name":              name,
				})
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

			data, err := client.Do(
				enableMutation,
				map[string]any{
					"input": map[string]any{
						"connectorId": args[0],
						"accounts":    accounts,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp struct {
				EnableConnectorAccounts struct {
					ConnectorAccounts []struct {
						ID                string `json:"id"`
						ExternalAccountID string `json:"externalAccountId"`
						Name              string `json:"name"`
					} `json:"connectorAccounts"`
				} `json:"enableConnectorAccounts"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			enabled := resp.EnableConnectorAccounts.ConnectorAccounts

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, enabled)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Enabled %d account(s)\n", len(enabled))

			rows := make([][]string, 0, len(enabled))
			for _, a := range enabled {
				rows = append(rows, []string{a.ID, a.ExternalAccountID, a.Name})
			}

			t := cmdutil.NewTable("ID", "EXTERNAL ID", "NAME").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	cmd.Flags().StringArrayVar(
		&flagAccounts,
		"account",
		nil,
		"Account to enable as external-id=name (repeatable)",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func parseAccount(raw string) (externalID string, name string, err error) {
	externalID, name, ok := strings.Cut(raw, "=")
	if !ok || externalID == "" {
		return "", "", fmt.Errorf("invalid --account %q: expected external-id=name", raw)
	}

	if name == "" {
		name = externalID
	}

	return externalID, name, nil
}
