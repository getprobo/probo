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
	"go.probo.inc/probo/pkg/filevalidation"
)

const uploadMutation = `
mutation($input: UploadFileInput!) {
  uploadFile(input: $input) {
    file {
      id
      fileName
      mimeType
      size
    }
  }
}
`

type uploadResponse struct {
	UploadFile struct {
		File struct {
			ID       string `json:"id"`
			FileName string `json:"fileName"`
			MimeType string `json:"mimeType"`
			Size     int64  `json:"size"`
		} `json:"file"`
	} `json:"uploadFile"`
}

func NewCmdUpload(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a private file for rich text",
		Example: `  # Upload an image to use in a task or document
  prb file upload ./diagram.png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return fmt.Errorf("cannot validate output: %w", err)
			}

			filePath := args[0]

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

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
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

			input := map[string]any{
				"organizationId": flagOrg,
				"file":           nil,
			}

			fileName := filepath.Base(filePath)

			if _, ok := filevalidation.ContentTypeForAttachment(fileName); !ok {
				return fmt.Errorf("cannot upload file: %s has an unsupported type", fileName)
			}

			data, err := client.DoUpload(
				uploadMutation,
				map[string]any{"input": input},
				"variables.input.file",
				fileName,
				file,
			)
			if err != nil {
				return fmt.Errorf("cannot upload file: %w", err)
			}

			var resp uploadResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			uploaded := resp.UploadFile.File
			if *flagOutput == cmdutil.OutputJSON {
				if err := cmdutil.PrintJSON(f.IOStreams.Out, uploaded); err != nil {
					return fmt.Errorf("cannot print file: %w", err)
				}

				return nil
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Uploaded %s (%s)\n", uploaded.FileName, uploaded.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
