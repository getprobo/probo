// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package docgen

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

var (
	//go:embed template.html
	htmlTemplateContent string

	templateFuncs = template.FuncMap{
		"now":                  func() time.Time { return time.Now() },
		"eq":                   func(a, b string) bool { return a == b },
		"string":               func(v fmt.Stringer) string { return v.String() },
		"lower":                func(s string) string { return strings.ToLower(s) },
		"classificationString": func(c Classification) string { return string(c) },
		"formatContent": func(content string) template.HTML {
			md := goldmark.New(
				goldmark.WithExtensions(extension.Table),
				goldmark.WithRendererOptions(
					gmhtml.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(content), &buf); err != nil {
				return template.HTML(fmt.Sprintf("<p>%s</p>", html.EscapeString(content)))
			}
			return template.HTML(buf.String())
		},
	}

	documentTemplate = template.Must(template.New("document").Funcs(templateFuncs).Parse(htmlTemplateContent))
)

type (
	Classification string

	DocumentData struct {
		Title          string
		Content        string
		Version        int
		Classification Classification
		Approver       string
		Description    string
		PublishedAt    *time.Time
		Signatures     []SignatureData
	}

	SignatureData struct {
		SignedBy    string
		SignedAt    *time.Time
		State       coredata.DocumentVersionSignatureState
		RequestedAt time.Time
		RequestedBy string
	}
)

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
)

func RenderHTML(data DocumentData) ([]byte, error) {
	data.Title = html.EscapeString(data.Title)
	data.Approver = html.EscapeString(data.Approver)
	data.Description = html.EscapeString(data.Description)

	for i := range data.Signatures {
		data.Signatures[i].SignedBy = html.EscapeString(data.Signatures[i].SignedBy)
		data.Signatures[i].RequestedBy = html.EscapeString(data.Signatures[i].RequestedBy)
	}

	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
