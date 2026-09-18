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
    connectorAccountEdges {
      node {
        id
        externalAccountId
        name
      }
    }
  }
}
`

func NewCmdEnableAccounts(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagAccounts []string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "enable-accounts <connector-id>",
		Short: "Record vendor accounts under a connector",
		Long: "Record accounts so module sources can attach to them. Enabling is " +
			"idempotent per vendor identifier and never creates a source.\n\n" +
			"Each --account is EXTERNAL_ID or EXTERNAL_ID=NAME.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if len(flagAccounts) == 0 {
				return fmt.Errorf("at least one --account is required")
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

			accounts := make([]map[string]any, 0, len(flagAccounts))

			for _, raw := range flagAccounts {
				externalID, name, hasName := strings.Cut(raw, "=")

				externalID = strings.TrimSpace(externalID)
				if externalID == "" {
					return fmt.Errorf("invalid --account %q: external id is empty", raw)
				}

				account := map[string]any{"externalAccountId": externalID}
				if hasName {
					account["name"] = strings.TrimSpace(name)
				}

				accounts = append(accounts, account)
			}

			data, err := client.Do(enableMutation, map[string]any{
				"input": map[string]any{
					"organizationId": flagOrg,
					"connectorId":    args[0],
					"accounts":       accounts,
				},
			})
			if err != nil {
				return err
			}

			var resp struct {
				EnableConnectorAccounts struct {
					ConnectorAccountEdges []struct {
						Node struct {
							ID                string  `json:"id"`
							ExternalAccountID *string `json:"externalAccountId"`
							Name              string  `json:"name"`
						} `json:"node"`
					} `json:"connectorAccountEdges"`
				} `json:"enableConnectorAccounts"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			edges := resp.EnableConnectorAccounts.ConnectorAccountEdges

			if *flagOutput == cmdutil.OutputJSON {
				nodes := make([]any, 0, len(edges))
				for _, edge := range edges {
					nodes = append(nodes, edge.Node)
				}

				return cmdutil.PrintJSON(f.IOStreams.Out, nodes)
			}

			rows := make([][]string, 0, len(edges))
			for _, edge := range edges {
				external := "(implicit)"
				if edge.Node.ExternalAccountID != nil {
					external = *edge.Node.ExternalAccountID
				}

				rows = append(rows, []string{edge.Node.ID, external, edge.Node.Name})
			}

			_, _ = fmt.Fprintln(
				f.IOStreams.Out,
				cmdutil.NewTable("ID", "EXTERNAL ID", "NAME").Rows(rows...),
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringArrayVar(&flagAccounts, "account", nil, "Account to enable, EXTERNAL_ID or EXTERNAL_ID=NAME (repeatable)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
