// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package upload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const uploadMutation = `
mutation($input: UploadThirdPartyComplianceReportInput!) {
  uploadThirdPartyComplianceReport(input: $input) {
    thirdPartyComplianceReportEdge {
      node {
        id
        reportName
        reportDate
      }
    }
  }
}
`

type uploadResponse struct {
	UploadThirdPartyComplianceReport struct {
		ThirdPartyComplianceReportEdge struct {
			Node struct {
				ID         string `json:"id"`
				ReportName string `json:"reportName"`
				ReportDate string `json:"reportDate"`
			} `json:"node"`
		} `json:"thirdPartyComplianceReportEdge"`
	} `json:"uploadThirdPartyComplianceReport"`
}

func NewCmdUpload(f *cmdutil.Factory) *cobra.Command {
	var (
		flagThirdParty string
		flagReportDate string
		flagReportName string
		flagValidUntil string
	)

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a compliance report for a third party",
		Example: `  # Upload a compliance report
  prb third-party report upload ./soc2.pdf --third-party <third-party-id> --report-date 2026-01-01

  # Upload with a custom name and validity date
  prb third-party report upload ./soc2.pdf --third-party <third-party-id> --report-date 2026-01-01 --report-name "SOC 2 Type II" --valid-until 2026-12-31`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}

			defer func() { _ = file.Close() }()

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

			if flagReportName == "" {
				flagReportName = filepath.Base(filePath)
			}

			input := map[string]any{
				"thirdPartyId": flagThirdParty,
				"reportDate":   flagReportDate,
				"reportName":   flagReportName,
				"file":         nil,
			}

			if flagValidUntil != "" {
				input["validUntil"] = flagValidUntil
			}

			data, err := client.DoUpload(
				uploadMutation,
				map[string]any{"input": input},
				"variables.input.file",
				filepath.Base(filePath),
				file,
			)
			if err != nil {
				return err
			}

			var resp uploadResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			node := resp.UploadThirdPartyComplianceReport.ThirdPartyComplianceReportEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Uploaded compliance report %s (%s)\n",
				node.ID,
				node.ReportName,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagThirdParty, "third-party", "", "Third party ID (required)")
	cmd.Flags().StringVar(&flagReportDate, "report-date", "", "Report date (e.g. 2026-01-01) (required)")
	cmd.Flags().StringVar(&flagReportName, "report-name", "", "Report name (defaults to the file name)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Valid until date (e.g. 2026-12-31)")
	_ = cmd.MarkFlagRequired("third-party")
	_ = cmd.MarkFlagRequired("report-date")

	return cmd
}
