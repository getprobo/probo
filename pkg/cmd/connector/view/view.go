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

package view

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Connector {
      id
      provider
      protocol
      displayName
      connectionStatus
      createdAt
      accounts(first: 100) {
        totalCount
        edges {
          node {
            id
            externalAccountId
            name
          }
        }
      }
    }
  }
}
`

type accountNode struct {
	ID                string  `json:"id"`
	ExternalAccountID *string `json:"externalAccountId"`
	Name              string  `json:"name"`
}

type connectorDetail struct {
	ID               string `json:"id"`
	Provider         string `json:"provider"`
	Protocol         string `json:"protocol"`
	DisplayName      string `json:"displayName"`
	ConnectionStatus string `json:"connectionStatus"`
	CreatedAt        string `json:"createdAt"`
	Accounts         struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node accountNode `json:"node"`
		} `json:"edges"`
	} `json:"accounts"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <connector-id>",
		Short: "Show a connector and the accounts it covers",
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

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp struct {
				Node *connectorDetail `json:"node"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			if resp.Node == nil {
				return fmt.Errorf("connector %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "ID:       %s\n", resp.Node.ID)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Provider: %s\n", resp.Node.DisplayName)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Protocol: %s\n", resp.Node.Protocol)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Status:   %s\n", resp.Node.ConnectionStatus)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Created:  %s\n", cmdutil.FormatTime(resp.Node.CreatedAt))

			if len(resp.Node.Accounts.Edges) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "\nNo accounts recorded.")

				return nil
			}

			rows := make([][]string, 0, len(resp.Node.Accounts.Edges))
			for _, edge := range resp.Node.Accounts.Edges {
				external := "(implicit)"
				if edge.Node.ExternalAccountID != nil {
					external = *edge.Node.ExternalAccountID
				}

				rows = append(rows, []string{edge.Node.ID, external, edge.Node.Name})
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, "\nAccounts:")
			_, _ = fmt.Fprintln(f.IOStreams.Out, cmdutil.NewTable("ID", "EXTERNAL ID", "NAME").Rows(rows...))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
