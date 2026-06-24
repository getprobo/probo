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
mutation($input: UploadMeasureEvidenceInput!) {
  uploadMeasureEvidence(input: $input) {
    evidence {
      id
      state
      type
    }
  }
}
`

type uploadResponse struct {
	UploadMeasureEvidence struct {
		Evidence struct {
			ID    string `json:"id"`
			State string `json:"state"`
			Type  string `json:"type"`
		} `json:"evidence"`
	} `json:"uploadMeasureEvidence"`
}

func NewCmdUpload(f *cmdutil.Factory) *cobra.Command {
	var flagMeasure string

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload evidence for a measure",
		Example: `  # Upload a file as evidence for a measure
  prb evidence upload ./report.pdf --measure <measure-id>`,
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

			variables := map[string]any{
				"input": map[string]any{
					"measureId": flagMeasure,
					"file":      nil,
				},
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

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Uploaded evidence %s\n", resp.UploadMeasureEvidence.Evidence.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagMeasure, "measure", "", "Measure ID (required)")
	_ = cmd.MarkFlagRequired("measure")

	return cmd
}
