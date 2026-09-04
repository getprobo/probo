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

package getgvlcatalog

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const catalogQuery = `
query {
  commonGVLCatalog {
    vendorListVersion
    tcfPolicyVersion
  }
}
`

type catalog struct {
	VendorListVersion *int `json:"vendorListVersion"`
	TcfPolicyVersion  *int `json:"tcfPolicyVersion"`
}

func NewCmdGetGVLCatalog(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "get-gvl-catalog",
		Short: "Show the current IAB GVL catalog versions",
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

			data, err := client.Do(catalogQuery, nil)
			if err != nil {
				return err
			}

			var resp struct {
				CommonGVLCatalog catalog `json:"commonGVLCatalog"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.CommonGVLCatalog)
			}

			rows := [][]string{{
				formatOptionalInt(resp.CommonGVLCatalog.VendorListVersion),
				formatOptionalInt(resp.CommonGVLCatalog.TcfPolicyVersion),
			}}
			t := cmdutil.NewTable("GVL", "TCF POLICY").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func formatOptionalInt(value *int) string {
	if value == nil {
		return "—"
	}

	return strconv.Itoa(*value)
}
