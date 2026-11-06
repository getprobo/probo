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

package recordnda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/complianceportal/management"
)

const recordMutation = `
mutation($input: UploadCompliancePortalAccessRecordedNDAInput!) {
  uploadCompliancePortalAccessRecordedNDA(input: $input) {
    compliancePortalAccess {
      id
      recordedNdaSignedAt
      recordedNda { fileName }
    }
  }
}
`

func NewCmdRecordNDA(f *cmdutil.Factory) *cobra.Command {
	var signedAt string

	cmd := &cobra.Command{
		Use:   "record-nda <id> <file>",
		Short: "Store an NDA signed outside Probo",
		Example: `  # Keep a PDF the visitor already signed in another system
  prb compliance-portal access record-nda <access-id> ./signed-nda.pdf --signed-at 2026-03-01T00:00:00Z`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			signed, err := time.Parse(time.RFC3339, signedAt)
			if err != nil {
				return fmt.Errorf("cannot parse --signed-at: use RFC3339, for example 2026-03-01T00:00:00Z")
			}

			filePath := args[1]

			info, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}

			if info.Size() > management.RecordedNDAMaxBytes {
				return fmt.Errorf("file exceeds the 10 MB recorded NDA limit")
			}

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}

			defer func() { _ = file.Close() }()

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("cannot load config: %w", err)
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return fmt.Errorf("cannot load default host: %w", err)
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			fileName := filepath.Base(filePath)
			input := map[string]any{
				"compliancePortalAccessId": args[0],
				"fileName":                 fileName,
				"file":                     nil,
				"signedAt":                 signed.Format(time.RFC3339),
			}

			data, err := client.DoUpload(
				recordMutation,
				map[string]any{"input": input},
				"variables.input.file",
				fileName,
				file,
			)
			if err != nil {
				return fmt.Errorf("cannot record signed NDA: %w", err)
			}

			var resp struct {
				UploadCompliancePortalAccessRecordedNDA struct {
					CompliancePortalAccess struct {
						ID                  string `json:"id"`
						RecordedNdaSignedAt string `json:"recordedNdaSignedAt"`
						RecordedNda         struct {
							FileName string `json:"fileName"`
						} `json:"recordedNda"`
					} `json:"compliancePortalAccess"`
				} `json:"uploadCompliancePortalAccessRecordedNDA"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			access := resp.UploadCompliancePortalAccessRecordedNDA.CompliancePortalAccess
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Recorded %s for visitor access %s\n",
				access.RecordedNda.FileName,
				access.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&signedAt, "signed-at", "", "When the NDA was signed (RFC3339)")
	_ = cmd.MarkFlagRequired("signed-at")

	return cmd
}
