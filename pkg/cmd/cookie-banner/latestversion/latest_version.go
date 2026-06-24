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

package latestversion

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const versionsQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on CookieBanner {
      latestVersion {
        id
        version
        state
        createdAt
        updatedAt
      }
    }
  }
}
`

type versionInfo struct {
	ID        string `json:"id"`
	Version   int    `json:"version"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func NewCmdLatestVersion(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "latest-version <id>",
		Short: "Show the latest version of a cookie banner",
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

			data, err := client.Do(versionsQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp struct {
				Node *struct {
					LatestVersion *versionInfo `json:"latestVersion"`
				} `json:"node"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			if resp.Node == nil || resp.Node.LatestVersion == nil {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No versions found.")
				return nil
			}

			v := resp.Node.LatestVersion

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, v)
			}

			rows := [][]string{
				{v.ID, strconv.Itoa(v.Version), v.State, cmdutil.FormatTime(v.CreatedAt)},
			}
			t := cmdutil.NewTable("ID", "VERSION", "STATE", "CREATED").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
