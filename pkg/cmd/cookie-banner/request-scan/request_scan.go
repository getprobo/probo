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

package requestscan

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const requestScanMutation = `
mutation($input: RequestCookieBannerScanInput!) {
  requestCookieBannerScan(input: $input) {
    cookieBannerScan {
      id
      state
      createdAt
    }
  }
}
`

func NewCmdRequestScan(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "request-scan <id>",
		Short: "Queue a cookie banner scan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(requestScanMutation, map[string]any{
				"input": map[string]any{"cookieBannerId": args[0]},
			})
			if err != nil {
				return err
			}

			var resp struct {
				RequestCookieBannerScan struct {
					CookieBannerScan struct {
						ID        string `json:"id"`
						State     string `json:"state"`
						CreatedAt string `json:"createdAt"`
					} `json:"cookieBannerScan"`
				} `json:"requestCookieBannerScan"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			scan := resp.RequestCookieBannerScan.CookieBannerScan
			if *flagOutput == "json" {
				return cmdutil.PrintJSON(f.IOStreams.Out, scan)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Queued cookie banner scan %s (%s)\n", scan.ID, scan.State)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
