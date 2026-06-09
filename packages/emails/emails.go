// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"context"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"net/url"
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
)

func DefaultPresenterConfig(baseURL string) PresenterConfig {
	return PresenterConfig{
		APIBaseURL:                      baseURL, // always API base URL
		BaseURL:                         baseURL, // can change to custom domain when needed
		PoweredByLogoPath:               brand.DefaultPoweredByLogoPath,
		SenderCompanyName:               "Probo",
		SenderCompanyWebsiteURL:         "https://www.getprobo.com",
		SenderCompanyLogoPath:           brand.DefaultSenderCompanyLogoPath,
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
	subjectConfirmEmail                      = "Confirm your email address"
	subjectPasswordReset                     = "Reset your password"
	subjectInvitation                        = "Invitation to join %s on Probo"
	subjectDocumentApproval                  = "Action Required – Please review and approve %s"
	subjectDocumentSigning                   = "Action Required – Please review and sign %s compliance documents"
	subjectDocumentExport                    = "Your document export is ready"
	subjectFrameworkExport                   = "Your framework export is ready"
	subjectTrustCenterAccess                 = "Compliance Page Access Invitation - %s"
	subjectTrustCenterDocumentAccessRejected = "Compliance Page Document Access Rejected - %s"
	subjectMagicLink                         = "Connect to %s"
	subjectMailingListSubscription           = "%s – Confirm Your Compliance Updates Subscription"
	subjectMailingListUnsubscription         = "%s – You've been unsubscribed"
	subjectMailingListUpdates                = "%s – %s"
)

var (
	confirmEmailHTMLTemplate                      = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/confirm-email.html.tmpl"))
	confirmEmailTextTemplate                      = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/confirm-email.txt.tmpl"))
	passwordResetHTMLTemplate                     = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/password-reset.html.tmpl"))
	passwordResetTextTemplate                     = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/password-reset.txt.tmpl"))
	invitationHTMLTemplate                        = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/invitation.html.tmpl"))
	invitationTextTemplate                        = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/invitation.txt.tmpl"))
	documentApprovalHTMLTemplate                  = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/document-approval.html.tmpl"))
	documentApprovalTextTemplate                  = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/document-approval.txt.tmpl"))
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
	electronicSignatureCertificateHTMLTemplate    = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/electronic-signature-certificate.html.tmpl"))
	electronicSignatureCertificateTextTemplate    = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/electronic-signature-certificate.txt.tmpl"))
	mailingListSubscriptionHTMLTemplate           = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-subscription.html.tmpl"))
	mailingListSubscriptionTextTemplate           = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-subscription.txt.tmpl"))
	mailingListUnsubscriptionHTMLTemplate         = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-unsubscription.html.tmpl"))
	mailingListUnsubscriptionTextTemplate         = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-unsubscription.txt.tmpl"))
	mailingListUpdatesHTMLTemplate                = htmltemplate.Must(htmltemplate.ParseFS(Templates, "dist/mailing-list-updates.html.tmpl"))
	mailingListUpdatesTextTemplate                = texttemplate.Must(texttemplate.ParseFS(Templates, "dist/mailing-list-updates.txt.tmpl"))
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
	approvalURLPath string,
	approvalURLQuery url.Values,
	organizationName string,
	documentName string,
) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	approvalURL := baseurl.MustParse(vars.BaseURL).
		AppendPath(approvalURLPath).
		WithQueryValues(approvalURLQuery).
		MustString()

	data := struct {
		*CommonVariables
		ApprovalUrl      string
		OrganizationName string
		DocumentName     string
	}{
		CommonVariables:  vars,
		ApprovalUrl:      approvalURL,
		OrganizationName: organizationName,
		DocumentName:     documentName,
	}

	textBody, htmlBody, err = renderEmail(documentApprovalTextTemplate, documentApprovalHTMLTemplate, data)

	return fmt.Sprintf(subjectDocumentApproval, documentName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentSigning(
	ctx context.Context,
	signingURLPath string,
	signingURLQuery url.Values,
	organizationName string,
) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables()
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	signingURL := baseurl.MustParse(vars.BaseURL).
		AppendPath(signingURLPath).
		WithQueryValues(signingURLQuery).
		MustString()

	data := struct {
		*CommonVariables
		SigningUrl       string
		OrganizationName string
	}{
		CommonVariables:  vars,
		SigningUrl:       signingURL,
		OrganizationName: organizationName,
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

func (p *Presenter) RenderTrustCenterAccess(ctx context.Context, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
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

	textBody, htmlBody, err = renderEmail(trustCenterAccessTextTemplate, trustCenterAccessHTMLTemplate, data)

	return fmt.Sprintf(subjectTrustCenterAccess, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderTrustCenterDocumentAccessRejected(
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

	textBody, htmlBody, err = renderEmail(trustCenterDocumentAccessRejectedTextTemplate, trustCenterDocumentAccessRejectedHTMLTemplate, data)

	return fmt.Sprintf(subjectTrustCenterDocumentAccessRejected, organizationName), textBody, htmlBody, err
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
