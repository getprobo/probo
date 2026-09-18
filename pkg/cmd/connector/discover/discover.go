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

package discover

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const discoverQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Connector {
      discoveredAccounts {
        externalAccountId
        name
        enabled
      }
    }
  }
}
`

type discoveredAccount struct {
	ExternalAccountID string `json:"externalAccountId"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
}

func NewCmdDiscover(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "discover <connector-id>",
		Short: "List the accounts a connector's credential can reach right now",
		Long: "Ask the vendor which accounts this credential can reach. This makes a " +
			"live outbound call, so it is slower than 'connector view' and reflects " +
			"the cloud rather than what Probo has recorded.",
		Args: cobra.ExactArgs(1),
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

			data, err := client.Do(discoverQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp struct {
				Node *struct {
					Typename           string              `json:"__typename"`
					DiscoveredAccounts []discoveredAccount `json:"discoveredAccounts"`
				} `json:"node"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			if resp.Node == nil {
				return fmt.Errorf("connector %s not found", args[0])
			}

			if resp.Node.Typename != "Connector" {
				return fmt.Errorf("expected Connector node, got %s", resp.Node.Typename)
			}

			accounts := resp.Node.DiscoveredAccounts

			if *flagOutput == cmdutil.OutputJSON {
				if accounts == nil {
					accounts = []discoveredAccount{}
				}

				return cmdutil.PrintJSON(f.IOStreams.Out, accounts)
			}

			if len(accounts) == 0 {
				_, _ = fmt.Fprintln(
					f.IOStreams.Out,
					"No accounts discovered. This provider may not support organization-wide install.",
				)

				return nil
			}

			rows := make([][]string, 0, len(accounts))
			for _, a := range accounts {
				enabled := "no"
				if a.Enabled {
					enabled = "yes"
				}

				rows = append(rows, []string{a.ExternalAccountID, a.Name, enabled})
			}

			_, _ = fmt.Fprintln(
				f.IOStreams.Out,
				cmdutil.NewTable("EXTERNAL ID", "NAME", "ENABLED").Rows(rows...),
			)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
