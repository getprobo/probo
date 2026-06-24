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

package browser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/agent"
)

type (
	downloadPDFParams struct {
		URL string `json:"url" jsonschema:"The URL of the PDF document to download and extract text from"`
	}

	downloadPDFResult struct {
		Text        string `json:"text"`
		PageCount   int    `json:"page_count"`
		ErrorDetail string `json:"error_detail,omitempty"`
	}
)

func DownloadPDFTool() agent.Tool {
	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = 30 * time.Second

	return agent.FunctionTool(
		"download_pdf",
		"Download a PDF document from a URL and extract its text content. Use this for DPAs, SOC 2 reports, privacy policies, and other documents hosted as PDFs.",
		func(ctx context.Context, p downloadPDFParams) (agent.ToolResult, error) {
			if err := validatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
					},
				), nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
					},
				), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot download PDF: %s", err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("PDF download returned status %d", resp.StatusCode),
					},
				), nil
			}

			// Read PDF into memory (max 20MB).
			body, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
			if err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot read PDF body: %s", err),
					},
				), nil
			}

			// Write to temp file for pdfcpu.
			tmpDir, err := os.MkdirTemp("", "pdf-extract-*")
			if err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot create temp dir: %s", err),
					},
				), nil
			}

			defer func() { _ = os.RemoveAll(tmpDir) }()

			tmpFile := filepath.Join(tmpDir, "input.pdf")
			if err := os.WriteFile(tmpFile, body, 0o600); err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot write temp file: %s", err),
					},
				), nil
			}

			// Get page count.
			conf := model.NewDefaultConfiguration()

			pageCount, err := api.PageCountFile(tmpFile)
			if err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot read PDF: %s", err),
					},
				), nil
			}

			// Extract content, digesting each page's content stream.
			var sb strings.Builder

			reader := bytes.NewReader(body)
			digest := func(r io.Reader, _ int) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return err
				}

				sb.Write(content)
				sb.WriteString("\n")

				return nil
			}

			if err := api.ExtractContent(reader, nil, digest, conf); err != nil {
				return agent.ResultJSON(
					downloadPDFResult{
						ErrorDetail: fmt.Sprintf("cannot extract PDF content: %s", err),
					},
				), nil
			}

			text := sb.String()
			if len(text) > maxTextLength {
				text = text[:maxTextLength] + "\n[... truncated]"
			}

			return agent.ResultJSON(
				downloadPDFResult{
					Text:      text,
					PageCount: pageCount,
				},
			), nil
		},
	)
}
