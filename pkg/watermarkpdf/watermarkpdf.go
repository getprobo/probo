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

package watermarkpdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func AddConfidentialWithTimestamp(pdfData []byte, email string) ([]byte, error) {
	reader := bytes.NewReader(pdfData)
	var buf bytes.Buffer

	// Replace email with invisible characters to prevent auto-linking
	formattedEmail := strings.ReplaceAll(email, "@", "\u200B@\u200B")
	formattedEmail = strings.ReplaceAll(formattedEmail, ".", "\u200B.\u200B")

	watermarkText := strings.Join([]string{
		"Confidential",
		formattedEmail,
		time.Now().Format("02/01/2006"),
	}, "\n")

	watermarkConf := model.DefaultWatermarkConfig()
	watermarkConf.Mode = model.WMText
	watermarkConf.TextString = watermarkText
	watermarkConf.FontName = "Helvetica"
	watermarkConf.FontSize = 120
	watermarkConf.Rotation = 55
	watermarkConf.Opacity = 0.20
	watermarkConf.OnTop = true
	watermarkConf.ScaleAbs = true
	watermarkConf.Update = false

	err := api.AddWatermarks(reader, &buf, nil, watermarkConf, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to add watermark: %w", err)
	}

	return buf.Bytes(), nil
}
