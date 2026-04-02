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
	"encoding/json"
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
)

type downloadPDFParams struct {
	URL string `json:"url" jsonschema:"The URL of the PDF document to download and extract text from"`
}

type downloadPDFResult struct {
	Text        string `json:"text"`
	PageCount   int    `json:"page_count"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

func DownloadPDFTool() (agent.Tool, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	return agent.FunctionTool[downloadPDFParams](
		"download_pdf",
		"Download a PDF document from a URL and extract its text content. Use this for DPAs, SOC 2 reports, privacy policies, and other documents hosted as PDFs.",
		func(ctx context.Context, p downloadPDFParams) (agent.ToolResult, error) {
			if err := validatePublicURL(p.URL); err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
				})
				return agent.ToolResult{Content: string(data), IsError: true}, nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			resp, err := client.Do(req)
			if err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot download PDF: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("PDF download returned status %d", resp.StatusCode),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			// Read PDF into memory (max 20MB).
			body, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
			if err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot read PDF body: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			// Write to temp file for pdfcpu.
			tmpDir, err := os.MkdirTemp("", "pdf-extract-*")
			if err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create temp dir: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer os.RemoveAll(tmpDir)

			tmpFile := filepath.Join(tmpDir, "input.pdf")
			if err := os.WriteFile(tmpFile, body, 0o600); err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot write temp file: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			// Get page count.
			conf := model.NewDefaultConfiguration()
			pageCount, err := api.PageCountFile(tmpFile)
			if err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot read PDF: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			// Extract content to output dir.
			outDir := filepath.Join(tmpDir, "out")
			if err := os.MkdirAll(outDir, 0o700); err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot create output dir: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			reader := bytes.NewReader(body)
			if err := api.ExtractContent(reader, outDir, "content", nil, conf); err != nil {
				data, _ := json.Marshal(downloadPDFResult{
					ErrorDetail: fmt.Sprintf("cannot extract PDF content: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
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

			result := downloadPDFResult{
				Text:      text,
				PageCount: pageCount,
			}

			data, _ := json.Marshal(result)
			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
