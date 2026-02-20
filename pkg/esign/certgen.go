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

package esign

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"html/template"
	"time"

	"github.com/digitorus/timestamp"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/html2pdf"
)

type (
	CertificateGenerator struct {
		HTML2PDFConverter *html2pdf.Converter
	}

	certificateData struct {
		SignatureID      string
		OrganizationID   string
		SignerFullName   string
		SignerEmail      string
		SignerIPAddress  string
		SignerUserAgent  string
		DocumentType     string
		DocumentTypeName string
		FileID           string
		FileHash         string
		Seal             string
		SealVersion      int
		ConsentText      string
		SignedAt         string
		TSAAuthority     string
		TSATime          string
		TSASerial        string
		Events           []certificateEvent
	}

	certificateEvent struct {
		EventType  string
		Source     string
		Actor      string
		IPAddress  string
		OccurredAt string
	}
)

var (
	//go:embed certificate.html.tmpl
	certificateTemplateHTML string

	certificateTemplate = template.Must(template.New("certificate").Parse(certificateTemplateHTML))
)

func (g *CertificateGenerator) Generate(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	events coredata.ElectronicSignatureEvents,
) (io.Reader, error) {
	data := certificateData{
		SignatureID:      signature.ID.String(),
		OrganizationID:   signature.OrganizationID.String(),
		SignerFullName:   ref.UnrefOrZero(signature.SignerFullName),
		SignerEmail:      signature.SignerEmail,
		SignerIPAddress:  ref.UnrefOrZero(signature.SignerIPAddress),
		SignerUserAgent:  ref.UnrefOrZero(signature.SignerUserAgent),
		DocumentType:     signature.DocumentType.String(),
		DocumentTypeName: signature.DocumentType.DisplayName(),
		FileID:           signature.FileID.String(),
		FileHash:         ref.UnrefOrZero(signature.FileHash),
		Seal:             ref.UnrefOrZero(signature.Seal),
		SealVersion:      signature.SealVersion,
		ConsentText:      signature.ConsentText,
	}

	if signature.SignedAt == nil {
		return nil, fmt.Errorf("cannot generate certificate: signature %s has no signed_at timestamp", signature.ID)
	}
	data.SignedAt = signature.SignedAt.UTC().Format(time.RFC3339)

	if len(signature.TSAToken) == 0 {
		return nil, fmt.Errorf("cannot generate certificate: signature %s has no TSA token", signature.ID)
	}

	tsResp, err := timestamp.ParseResponse(signature.TSAToken)
	if err != nil {
		return nil, fmt.Errorf("cannot parse TSA token for signature %s: %w", signature.ID, err)
	}

	data.TSATime = tsResp.Time.UTC().Format(time.RFC3339)
	data.TSASerial = tsResp.SerialNumber.String()
	data.TSAAuthority = tsaAuthorityName(tsResp)

	for _, evt := range events {
		data.Events = append(
			data.Events,
			certificateEvent{
				EventType:  evt.EventType.String(),
				Source:     evt.EventSource.String(),
				Actor:      evt.ActorEmail,
				IPAddress:  evt.ActorIPAddress,
				OccurredAt: evt.OccurredAt.UTC().Format(time.RFC3339),
			},
		)
	}

	var htmlBuf bytes.Buffer
	if err := certificateTemplate.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("cannot render certificate template: %w", err)
	}

	pdfReader, err := g.HTML2PDFConverter.GeneratePDF(
		ctx,
		htmlBuf.Bytes(),
		html2pdf.RenderConfig{
			PageFormat:      html2pdf.PageFormatA4,
			Orientation:     html2pdf.OrientationPortrait,
			MarginTop:       html2pdf.NewMarginMillimeters(20),
			MarginBottom:    html2pdf.NewMarginMillimeters(20),
			MarginLeft:      html2pdf.NewMarginMillimeters(20),
			MarginRight:     html2pdf.NewMarginMillimeters(20),
			PrintBackground: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot generate certificate PDF: %w", err)
	}

	return pdfReader, nil
}

func tsaAuthorityName(ts *timestamp.Timestamp) string {
	if len(ts.Certificates) == 0 {
		return ""
	}

	cert := ts.Certificates[0]

	if len(cert.Subject.Organization) > 0 {
		org := strings.Join(cert.Subject.Organization, ", ")
		if cert.Subject.CommonName != "" {
			return org + " (" + cert.Subject.CommonName + ")"
		}
		return org
	}

	return cert.Subject.CommonName
}
