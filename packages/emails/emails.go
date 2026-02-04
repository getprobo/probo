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
	"context"
	"embed"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	"mime"
	"net/url"
	"path/filepath"
	texttemplate "text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/filevalidation"
)

//go:embed dist
var Templates embed.FS

var (
	//go:embed assets
	staticAssets embed.FS

	staticAssetsValidator = filevalidation.NewValidator(
		filevalidation.WithMaxFileSize(5*1024*1024),
		filevalidation.WithCategories(
			filevalidation.CategoryImage,
			filevalidation.CategoryVideo,
		),
	)
)

type StaticAssetURLs map[string]string

type (
	PresenterConfig struct {
		BaseURL                         string
		PoweredByLogoURL                string
		SenderCompanyName               string
		SenderCompanyWebsiteURL         string
		SenderCompanyLogoURL            string
		SenderCompanyHeadquarterAddress string
	}

	PresenterVariables struct {
		// Static variables
		BaseURL                         string
		PoweredByLogoURL                string
		SenderCompanyName               string
		SenderCompanyWebsiteURL         string
		SenderCompanyLogoURL            string
		SenderCompanyHeadquarterAddress string

		// Common variables
		RecipientFullName string
		// Not to confuse with the SenderCompanyName, which is the brand of the product being used
		OrganizationName string
	}

	Presenter struct {
		variables PresenterVariables
	}
)

func DefaultPresenterConfig(baseURL string, staticAssetURLs StaticAssetURLs) PresenterConfig {
	return PresenterConfig{
		BaseURL:                         baseURL,
		PoweredByLogoURL:                staticAssetURLs["probo-gray-small.png"],
		SenderCompanyName:               "Probo",
		SenderCompanyWebsiteURL:         "https://www.getprobo.com",
		SenderCompanyLogoURL:            staticAssetURLs["probo.png"],
		SenderCompanyHeadquarterAddress: "Probo Inc, 490 Post St, STE 640, San Francisco, CA, 94102, US",
	}
}

func NewPresenterFromConfig(cfg PresenterConfig, fullName string) *Presenter {
	return &Presenter{
		variables: PresenterVariables{
			BaseURL:                         cfg.BaseURL,
			PoweredByLogoURL:                cfg.PoweredByLogoURL,
			SenderCompanyName:               cfg.SenderCompanyName,
			SenderCompanyWebsiteURL:         cfg.SenderCompanyWebsiteURL,
			SenderCompanyLogoURL:            cfg.SenderCompanyLogoURL,
			SenderCompanyHeadquarterAddress: cfg.SenderCompanyHeadquarterAddress,

			RecipientFullName: fullName,
		},
	}
}

func NewPresenter(baseURL string, staticAssetURLs StaticAssetURLs, fullName string) *Presenter {
	return NewPresenterFromConfig(
		DefaultPresenterConfig(baseURL, staticAssetURLs),
		fullName,
	)
}

func GenerateStaticAssetURLs(ctx context.Context, s3Client *s3.Client, bucket string) (StaticAssetURLs, error) {
	assetURLs := make(map[string]string)

	subFS, err := fs.Sub(staticAssets, "assets")
	if err != nil {
		return nil, fmt.Errorf("cannot create subtree file system: %w", err)
	}

	err = fs.WalkDir(subFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}

			return fmt.Errorf("cannot get dir entry info: %w", err)
		}

		ext := filepath.Ext(info.Name())
		mimeType := mime.TypeByExtension(ext)

		if err := staticAssetsValidator.Validate(info.Name(), mimeType, info.Size()); err != nil {
			return fmt.Errorf("cannot validate file: %w", err)
		}

		file, err := subFS.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = s3Client.PutObject(
			ctx,
			&s3.PutObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(path),
				Body:   file,
				Metadata: map[string]string{
					"type": "static-email-asset",
				},
				ContentType: aws.String(mimeType),
			},
		)
		if err != nil {
			return fmt.Errorf("cannot upload file to S3: %w", err)
		}

		presignClient := s3.NewPresignClient(s3Client)

		encodedFilename := url.QueryEscape(info.Name())
		contentDisposition := fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s",
			encodedFilename, encodedFilename)

		presignedReq, err := presignClient.PresignGetObject(
			ctx,
			&s3.GetObjectInput{
				Bucket:                     aws.String(bucket),
				Key:                        aws.String(path),
				ResponseCacheControl:       aws.String("max-age=3600, public"),
				ResponseContentDisposition: &contentDisposition,
			},
			func(opts *s3.PresignOptions) {
				opts.Expires = 7 * 24 * time.Hour
			},
		)
		if err != nil {
			return fmt.Errorf("cannot presign GetObject request: %w", err)
		}

		assetURLs[path] = presignedReq.URL

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("cannot generate asset URLs: %w", err)
	}

	return assetURLs, nil
}

const (
	subjectConfirmEmail                      = "Confirm your email address"
	subjectPasswordReset                     = "Reset your password"
	subjectInvitation                        = "Invitation to join %s on Probo"
	subjectDocumentSigning                   = "Action Required – Please review and sign %s compliance documents"
	subjectDocumentExport                    = "Your document export is ready"
	subjectFrameworkExport                   = "Your framework export is ready"
	subjectTrustCenterAccess                 = "Compliance Page Access Invitation - %s"
	subjectTrustCenterDocumentAccessRejected = "Compliance Page Document Access Rejected - %s"
	subjectMagicLink                         = "Connect to %s"
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

func (p *Presenter) RenderConfirmEmail(confirmationURLPath string, confirmationTokenParam string) (subject string, textBody string, htmlBody *string, err error) {
	confirmationUrl := baseurl.
		MustParse(p.variables.BaseURL).
		AppendPath(confirmationURLPath).
		WithQuery("token", confirmationTokenParam).
		MustString()

	data := struct {
		PresenterVariables
		ConfirmationUrl string
	}{
		PresenterVariables: p.variables,
		ConfirmationUrl:    confirmationUrl,
	}

	textBody, htmlBody, err = renderEmail(confirmEmailTextTemplate, confirmEmailHTMLTemplate, data)
	return subjectConfirmEmail, textBody, htmlBody, err
}

func (p *Presenter) RenderPasswordReset(resetPasswordURLPath string, resetPasswordToken string) (subject string, textBody string, htmlBody *string, err error) {
	resetUrl := baseurl.
		MustParse(p.variables.BaseURL).
		AppendPath(resetPasswordURLPath).
		WithQuery("token", resetPasswordToken).
		MustString()

	data := struct {
		PresenterVariables
		ResetUrl string
	}{
		PresenterVariables: p.variables,
		ResetUrl:           resetUrl,
	}

	textBody, htmlBody, err = renderEmail(passwordResetTextTemplate, passwordResetHTMLTemplate, data)
	return subjectPasswordReset, textBody, htmlBody, err
}

func (p *Presenter) RenderInvitation(invitationURLPath string, invitationToken string, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	invitationURL := baseurl.
		MustParse(p.variables.BaseURL).
		AppendPath(invitationURLPath).
		WithQuery("token", invitationToken).
		WithQuery("fullName", p.variables.RecipientFullName).
		MustString()

	data := struct {
		PresenterVariables
		InvitationUrl    string
		OrganizationName string
	}{
		PresenterVariables: p.variables,
		InvitationUrl:      invitationURL,
		OrganizationName:   organizationName,
	}

	textBody, htmlBody, err = renderEmail(invitationTextTemplate, invitationHTMLTemplate, data)
	return fmt.Sprintf(subjectInvitation, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentSigning(signinURLPath string, token string, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	signingURL := baseurl.MustParse(p.variables.BaseURL).
		AppendPath(signinURLPath).
		WithQuery("token", token).
		MustString()

	data := struct {
		PresenterVariables
		SigningUrl       string
		OrganizationName string
	}{
		PresenterVariables: p.variables,
		SigningUrl:         signingURL,
		OrganizationName:   organizationName,
	}

	textBody, htmlBody, err = renderEmail(documentSigningTextTemplate, documentSigningHTMLTemplate, data)
	return fmt.Sprintf(subjectDocumentSigning, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderDocumentExport(downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		PresenterVariables
		DownloadUrl string
	}{
		PresenterVariables: p.variables,
		DownloadUrl:        downloadUrl,
	}

	textBody, htmlBody, err = renderEmail(documentExportTextTemplate, documentExportHTMLTemplate, data)
	return subjectDocumentExport, textBody, htmlBody, err
}

func (p *Presenter) RenderFrameworkExport(downloadUrl string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		PresenterVariables
		DownloadUrl string
	}{
		PresenterVariables: p.variables,
		DownloadUrl:        downloadUrl,
	}

	textBody, htmlBody, err = renderEmail(frameworkExportTextTemplate, frameworkExportHTMLTemplate, data)
	return subjectFrameworkExport, textBody, htmlBody, err
}

func (p *Presenter) RenderTrustCenterAccess(organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		PresenterVariables
		OrganizationName string
	}{
		PresenterVariables: p.variables,
		OrganizationName:   organizationName,
	}

	textBody, htmlBody, err = renderEmail(trustCenterAccessTextTemplate, trustCenterAccessHTMLTemplate, data)
	return fmt.Sprintf(subjectTrustCenterAccess, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderTrustCenterDocumentAccessRejected(
	fileNames []string,
	organizationName string,
) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		PresenterVariables
		FileNames        []string
		OrganizationName string
	}{
		PresenterVariables: p.variables,
		FileNames:          fileNames,
		OrganizationName:   organizationName,
	}

	textBody, htmlBody, err = renderEmail(trustCenterDocumentAccessRejectedTextTemplate, trustCenterDocumentAccessRejectedHTMLTemplate, data)
	return fmt.Sprintf(subjectTrustCenterDocumentAccessRejected, organizationName), textBody, htmlBody, err
}

func (p *Presenter) RenderMagicLink(magicLinkUrlPath string, tokenString string, tokenDuration time.Duration, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	data := struct {
		PresenterVariables
		MagicLinkURL      string
		DurationInMinutes int
		OrganizationName  string
	}{
		PresenterVariables: p.variables,
		MagicLinkURL:       baseurl.MustParse(p.variables.BaseURL).AppendPath(magicLinkUrlPath).WithQuery("token", tokenString).MustString(),
		DurationInMinutes:  int(tokenDuration.Minutes()),
		OrganizationName:   organizationName,
	}

	textBody, htmlBody, err = renderEmail(magicLinkTextTemplate, magicLinkHTMLTemplate, data)
	return fmt.Sprintf(subjectMagicLink, organizationName), textBody, htmlBody, err
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
