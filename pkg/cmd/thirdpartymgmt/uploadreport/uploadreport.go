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

package uploadreport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
        validUntil
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
				ValidUntil string `json:"validUntil"`
			} `json:"node"`
		} `json:"thirdPartyComplianceReportEdge"`
	} `json:"uploadThirdPartyComplianceReport"`
}

// parseDate accepts either a plain date (2026-03-31) or a full RFC 3339
// timestamp and normalizes it to RFC 3339 for the Datetime scalar.
func parseDate(value string) (string, error) {
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("invalid date %q: expected YYYY-MM-DD or RFC 3339", value)
}

func NewCmdUploadReport(f *cmdutil.Factory) *cobra.Command {
	var (
		flagThirdParty string
		flagName       string
		flagReportDate string
		flagValidUntil string
	)

	cmd := &cobra.Command{
		Use:   "upload-report <file>",
		Short: "Upload a compliance report for a thirdParty",
		Example: `  # Upload a SOC 2 report for a third party
  prb thirdParty upload-report ./soc2.pdf \
    --third-party <third-party-id> \
    --name "Acme Corp - SOC 2 Type 2 - 2026-03-31" \
    --report-date 2026-03-31 \
    --valid-until 2027-03-31`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			reportDate, err := parseDate(flagReportDate)
			if err != nil {
				return err
			}

			input := map[string]any{
				"thirdPartyId": flagThirdParty,
				"reportName":   flagName,
				"reportDate":   reportDate,
				"file":         nil,
			}

			if flagValidUntil != "" {
				validUntil, err := parseDate(flagValidUntil)
				if err != nil {
					return err
				}

				input["validUntil"] = validUntil
			}

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

			variables := map[string]any{
				"input": input,
			}

			data, err := client.DoUpload(
				uploadMutation,
				variables,
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
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Uploaded compliance report %s (%s)\n", node.ID, node.ReportName)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagThirdParty, "third-party", "", "ThirdParty ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Report name (required)")
	cmd.Flags().StringVar(&flagReportDate, "report-date", "", "Report date, YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Valid-until date, YYYY-MM-DD (optional)")
	_ = cmd.MarkFlagRequired("third-party")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("report-date")

	return cmd
}
