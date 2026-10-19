// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package itam

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/page"
)

const (
	maxDeviceEvidenceDevices = 100
	deviceEvidenceURLExpiry  = 15 * time.Minute

	deviceEvidenceFilename    = "device-evidence.pdf"
	deviceEvidenceContentType = "application/pdf"

	// Same mark as the cookie banner branding.
	deviceEvidenceProboLogo = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" height="14" viewBox="0 0 90 24" aria-hidden="true"><g clip-path="url(#pc)"><rect width="24" height="24" x=".85" fill="currentColor" rx="5.14"/><path stroke="#fff" stroke-width="3.53" d="M20.57 6.17c2.96 4.47-6.43 13.97-20.32 5.66"/><path stroke="url(#pg)" stroke-width="3.53" d="M20.57 6.17c2.96 4.47-6.43 13.97-20.32 5.66"/><path fill="#fff" d="M19.07 7.1a1.77 1.77 0 0 0 3-1.86l-1.5.93-1.5.94ZM8.82 25.18l1.71-.41c-2.3-9.67.01-14.66 2.54-16.83A6.25 6.25 0 0 1 17 6.33c1.22-.01 1.87.44 2.08.78l1.5-.94 1.5-.93c-1.1-1.74-3.16-2.46-5.13-2.44-2.03.03-4.26.81-6.17 2.45-3.9 3.35-6.16 9.92-3.67 20.33l1.72-.41Z"/></g><path fill="currentColor" d="M66.87 0v7.32a4.84 4.84 0 0 1 1.48-.74 5.63 5.63 0 0 1 1.76-.28c1.63 0 3 .55 4.1 1.68a5.5 5.5 0 0 1 1.67 4.06c0 1.6-.56 2.93-1.69 4.06a5.49 5.49 0 0 1-4.08 1.69 5.8 5.8 0 0 1-4.06-1.56 5.48 5.48 0 0 1-1.71-4.27V0h2.53Zm3.24 8.72a3.2 3.2 0 0 0-3.24 3.32 3.18 3.18 0 0 0 3.24 3.32 3.2 3.2 0 0 0 3.24-3.32c0-1.9-1.43-3.32-3.24-3.32ZM52.75 7.98a5.53 5.53 0 0 1 4.08-1.68c1.61 0 2.93.55 4.06 1.68a5.51 5.51 0 0 1 1.69 4.06c0 1.58-.55 2.95-1.69 4.09a5.59 5.59 0 0 1-4.06 1.66 5.64 5.64 0 0 1-4.08-1.66 5.53 5.53 0 0 1-1.69-4.09c0-1.6.56-2.92 1.7-4.06Zm.84 4.06c0 .98.32 1.77.93 2.4.63.6 1.4.92 2.31.92.92 0 1.7-.31 2.3-.92.63-.64.94-1.42.94-2.4 0-.97-.31-1.76-.94-2.37a3.03 3.03 0 0 0-2.3-.95 3.22 3.22 0 0 0-3.24 3.32ZM79.32 7.98A5.53 5.53 0 0 1 83.4 6.3c1.6 0 2.93.55 4.06 1.68a5.51 5.51 0 0 1 1.69 4.06c0 1.58-.56 2.95-1.7 4.09a5.59 5.59 0 0 1-4.05 1.66 5.64 5.64 0 0 1-4.08-1.66 5.53 5.53 0 0 1-1.69-4.09c0-1.6.55-2.92 1.69-4.06Zm.84 4.06c0 .98.32 1.77.92 2.4.64.6 1.4.92 2.32.92.92 0 1.69-.31 2.3-.92.63-.64.94-1.42.94-2.4 0-.97-.31-1.76-.95-2.37a3.03 3.03 0 0 0-2.29-.95 3.22 3.22 0 0 0-3.24 3.32ZM34.24 16.76V24h-2.53V12.12c0-1.82.58-3.24 1.71-4.27a5.87 5.87 0 0 1 4.06-1.55c1.56 0 2.98.55 4.08 1.68a5.51 5.51 0 0 1 1.7 4.06 5.66 5.66 0 0 1-5.77 5.75 5.43 5.43 0 0 1-2.88-.77l-.36-.26Zm3.24-8.04a3.2 3.2 0 0 0-3.24 3.32c0 1.9 1.42 3.32 3.24 3.32a3.2 3.2 0 0 0 3.24-3.32 3.2 3.2 0 0 0-3.24-3.32ZM47.4 17.65h-2.5v-6.7c0-3.23 1.82-4.52 4.27-4.52h.6v2.42h-.47c-1.26 0-1.9.71-1.9 2.1v6.7Z"/><defs><linearGradient id="pg" x1="18.77" x2="1.2" y1="13.56" y2="13.56" gradientUnits="userSpaceOnUse"><stop stop-color="#101E1C" stop-opacity="0"/><stop offset="1" stop-color="#fff"/></linearGradient><clipPath id="pc"><rect width="24" height="24" x=".85" fill="#fff" rx="5.14"/></clipPath></defs></svg>`

	// The print frame defaults to a zero font size and ignores page CSS, so
	// all styles are inline.
	deviceEvidenceFooterTemplate = `<div style="width:100%; margin:0; padding:8px 16mm 0; box-sizing:border-box; border-top:1px solid #dce1d6; display:grid; grid-template-columns:1fr auto 1fr; align-items:center; font-size:8pt; line-height:1.2; letter-spacing:0.04em; text-transform:uppercase; font-family:Menlo, Consolas, 'Liberation Mono', Monaco, monospace; color:#687360;">` +
		`<div style="display:flex; align-items:center;">` + deviceEvidenceProboLogo + `</div>` +
		`<div><span style="display:inline-block; padding:2px 6px; background-color:#e5484d; color:#fff; -webkit-print-color-adjust:exact; print-color-adjust:exact;">CONFIDENTIAL</span></div>` +
		`<div style="text-align:right;"><span class="pageNumber"></span> / <span class="totalPages"></span></div>` +
		`</div>`

	evidencePendingLabel    = "(pending)"
	evidenceUnassignedLabel = "Unassigned"
	evidenceNeverLabel      = "Never"
)

var (
	//go:embed evidence_template.html
	deviceEvidenceTemplateSource string

	deviceEvidenceTemplate = template.Must(
		template.New("device-evidence").Parse(deviceEvidenceTemplateSource),
	)

	deviceEvidenceFileValidator = filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
	)

	deviceEvidenceCheckLabels = map[coredata.DevicePostureCheckKey]string{
		coredata.DevicePostureCheckKeyDiskEncryption:    "Disk encryption",
		coredata.DevicePostureCheckKeyScreenLock:        "Screen lock",
		coredata.DevicePostureCheckKeyFirewallEnabled:   "Firewall enabled",
		coredata.DevicePostureCheckKeyTimeSync:          "Time sync",
		coredata.DevicePostureCheckKeyOSVersion:         "OS version",
		coredata.DevicePostureCheckKeyAutoUpdate:        "Auto update",
		coredata.DevicePostureCheckKeyPasswordPolicy:    "Password policy",
		coredata.DevicePostureCheckKeyRemoteLogin:       "Remote login",
		coredata.DevicePostureCheckKeyMalwareProtection: "Malware protection",
	}
)

type (
	deviceEvidenceDocument struct {
		OrganizationName string
		LogoSrc          template.URL
		Kicker           string
		GeneratedAt      string
		GeneratedBy      string
		GeneratedOn      string
		Devices          []deviceEvidencePage
	}

	deviceEvidencePage struct {
		Hostname     string
		State        string
		Owner        string
		HardwareUUID string
		SerialNumber string
		Platform     string
		OSVersion    string
		AgentVersion string
		EnrolledAt   string
		LastSeenAt   string
		Postures     []deviceEvidencePosture
	}

	deviceEvidencePosture struct {
		Check  string
		Value  string
		Status string
	}
)

// ExportDeviceEvidence renders a PDF of the selected devices and returns a
// presigned download URL for the file. generatedByName is the exporter's
// name; generatedByEmail is used when that name is blank.
func (s *Service) ExportDeviceEvidence(
	ctx context.Context,
	scope coredata.Scoper,
	deviceIDs []gid.GID,
	generatedByName string,
	generatedByEmail string,
) (string, error) {
	if err := validateDeviceIDs(deviceIDs, maxDeviceEvidenceDevices); err != nil {
		return "", fmt.Errorf("invalid request: %w", err)
	}

	deviceIDs = gid.NewSet(deviceIDs...).Values()

	var (
		pages            []deviceEvidencePage
		organizationID   gid.GID
		organizationName string
		logoFile         *coredata.File
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			pages = make([]deviceEvidencePage, 0, len(deviceIDs))

			for _, deviceID := range deviceIDs {
				device := &coredata.Device{}
				if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
					return fmt.Errorf("cannot load device: %w", err)
				}

				organizationID = device.OrganizationID

				owner, err := deviceEvidenceOwner(ctx, conn, scope, device)
				if err != nil {
					return fmt.Errorf("cannot load device owner: %w", err)
				}

				postures, err := loadLatestDevicePostures(ctx, conn, scope, device.ID)
				if err != nil {
					return fmt.Errorf("cannot load device postures: %w", err)
				}

				pages = append(pages, newDeviceEvidencePage(device, owner, postures))
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organizationName = organization.Name

			file, err := deviceEvidenceLogoFile(ctx, conn, scope, organization)
			if err != nil {
				return err
			}

			logoFile = file

			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot build device evidence: %w", err)
	}

	generatedAt := time.Now()

	html, err := renderDeviceEvidenceHTML(
		deviceEvidenceDocument{
			OrganizationName: organizationName,
			LogoSrc:          s.deviceEvidenceLogoSrc(ctx, logoFile),
			Kicker:           evidenceDate(&generatedAt) + " / Security & trust",
			GeneratedAt:      evidenceDateTime(&generatedAt),
			GeneratedBy:      evidenceAuthor(generatedByName, generatedByEmail),
			GeneratedOn:      evidencePlatformURL(s.baseURL),
			Devices:          pages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot render device evidence: %w", err)
	}

	pdfReader, err := s.html2pdfConverter.GeneratePDF(
		ctx,
		html,
		html2pdf.RenderConfig{
			PageFormat:      html2pdf.PageFormatA4,
			Orientation:     html2pdf.OrientationPortrait,
			MarginTop:       html2pdf.NewMarginMillimeters(16),
			MarginBottom:    html2pdf.NewMarginMillimeters(16),
			MarginLeft:      html2pdf.NewMarginMillimeters(16),
			MarginRight:     html2pdf.NewMarginMillimeters(16),
			PrintBackground: true,
			FooterTemplate:  deviceEvidenceFooterTemplate,
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate device evidence pdf: %w", err)
	}

	pdfBytes, err := io.ReadAll(pdfReader)
	if err != nil {
		return "", fmt.Errorf("cannot read device evidence pdf: %w", err)
	}

	downloadURL, err := s.storeDeviceEvidencePDF(
		ctx,
		scope,
		organizationID,
		bytes.NewReader(pdfBytes),
		int64(len(pdfBytes)),
	)
	if err != nil {
		return "", fmt.Errorf("cannot store device evidence pdf: %w", err)
	}

	return downloadURL, nil
}

func (s *Service) storeDeviceEvidencePDF(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	content io.Reader,
	size int64,
) (string, error) {
	err := deviceEvidenceFileValidator.Validate(
		deviceEvidenceFilename,
		deviceEvidenceContentType,
		size,
	)
	if err != nil {
		return "", fmt.Errorf("cannot validate device evidence pdf: %w", err)
	}

	now := time.Now()
	file := &coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: organizationID,
		BucketName:     s.bucket,
		MimeType:       deviceEvidenceContentType,
		FileName:       deviceEvidenceFilename,
		FileKey:        uuid.MustNewV4().String(),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	fileSize, err := s.fileManager.PutFile(
		ctx,
		file,
		content,
		map[string]string{
			"organization-id": organizationID.String(),
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot upload device evidence: %w", err)
	}

	file.FileSize = fileSize

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert device evidence file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot save device evidence file: %w", err)
	}

	downloadURL, err := s.fileManager.GeneratePresignedURL(ctx, file, deviceEvidenceURLExpiry)
	if err != nil {
		return "", fmt.Errorf("cannot presign device evidence: %w", err)
	}

	return downloadURL, nil
}

func deviceEvidenceLogoFile(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organization *coredata.Organization,
) (*coredata.File, error) {
	fileID := organization.LogoFileID
	if fileID == nil {
		fileID = organization.HorizontalLogoFileID
	}

	if fileID == nil {
		return nil, nil
	}

	file := &coredata.File{}
	if err := file.LoadByID(ctx, conn, scope, *fileID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load organization logo: %w", err)
	}

	return file, nil
}

func (s *Service) deviceEvidenceLogoSrc(ctx context.Context, file *coredata.File) template.URL {
	if file == nil || !deviceEvidenceImageMIME(file.MimeType) {
		return ""
	}

	encoded, mimeType, err := s.fileManager.GetFileBase64(ctx, file)
	if err != nil || !deviceEvidenceImageMIME(mimeType) {
		return ""
	}

	return template.URL("data:" + mimeType + ";base64," + encoded)
}

func deviceEvidenceImageMIME(mimeType string) bool {
	switch mimeType {
	case "image/png", "image/jpeg", "image/webp", "image/svg+xml":
		return true
	default:
		return false
	}
}

func evidencePlatformURL(base *baseurl.BaseURL) string {
	if base == nil {
		return ""
	}

	return (&url.URL{
		Scheme: base.Scheme(),
		Host:   base.Host(),
	}).String()
}

func evidenceAuthor(fullName, email string) string {
	name := strings.TrimSpace(fullName)
	if name != "" {
		return name
	}

	return strings.TrimSpace(email)
}

func deviceEvidenceOwner(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	device *coredata.Device,
) (string, error) {
	if device.OwnerID == nil {
		return evidenceUnassignedLabel, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, *device.OwnerID); err != nil {
		return "", fmt.Errorf("cannot load owner profile: %w", err)
	}

	name := strings.TrimSpace(profile.FullName)
	if name == "" {
		return evidenceUnassignedLabel, nil
	}

	return name, nil
}

func loadLatestDevicePostures(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	deviceID gid.GID,
) (coredata.DevicePostures, error) {
	loaded, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.DevicePostureOrderField]{
			Field:     coredata.DevicePostureOrderFieldCheckKey,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DevicePostureOrderField]) ([]*coredata.DevicePosture, error) {
			var batch coredata.DevicePostures
			if err := batch.LoadLatestByDeviceID(ctx, conn, scope, deviceID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load latest device postures: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot collect latest device postures: %w", err)
	}

	return loaded, nil
}

func newDeviceEvidencePage(
	device *coredata.Device,
	owner string,
	postures coredata.DevicePostures,
) deviceEvidencePage {
	page := deviceEvidencePage{
		Hostname:     evidenceText(device.Hostname),
		State:        string(device.State),
		Owner:        owner,
		HardwareUUID: evidenceText(device.HardwareUUID),
		SerialNumber: evidenceText(device.SerialNumber),
		Platform:     evidencePlatform(device.Platform),
		OSVersion:    evidenceText(device.OSVersion),
		AgentVersion: evidenceText(device.AgentVersion),
		EnrolledAt:   evidenceDate(device.EnrolledAt),
		LastSeenAt:   evidenceDateTime(device.LastSeenAt),
		Postures:     make([]deviceEvidencePosture, 0, len(postures)),
	}

	for _, posture := range postures {
		value := coredata.ParseDevicePostureValue(posture.CheckKey, posture.Evidence)
		page.Postures = append(page.Postures, deviceEvidencePosture{
			Check:  evidenceCheckLabel(posture.CheckKey),
			Value:  evidencePostureValue(value),
			Status: evidencePostureStatus(posture.Status),
		})
	}

	return page
}

func renderDeviceEvidenceHTML(document deviceEvidenceDocument) ([]byte, error) {
	var buf bytes.Buffer
	if err := deviceEvidenceTemplate.Execute(&buf, document); err != nil {
		return nil, fmt.Errorf("cannot execute device evidence template: %w", err)
	}

	return buf.Bytes(), nil
}

func evidenceText(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return evidencePendingLabel
	}

	return *value
}

func evidencePlatform(platform *coredata.DevicePlatform) string {
	if platform == nil {
		return evidencePendingLabel
	}

	switch *platform {
	case coredata.DevicePlatformDarwin:
		return "macOS"
	case coredata.DevicePlatformLinux:
		return "Linux"
	case coredata.DevicePlatformFreeBSD:
		return "FreeBSD"
	case coredata.DevicePlatformWindows:
		return "Windows"
	default:
		return string(*platform)
	}
}

func evidenceDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return evidencePendingLabel
	}

	return value.UTC().Format("2 Jan 2006")
}

func evidenceDateTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return evidenceNeverLabel
	}

	return value.UTC().Format("2 Jan 2006, 15:04 UTC")
}

func evidenceCheckLabel(checkKey string) string {
	if label, ok := deviceEvidenceCheckLabels[coredata.DevicePostureCheckKey(checkKey)]; ok {
		return label
	}

	return checkKey
}

func evidencePostureValue(value coredata.DevicePostureValue) string {
	switch value.Kind {
	case coredata.DevicePostureValueKindOn:
		return "On"
	case coredata.DevicePostureValueKindOff:
		return "Off"
	case coredata.DevicePostureValueKindImmediate:
		return "Immediate"
	case coredata.DevicePostureValueKindSeconds:
		n := 0
		if value.Number != nil {
			n = *value.Number
		}

		return fmt.Sprintf("%ds", n)
	case coredata.DevicePostureValueKindMinPasswordLength:
		n := 0
		if value.Number != nil {
			n = *value.Number
		}

		return fmt.Sprintf("Min length %d", n)
	case coredata.DevicePostureValueKindConfigured:
		return "Configured"
	case coredata.DevicePostureValueKindNone:
		return "None"
	case coredata.DevicePostureValueKindText:
		text := strings.TrimSpace(value.Text)
		if text == "" {
			return "Unknown"
		}

		return text
	default:
		return "Unknown"
	}
}

func evidencePostureStatus(status coredata.DevicePostureStatus) string {
	switch status {
	case coredata.DevicePostureStatusPass:
		return "Pass"
	case coredata.DevicePostureStatusFail:
		return "Fail"
	case coredata.DevicePostureStatusUnknown:
		return "Unknown"
	case coredata.DevicePostureStatusNotApplicable:
		return "Not applicable"
	default:
		return string(status)
	}
}
