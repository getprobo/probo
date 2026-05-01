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
		Args:    cobra.ExactArgs(1),
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
