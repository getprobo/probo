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

package providers

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/connector/catalog"
)

func NewCmdProviders(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "providers",
		Short: "List providers prb connector create can connect",
		Long:  "Lists providers that connect with an API key or client credentials and no extra setting. A provider that needs a base URL, slug, or region is connected in the console.",
		Args:  cobra.NoArgs,
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

			all, err := catalog.List(client)
			if err != nil {
				return err
			}

			listed := make([]catalog.Provider, 0, len(all))
			for _, provider := range all {
				if catalog.APIKeyCreatable(provider) || catalog.ClientCredentialsCreatable(provider) {
					listed = append(listed, provider)
				}
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, listed)
			}

			if len(listed) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No providers found.")
				return nil
			}

			rows := make([][]string, 0, len(listed))
			for _, provider := range listed {
				protocols := ""
				if catalog.APIKeyCreatable(provider) {
					protocols = "api-key"
				}

				if catalog.ClientCredentialsCreatable(provider) {
					if protocols != "" {
						protocols += ", "
					}

					protocols += "client-credentials"
				}

				rows = append(rows, []string{provider.Provider, provider.DisplayName, protocols})
			}

			t := cmdutil.NewTable("PROVIDER", "NAME", "PROTOCOL").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
