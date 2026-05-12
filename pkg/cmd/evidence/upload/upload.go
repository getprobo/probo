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
