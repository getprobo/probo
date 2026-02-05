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
	"path/filepath"
	texttemplate "text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/filemanager"
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

	staticAssetsDuration = 7 * 24 * time.Hour
)

type (
	Asset struct {
		Name       string
		ObjectKey  string
		BucketName string
		MimeType   string
	}

	PresenterConfig struct {
		BaseURL                         string
		PoweredByLogo                   Asset
		SenderCompanyName               string
		SenderCompanyWebsiteURL         string
		SenderCompanyLogo               Asset
		SenderCompanyHeadquarterAddress string
	}

	CommonVariables struct {
		// Static variables
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
		fm                *filemanager.Service
		config            PresenterConfig
		RecipientFullName string
	}
)

func (a *Asset) GetObjectKey() string {
	return a.ObjectKey
}

func (a *Asset) GetName() string {
	return a.Name
}

func (a *Asset) GetBucketName() string {
	return a.BucketName
}

func (a *Asset) GetMimeType() string {
	return a.MimeType
}

var _ filemanager.File = (*Asset)(nil)

func DefaultPresenterConfig(staticAssetsBucket string, baseURL string) PresenterConfig {
	return PresenterConfig{
		BaseURL: baseURL,
		PoweredByLogo: Asset{
			Name:       "probo-gray-small.png",
			ObjectKey:  "probo-gray-small.png",
			BucketName: staticAssetsBucket,
			MimeType:   "image/png",
		},
		SenderCompanyName:       "Probo",
		SenderCompanyWebsiteURL: "https://www.getprobo.com",
		SenderCompanyLogo: Asset{
			Name:       "probo.png",
			ObjectKey:  "probo.png",
			BucketName: staticAssetsBucket,
			MimeType:   "image/png",
		},
		SenderCompanyHeadquarterAddress: "Probo Inc, 490 Post St, STE 640, San Francisco, CA, 94102, US",
	}
}

func NewPresenterFromConfig(fileService *filemanager.Service, cfg PresenterConfig, fullName string) *Presenter {
	return &Presenter{
		fm:                fileService,
		config:            cfg,
		RecipientFullName: fullName,
	}
}

func NewPresenter(fileService *filemanager.Service, staticAssetsBucket string, baseURL string, fullName string) *Presenter {
	return NewPresenterFromConfig(
		fileService,
		DefaultPresenterConfig(staticAssetsBucket, baseURL),
		fullName,
	)
}

func UploadStaticAssets(ctx context.Context, s3Client *s3.Client, staticAssetsBucket string) error {
	subFS, err := fs.Sub(staticAssets, "assets")
	if err != nil {
		return fmt.Errorf("cannot create subtree file system: %w", err)
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
				Bucket: aws.String(staticAssetsBucket),
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

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot generate asset URLs: %w", err)
	}

	return nil
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

func (p *Presenter) getCommonVariables(ctx context.Context) (*CommonVariables, error) {
	poweredByLogoURL, err := p.fm.GenerateFileUrl(ctx, &p.config.PoweredByLogo, staticAssetsDuration)
	if err != nil {
		return nil, fmt.Errorf("cannot generate probo logo URL: %w", err)
	}
	senderCompanyLogoURL, err := p.fm.GenerateFileUrl(ctx, &p.config.SenderCompanyLogo, staticAssetsDuration)
	if err != nil {
		return nil, fmt.Errorf("cannot generate sender logo URL: %w", err)
	}

	return &CommonVariables{
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	invitationURL := baseurl.
		MustParse(vars.BaseURL).
		AppendPath(invitationURLPath).
		WithQuery("token", invitationToken).
		WithQuery("fullName", vars.RecipientFullName).
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

func (p *Presenter) RenderDocumentSigning(ctx context.Context, signinURLPath string, token string, organizationName string) (subject string, textBody string, htmlBody *string, err error) {
	vars, err := p.getCommonVariables(ctx)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get common variables: %w", err)
	}

	signingURL := baseurl.MustParse(vars.BaseURL).
		AppendPath(signinURLPath).
		WithQuery("token", token).
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
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
	vars, err := p.getCommonVariables(ctx)
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
