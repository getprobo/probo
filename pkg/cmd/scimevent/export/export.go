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

package export

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const exportMutation = `
mutation($input: RequestSCIMEventExportInput!) {
  requestSCIMEventExport(input: $input) {
    logExportId
  }
}
`

func NewCmdExport(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg  string
		flagFrom string
		flagTo   string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export SCIM events",
		Example: `  prb scim-event export --org <id> --from 2026-01-01T00:00:00Z --to 2026-02-01T00:00:00Z
  prb scim-event export --from 2026-03-01T00:00:00Z --to 2026-03-24T00:00:00Z`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if flagFrom == "" {
				return fmt.Errorf("--from is required (RFC3339 timestamp)")
			}

			if flagTo == "" {
				return fmt.Errorf("--to is required (RFC3339 timestamp)")
			}

			variables := map[string]any{
				"input": map[string]any{
					"organizationId": flagOrg,
					"fromTime":       flagFrom,
					"toTime":         flagTo,
				},
			}

			data, err := client.Do(exportMutation, variables)
			if err != nil {
				return err
			}

			var resp struct {
				RequestSCIMEventExport struct {
					LogExportID string `json:"logExportId"`
				} `json:"requestSCIMEventExport"`
			}

			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"SCIM event export requested %s\nYou will receive an email with a download link when the export is ready.\n",
				resp.RequestSCIMEventExport.LogExportID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagFrom, "from", "", "Start time in RFC3339 format (e.g. 2026-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&flagTo, "to", "", "End time in RFC3339 format (e.g. 2026-02-01T00:00:00Z)")

	return cmd
}
