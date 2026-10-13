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
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
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
		Devices []deviceEvidencePage
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

// ExportDeviceEvidence renders one PDF page per device and returns a
// presigned download URL for the file.
func (s *Service) ExportDeviceEvidence(
	ctx context.Context,
	scope coredata.Scoper,
	deviceIDs []gid.GID,
) (string, error) {
	if err := validateDeviceEvidenceIDs(deviceIDs); err != nil {
		return "", err
	}

	var (
		html           []byte
		organizationID gid.GID
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			pages := make([]deviceEvidencePage, 0, len(deviceIDs))

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

			rendered, err := renderDeviceEvidenceHTML(deviceEvidenceDocument{Devices: pages})
			if err != nil {
				return fmt.Errorf("cannot render device evidence: %w", err)
			}

			html = rendered

			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot build device evidence: %w", err)
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

func validateDeviceEvidenceIDs(deviceIDs []gid.GID) error {
	if len(deviceIDs) == 0 {
		return fmt.Errorf("%w: device ids are required", ErrInvalidDeviceEvidenceRequest)
	}

	if len(deviceIDs) > maxDeviceEvidenceDevices {
		return fmt.Errorf("%w: at most %d devices", ErrInvalidDeviceEvidenceRequest, maxDeviceEvidenceDevices)
	}

	seen := make(map[gid.GID]struct{}, len(deviceIDs))
	for _, deviceID := range deviceIDs {
		if deviceID.EntityType() != coredata.DeviceEntityType {
			return fmt.Errorf("%w: device id has the wrong type", ErrInvalidDeviceEvidenceRequest)
		}

		if _, ok := seen[deviceID]; ok {
			return fmt.Errorf("%w: duplicate device id", ErrInvalidDeviceEvidenceRequest)
		}

		seen[deviceID] = struct{}{}
	}

	return nil
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
