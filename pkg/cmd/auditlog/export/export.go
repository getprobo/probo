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

package export

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const exportMutation = `
mutation($input: RequestAuditLogExportInput!) {
  requestAuditLogExport(input: $input) {
    exportJobId
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
		Short: "Export audit log entries",
		Example: `  prb audit-log export --org <id> --from 2026-01-01T00:00:00Z --to 2026-02-01T00:00:00Z
  prb audit-log export --from 2026-03-01T00:00:00Z --to 2026-03-24T00:00:00Z`,
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
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
				RequestAuditLogExport struct {
					ExportJobID string `json:"exportJobId"`
				} `json:"requestAuditLogExport"`
			}

			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Audit log export requested %s\nYou will receive an email with a download link when the export is ready.\n",
				resp.RequestAuditLogExport.ExportJobID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagFrom, "from", "", "Start time in RFC3339 format (e.g. 2026-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&flagTo, "to", "", "End time in RFC3339 format (e.g. 2026-02-01T00:00:00Z)")

	return cmd
}
