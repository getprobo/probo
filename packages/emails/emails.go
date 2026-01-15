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

package emails

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"math"
	texttemplate "text/template"
	"time"
)

//go:embed dist
var Templates embed.FS

const (
	logoURLPath = "/logos/probo.png"

	subjectConfirmEmail                      = "Confirm your email address"
	subjectPasswordReset                     = "Reset your password"
	subjectInvitation                        = "Invitation to join %s on Probo"
	subjectDocumentSigning                   = "Action Required – Please review and sign %s compliance documents"
	subjectDocumentExport                    = "Your document export is ready"
	subjectFrameworkExport                   = "Your framework export is ready"
	subjectTrustCenterAccess                 = "Trust Center Access Invitation - %s"
	subjectTrustCenterDocumentAccessRejected = "Trust Center Document Access Rejected - %s"
	subjectMagicLink                         = "Connect to Probo"
)

var (
	confirmEmailHTMLTemplate                      = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/confirm-email.html.tmpl"))
	confirmEmailTextTemplate                      = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/confirm-email.txt.tmpl"))
	passwordResetHTMLTemplate                     = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/password-reset.html.tmpl"))
	passwordResetTextTemplate                     = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/password-reset.txt.tmpl"))
	invitationHTMLTemplate                        = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/invitation.html.tmpl"))
	invitationTextTemplate                        = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/invitation.txt.tmpl"))
	documentSigningHTMLTemplate                   = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-signing.html.tmpl"))
	documentSigningTextTemplate                   = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-signing.txt.tmpl"))
	documentExportHTMLTemplate                    = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-export.html.tmpl"))
	documentExportTextTemplate                    = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-export.txt.tmpl"))
	frameworkExportHTMLTemplate                   = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/framework-export.html.tmpl"))
	frameworkExportTextTemplate                   = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/framework-export.txt.tmpl"))
	trustCenterAccessHTMLTemplate                 = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/trust-center-access.html.tmpl"))
	trustCenterAccessTextTemplate                 = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/trust-center-access.txt.tmpl"))
	trustCenterDocumentAccessRejectedHTMLTemplate = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/trust-center-document-access-rejected.html.tmpl"))
	trustCenterDocumentAccessRejectedTextTemplate = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/trust-center-document-access-rejected.txt.tmpl"))
	magicLinkHTMLTemplate                         = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/magic-link.html.tmpl"))
	magicLinkTextTemplate                         = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/magic-link.txt.tmpl"))
)

func RenderConfirmEmail(baseURL, fullName, confirmationUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName        string
		ConfirmationUrl string
		LogoURL         string
	}{
		FullName:        fullName,
		ConfirmationUrl: confirmationUrl,
		LogoURL:         baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(confirmEmailTextTemplate, confirmEmailHTMLTemplate, data)
	return subjectConfirmEmail, textBody, htmlBody, err
}

func RenderPasswordReset(baseURL, fullName, resetUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName string
		ResetUrl string
		LogoURL  string
	}{
		FullName: fullName,
		ResetUrl: resetUrl,
		LogoURL:  baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(passwordResetTextTemplate, passwordResetHTMLTemplate, data)
	return subjectPasswordReset, textBody, htmlBody, err
}

func RenderInvitation(baseURL, fullName, organizationName, invitationUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName         string
		OrganizationName string
		InvitationUrl    string
		LogoURL          string
	}{
		FullName:         fullName,
		OrganizationName: organizationName,
		InvitationUrl:    invitationUrl,
		LogoURL:          baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(invitationTextTemplate, invitationHTMLTemplate, data)
	return fmt.Sprintf(subjectInvitation, organizationName), textBody, htmlBody, err
}

func RenderDocumentSigning(baseURL, fullName, organizationName, signingUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName         string
		OrganizationName string
		SigningUrl       string
		LogoURL          string
	}{
		FullName:         fullName,
		OrganizationName: organizationName,
		SigningUrl:       signingUrl,
		LogoURL:          baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(documentSigningTextTemplate, documentSigningHTMLTemplate, data)
	return fmt.Sprintf(subjectDocumentSigning, organizationName), textBody, htmlBody, err
}

func RenderDocumentExport(baseURL, fullName, downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName    string
		DownloadUrl string
		LogoURL     string
	}{
		FullName:    fullName,
		DownloadUrl: downloadUrl,
		LogoURL:     baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(documentExportTextTemplate, documentExportHTMLTemplate, data)
	return subjectDocumentExport, textBody, htmlBody, err
}

func RenderFrameworkExport(baseURL, fullName, downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName    string
		DownloadUrl string
		LogoURL     string
	}{
		FullName:    fullName,
		DownloadUrl: downloadUrl,
		LogoURL:     baseURL + logoURLPath,
	}

	textBody, htmlBody, err = renderEmail(frameworkExportTextTemplate, frameworkExportHTMLTemplate, data)
	return subjectFrameworkExport, textBody, htmlBody, err
}

func RenderTrustCenterAccess(baseURL, fullName, organizationName, accessUrl string, tokenDuration time.Duration) (subject string, textBody string, htmlBody *string, err error) {
	durationInDays := int(math.Round(tokenDuration.Hours() / 24))

	data := struct {
		FullName         string
		OrganizationName string
		AccessUrl        string
		LogoURL          string
		DurationInDays   int
	}{
		FullName:         fullName,
		OrganizationName: organizationName,
		AccessUrl:        accessUrl,
		LogoURL:          baseURL + logoURLPath,
		DurationInDays:   durationInDays,
	}

	textBody, htmlBody, err = renderEmail(trustCenterAccessTextTemplate, trustCenterAccessHTMLTemplate, data)
	return fmt.Sprintf(subjectTrustCenterAccess, organizationName), textBody, htmlBody, err
}

func RenderTrustCenterDocumentAccessRejected(
	baseURL string,
	fullName string,
	organizationName string,
	fileNames []string,
) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName         string
		OrganizationName string
		LogoURL          string
		FileNames        []string
	}{
		FullName:         fullName,
		OrganizationName: organizationName,
		LogoURL:          baseURL + logoURLPath,
		FileNames:        fileNames,
	}

	textBody, htmlBody, err = renderEmail(trustCenterDocumentAccessRejectedTextTemplate, trustCenterDocumentAccessRejectedHTMLTemplate, data)
	return fmt.Sprintf(subjectTrustCenterDocumentAccessRejected, organizationName), textBody, htmlBody, err
}

func RenderMagicLink(baseURL, fullName, magicLinkUrl string, tokenDuration time.Duration) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		FullName          string
		MagicLinkURL      string
		LogoURL           string
		DurationInMinutes int
	}{
		FullName:          fullName,
		MagicLinkURL:      magicLinkUrl,
		LogoURL:           baseURL + logoURLPath,
		DurationInMinutes: int(tokenDuration.Minutes()),
	}

	textBody, htmlBody, err = renderEmail(magicLinkTextTemplate, magicLinkHTMLTemplate, data)
	return subjectMagicLink, textBody, htmlBody, err
}

func renderEmail(textTemplate *texttemplate.Template, htmlTemplate *htmltemplate.Template, data any) (textBody string, htmlBody *string, err error) {
	var textBuf bytes.Buffer
	if err := textTemplate.Execute(&textBuf, data); err != nil {
		return "", nil, fmt.Errorf("cannot execute text template: %w", err)
	}
	textBody = textBuf.String()

	var htmlBuf bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuf, data); err != nil {
		return "", nil, fmt.Errorf("cannot execute html template: %w", err)
	}
	htmlBodyStr := htmlBuf.String()
	htmlBody = &htmlBodyStr

	return textBody, htmlBody, nil
}
