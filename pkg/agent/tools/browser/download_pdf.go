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
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
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
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: netcheck.NewPinnedTransport(),
	}

	return agent.FunctionTool(
		"download_pdf",
		"Download a PDF document from a URL and extract its text content. Use this for DPAs, SOC 2 reports, privacy policies, and other documents hosted as PDFs.",
		func(ctx context.Context, p downloadPDFParams) (agent.ToolResult, error) {
			if err := validatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
				}), nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
				}), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot download PDF: %s", err),
				}), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("PDF download returned status %d", resp.StatusCode),
				}), nil
			}

			// Read PDF into memory (max 20MB).
			body, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
			if err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot read PDF body: %s", err),
				}), nil
			}

			// Write to temp file for pdfcpu.
			tmpDir, err := os.MkdirTemp("", "pdf-extract-*")
			if err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create temp dir: %s", err),
				}), nil
			}
			defer os.RemoveAll(tmpDir)

			tmpFile := filepath.Join(tmpDir, "input.pdf")
			if err := os.WriteFile(tmpFile, body, 0o600); err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot write temp file: %s", err),
				}), nil
			}

			// Get page count.
			conf := model.NewDefaultConfiguration()
			pageCount, err := api.PageCountFile(tmpFile)
			if err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot read PDF: %s", err),
				}), nil
			}

			// Extract content to output dir.
			outDir := filepath.Join(tmpDir, "out")
			if err := os.MkdirAll(outDir, 0o700); err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create output dir: %s", err),
				}), nil
			}

			reader := bytes.NewReader(body)
			if err := api.ExtractContent(reader, outDir, "content", nil, conf); err != nil {
				return agent.ResultJSON(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot extract PDF content: %s", err),
				}), nil
			}

			// Read all extracted content files.
			var sb strings.Builder
			entries, _ := os.ReadDir(outDir)
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				content, err := os.ReadFile(filepath.Join(outDir, entry.Name()))
				if err != nil {
					continue
				}
				sb.Write(content)
				sb.WriteString("\n")
			}

			text := sb.String()
			if len(text) > maxTextLength {
				text = text[:maxTextLength] + "\n[... truncated]"
			}

			return agent.ResultJSON(downloadPDFResult{
				Text:      text,
				PageCount: pageCount,
			}), nil
		},
	)
}
