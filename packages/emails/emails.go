// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package emails

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
	"time"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/brand"
)

//go:embed dist
var Templates embed.FS

type (
	PresenterConfig struct {
		APIBaseURL                      string
		BaseURL                         string
		PoweredByLogoPath               string
		SenderCompanyName               string
		SenderCompanyWebsiteURL         string
		SenderCompanyLogoPath           string
		SenderCompanyHeadquarterAddress string
	}

	CommonVariables struct {
		// Static variables
		APIBaseURL                      string
		BaseURL                         string
		PoweredByLogoURL                string
		SenderCompanyName               string
		SenderCompanyWebsiteURL         string
		SenderCompanyLogoURL            string
		SenderCompanyHeadquarterAddress string

		// Other common variables
		RecipientFullName string
	}

	Presenter struct {
		config            PresenterConfig
		RecipientFullName string
	}

	DocumentSummary struct {
		Title string
		Type  string
		URL   string
	}
)

func DefaultPresenterConfig(baseURL string) PresenterConfig {
	return PresenterConfig{
		APIBaseURL:                      baseURL, // always API base URL
		BaseURL:                         baseURL, // can change to custom domain when needed
		PoweredByLogoPath:               brand.StaticPath(brand.PoweredByLogo),
		SenderCompanyName:               "Probo",
		SenderCompanyWebsiteURL:         "https://www.probo.com",
		SenderCompanyLogoPath:           brand.StaticPath(brand.SenderCompanyLogo),
		SenderCompanyHeadquarterAddress: "Probo Inc, 490 Post St, STE 640, San Francisco, CA, 94102, US",
	}
}

func NewPresenterFromConfig(cfg PresenterConfig, fullName string) *Presenter {
	return &Presenter{
		config:            cfg,
		RecipientFullName: fullName,
	}
}

func NewPresenter(baseURL string, fullName string) *Presenter {
	return NewPresenterFromConfig(
		DefaultPresenterConfig(baseURL),
		fullName,
	)
}

const (
	subjectConfirmEmail                           = "Confirm your email address"
	subjectPasswordReset                          = "Reset your password"
	subjectInvitation                             = "Invitation to join %s on Probo"
	subjectDocumentApproval                       = "Action Required – Please review and approve %s compliance documents"
	subjectDocumentSigning                        = "Action Required – Please review and sign %s compliance documents"
	subjectDocumentExport                         = "Your document export is ready"
	subjectFrameworkExport                        = "Your framework export is ready"
	subjectCompliancePortalAccess                 = "Compliance Page Access Invitation - %s"
	subjectCompliancePortalDocumentAccessRejected = "Compliance Page Document Access Rejected - %s"
	subjectMagicLink                              = "Connect to %s"
	subjectMailingListSubscription                = "%s – Confirm Your Compliance Updates Subscription"
	subjectMailingListUnsubscription              = "%s – You've been unsubscribed"
	subjectMailingListUpdates                     = "%s – %s"
)

var (
	confirmEmailHTMLTemplate                           = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/confirm-email.html.tmpl"))
	confirmEmailTextTemplate                           = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/confirm-email.txt.tmpl"))
	passwordResetHTMLTemplate                          = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/password-reset.html.tmpl"))
	passwordResetTextTemplate                          = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/password-reset.txt.tmpl"))
	invitationHTMLTemplate                             = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/invitation.html.tmpl"))
	invitationTextTemplate                             = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/invitation.txt.tmpl"))
	documentApprovalHTMLTemplate                       = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-approval.html.tmpl"))
	documentApprovalTextTemplate                       = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-approval.txt.tmpl"))
	documentSigningHTMLTemplate                        = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-signing.html.tmpl"))
	documentSigningTextTemplate                        = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-signing.txt.tmpl"))
	documentExportHTMLTemplate                         = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-export.html.tmpl"))
	documentExportTextTemplate                         = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-export.txt.tmpl"))
	frameworkExportHTMLTemplate                        = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/framework-export.html.tmpl"))
	frameworkExportTextTemplate                        = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/framework-export.txt.tmpl"))
	compliancePortalAccessHTMLTemplate                 = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/compliance-portal-access.html.tmpl"))
	compliancePortalAccessTextTemplate                 = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/compliance-portal-access.txt.tmpl"))
	compliancePortalDocumentAccessRejectedHTMLTemplate = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/compliance-portal-document-access-rejected.html.tmpl"))
	compliancePortalDocumentAccessRejectedTextTemplate = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/compliance-portal-document-access-rejected.txt.tmpl"))
	magicLinkHTMLTemplate                              = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/magic-link.html.tmpl"))
	magicLinkTextTemplate                              = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/magic-link.txt.tmpl"))
	electronicSignatureCertificateHTMLTemplate         = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/electronic-signature-certificate.html.tmpl"))
	electronicSignatureCertificateTextTemplate         = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/electronic-signature-certificate.txt.tmpl"))
	mailingListSubscriptionHTMLTemplate                = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-subscription.html.tmpl"))
	mailingListSubscriptionTextTemplate                = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-subscription.txt.tmpl"))
	mailingListUnsubscriptionHTMLTemplate              = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-unsubscription.html.tmpl"))
	mailingListUnsubscriptionTextTemplate              = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-unsubscription.txt.tmpl"))
	mailingListUpdatesHTMLTemplate                     = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-updates.html.tmpl"))
	mailingListUpdatesTextTemplate                     = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-updates.txt.tmpl"))
)

func (p *Presenter) getCommonVariables() (*CommonVariables, error) {
	apiBaseURL := baseurl.MustParse(p.config.APIBaseURL)
	poweredByLogoURL := apiBaseURL.AppendPath(p.config.PoweredByLogoPath).MustString()
	senderCompanyLogoURL := apiBaseURL.AppendPath(p.config.SenderCompanyLogoPath).MustString()

	return &CommonVariables{
		APIBaseURL:                      p.config.APIBaseURL,
		BaseURL:                         p.config.BaseURL,
		PoweredByLogoURL:                poweredByLogoURL,
		SenderCompanyName:               p.config.SenderCompanyName,
		SenderCompanyWebsiteURL:         p.config.SenderCompanyWebsiteURL,
		SenderCompanyLogoURL:            senderCompanyLogoURL,
		SenderCompanyHeadquarterAddress: p.config.SenderCompanyHeadquarterAddress,
		RecipientFullName:               p.RecipientFullName,
	}, nil
}

func (p *Presenter) RenderConfirmEmail(ctx context.Context, confirmationURLPath string, confirmationTokenParam string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	confirmationUrl := baseurl.
		MustParse(vars.BaseURL).
		AppendPath(confirmationURLPath).
		WithQuery("token", confirmationTokenParam).
		MustString()

	data := struct {
		*CommonVariables
		ConfirmationUrl string
	}{
		CommonVariables: vars,
		ConfirmationUrl: confirmationUrl,
	}

	textBody, htmlBody, err = renderEmail(confirmEmailTextTemplate, confirmEmailHTMLTemplate, data)

	return subjectConfirmEmail, textBody, htmlBody, err
}

func (p *Presenter) RenderPasswordReset(ctx context.Context, resetPasswordURLPath string, resetPasswordToken string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	resetUrl := baseurl.
		MustParse(vars.BaseURL).
		AppendPath(resetPasswordURLPath).
		WithQuery("token", resetPasswordToken).
		MustString()

	data := struct {
		*CommonVariables
		ResetUrl string
	}{
		CommonVariables: vars,
		ResetUrl:        resetUrl,
	}

	textBody, htmlBody, err = renderEmail(passwordResetTextTemplate, passwordResetHTMLTemplate, data)

	return subjectPasswordReset, textBody, htmlBody, err
}

func (p *Presenter) RenderInvitation(ctx context.Context, invitationURLPath string, invitationToken string, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	invitationURL := baseurl.
		MustParse(vars.BaseURL).
		AppendPath(invitationURLPath).
		WithQuery("token", invitationToken).
		MustString()

	data := struct {
		*CommonVariables
		InvitationUrl    string
		OrganizationName string
	}{
		CommonVariables:  vars,
		InvitationUrl:    invitationURL,
		OrganizationName: organizationName,
	}

	textBody, htmlBody, err = renderEmail(invitationTextTemplate, invitationHTMLTemplate, data)

	return fmt.Sprintf(subjectInvitation, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentApproval(
	ctx context.Context,
	approvalURL string,
	organizationName string,
	documents []DocumentSummary,
) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		ApprovalUrl      string
		OrganizationName string
		Documents        []DocumentSummary
	}{
		CommonVariables:  vars,
		ApprovalUrl:      approvalURL,
		OrganizationName: organizationName,
		Documents:        documents,
	}

	textBody, htmlBody, err = renderEmail(documentApprovalTextTemplate, documentApprovalHTMLTemplate, data)

	return fmt.Sprintf(subjectDocumentApproval, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentSigning(
	ctx context.Context,
	signingURL string,
	organizationName string,
	documents []DocumentSummary,
) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		SigningUrl       string
		OrganizationName string
		Documents        []DocumentSummary
	}{
		CommonVariables:  vars,
		SigningUrl:       signingURL,
		OrganizationName: organizationName,
		Documents:        documents,
	}

	textBody, htmlBody, err = renderEmail(documentSigningTextTemplate, documentSigningHTMLTemplate, data)

	return fmt.Sprintf(subjectDocumentSigning, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentExport(ctx context.Context, downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		DownloadUrl string
	}{
		CommonVariables: vars,
		DownloadUrl:     downloadUrl,
	}

	textBody, htmlBody, err = renderEmail(documentExportTextTemplate, documentExportHTMLTemplate, data)

	return subjectDocumentExport, textBody, htmlBody, err
}

func (p *Presenter) RenderFrameworkExport(ctx context.Context, downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		DownloadUrl string
	}{
		CommonVariables: vars,
		DownloadUrl:     downloadUrl,
	}

	textBody, htmlBody, err = renderEmail(frameworkExportTextTemplate, frameworkExportHTMLTemplate, data)

	return subjectFrameworkExport, textBody, htmlBody, err
}

func (p *Presenter) RenderCompliancePortalAccess(ctx context.Context, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		OrganizationName string
	}{
		CommonVariables:  vars,
		OrganizationName: organizationName,
	}

	textBody, htmlBody, err = renderEmail(compliancePortalAccessTextTemplate, compliancePortalAccessHTMLTemplate, data)

	return fmt.Sprintf(subjectCompliancePortalAccess, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderCompliancePortalDocumentAccessRejected(
	ctx context.Context,
	fileNames []string,
	organizationName string,
) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		FileNames        []string
		OrganizationName string
	}{
		CommonVariables:  vars,
		FileNames:        fileNames,
		OrganizationName: organizationName,
	}

	textBody, htmlBody, err = renderEmail(compliancePortalDocumentAccessRejectedTextTemplate, compliancePortalDocumentAccessRejectedHTMLTemplate, data)

	return fmt.Sprintf(subjectCompliancePortalDocumentAccessRejected, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderMagicLink(ctx context.Context, magicLinkUrlPath string, tokenString string, tokenDuration time.Duration, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		MagicLinkURL      string
		DurationInMinutes int
		OrganizationName  string
	}{
		CommonVariables:   vars,
		MagicLinkURL:      baseurl.MustParse(vars.BaseURL).AppendPath(magicLinkUrlPath).WithQuery("token", tokenString).MustString(),
		DurationInMinutes: int(tokenDuration.Minutes()),
		OrganizationName:  organizationName,
	}

	textBody, htmlBody, err = renderEmail(magicLinkTextTemplate, magicLinkHTMLTemplate, data)

	return fmt.Sprintf(subjectMagicLink, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderElectronicSignatureCertificate(ctx context.Context, signerName string, documentName string, subject string) (textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		SignerName   string
		DocumentName string
		Subject      string
	}{
		CommonVariables: vars,
		SignerName:      signerName,
		DocumentName:    documentName,
		Subject:         subject,
	}

	return renderEmail(electronicSignatureCertificateTextTemplate, electronicSignatureCertificateHTMLTemplate, data)
}

func (p *Presenter) RenderMailingListSubscription(ctx context.Context, organizationName string, confirmURL string, unsubscribeURL string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		OrganizationName string
		ConfirmURL       string
		UnsubscribeURL   string
	}{
		CommonVariables:  vars,
		OrganizationName: organizationName,
		ConfirmURL:       confirmURL,
		UnsubscribeURL:   unsubscribeURL,
	}

	textBody, htmlBody, err = renderEmail(mailingListSubscriptionTextTemplate, mailingListSubscriptionHTMLTemplate, data)
	if err != nil {
		return "", "", nil, err
	}

	return fmt.Sprintf(subjectMailingListSubscription, organizationName), textBody, htmlBody, nil
}

func (p *Presenter) RenderMailingListUnsubscription(ctx context.Context, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		OrganizationName string
	}{
		CommonVariables:  vars,
		OrganizationName: organizationName,
	}

	textBody, htmlBody, err = renderEmail(mailingListUnsubscriptionTextTemplate, mailingListUnsubscriptionHTMLTemplate, data)
	if err != nil {
		return "", "", nil, err
	}

	return fmt.Sprintf(subjectMailingListUnsubscription, organizationName), textBody, htmlBody, nil
}

func (p *Presenter) RenderMailingListNews(ctx context.Context, organizationName string, newsTitle string, newsBody string, compliancePageURL string, unsubscribeURL string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	data := struct {
		*CommonVariables
		OrganizationName  string
		NewsTitle         string
		NewsBody          string
		CompliancePageURL string
		UnsubscribeURL    string
	}{
		CommonVariables:   vars,
		OrganizationName:  organizationName,
		NewsTitle:         newsTitle,
		NewsBody:          newsBody,
		CompliancePageURL: compliancePageURL,
		UnsubscribeURL:    unsubscribeURL,
	}

	textBody, htmlBody, err = renderEmail(mailingListUpdatesTextTemplate, mailingListUpdatesHTMLTemplate, data)
	if err != nil {
		return "", "", nil, err
	}

	return fmt.Sprintf(subjectMailingListUpdates, organizationName, newsTitle), textBody, htmlBody, nil
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
