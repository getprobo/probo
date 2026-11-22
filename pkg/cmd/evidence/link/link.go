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

package link

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkMutation = `
mutation EvidenceLink($input: UploadInternalControlEvidenceInput!) {
  uploadInternalControlEvidence(input: $input) {
    evidenceEdge {
      node {
        id
      }
    }
  }
}
`

type linkResponse struct {
	UploadInternalControlEvidence struct {
		EvidenceEdge struct {
			Node struct {
				ID string `json:"id"`
			} `json:"node"`
		} `json:"evidenceEdge"`
	} `json:"uploadInternalControlEvidence"`
}

func NewCmdLink(f *cmdutil.Factory) *cobra.Command {
	var (
		flagInternalControl string
		flagURL             string
		flagName            string
	)

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Upload a .uri file as evidence for an internal control",
		Example: `  # Attach a policy URL to an internal control
  prb evidence link --internal-control <internal-control-id> --url https://example.com/policy --name "Access policy"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fileName, content, err := linkEvidenceFile(flagName, flagURL)
			if err != nil {
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

			data, err := client.DoUpload(
				linkMutation,
				map[string]any{
					"input": map[string]any{
						"internalControlId": flagInternalControl,
						"file":              nil,
					},
				},
				"variables.input.file",
				fileName,
				bytes.NewReader(content),
			)
			if err != nil {
				return err
			}

			var resp linkResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Uploaded evidence %s\n",
				resp.UploadInternalControlEvidence.EvidenceEdge.Node.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagInternalControl, "internal-control", "", "Internal control ID (required)")
	cmd.Flags().StringVar(&flagURL, "url", "", "HTTP or HTTPS URL (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "File name without the .uri extension (required)")
	_ = cmd.MarkFlagRequired("internal-control")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// linkEvidenceFile builds the text/uri-list payload the console link form uploads.
func linkEvidenceFile(name, rawURL string) (string, []byte, error) {
	name = strings.TrimSpace(name)
	rawURL = strings.TrimSpace(rawURL)

	if name == "" {
		return "", nil, fmt.Errorf("name is required")
	}

	if strings.ContainsAny(name, "\r\n") {
		return "", nil, fmt.Errorf("name must be a single line")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", nil, fmt.Errorf("url must be an http or https address")
	}

	return name + ".uri", []byte(rawURL), nil
}
