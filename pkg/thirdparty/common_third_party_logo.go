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

package thirdparty

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/webinspect"
)

// maxLogoSize caps a downloaded logo image. Logos are small; the cap
// guards against an oversized or malicious response.
const maxLogoSize = 5 << 20 // 5 MiB

// prepareLogo discovers and uploads a vendor logo to S3, returning a
// fully-populated (but not yet inserted) File record for the caller to
// persist in its transaction. It is deterministic and best-effort: any
// failure logs and returns (nil, nil) so a missing logo never fails the
// enrichment run.
//
// It no-ops when the row already has a logo, when logo storage is not
// configured, or when no website is known. The S3 upload happens here,
// outside any transaction; the caller inserts the File row and links it
// via UpdateLogoFileID.
func (h *enrichmentHandler) prepareLogo(
	ctx context.Context,
	party coredata.CommonThirdParty,
	websiteURL string,
) *coredata.File {
	if party.LogoFileID != nil {
		return nil
	}

	if h.cfg.FileManager == nil || h.cfg.Bucket == "" {
		return nil
	}

	website := strings.TrimSpace(websiteURL)
	if website == "" {
		return nil
	}

	data, contentType, err := fetchCommonThirdPartyLogo(ctx, h.httpClient, website)
	if err != nil {
		h.logger.InfoCtx(
			ctx,
			"could not fetch common third party logo",
			log.String("common_third_party_id", party.ID.String()),
			log.Error(err),
		)

		return nil
	}

	objectKey, err := uuid.NewV7()
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot generate logo object key", log.Error(err))

		return nil
	}

	now := time.Now()
	fileRecord := &coredata.File{
		ID:             gid.New(gid.NilTenant, coredata.FileEntityType),
		OrganizationID: gid.Nil,
		BucketName:     h.cfg.Bucket,
		MimeType:       contentType,
		FileName:       party.Name + "-logo" + webinspect.ExtensionForMIME(contentType),
		FileKey:        objectKey.String(),
		FileSize:       int64(len(data)),
		Visibility:     coredata.FileVisibilityPublic,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	size, err := h.cfg.FileManager.PutFile(
		ctx,
		fileRecord,
		bytes.NewReader(data),
		map[string]string{
			"type":                  "common-third-party-logo",
			"common-third-party-id": party.ID.String(),
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot upload common third party logo",
			log.String("common_third_party_id", party.ID.String()),
			log.Error(err),
		)

		return nil
	}

	fileRecord.FileSize = size

	return fileRecord
}

// fetchCommonThirdPartyLogo finds the best logo for a website and
// downloads it. It first parses the page's <head> for icon links
// (webinspect), then falls back to well-known icon paths on the same
// host. The supplied client must enforce SSRF protection.
func fetchCommonThirdPartyLogo(
	ctx context.Context,
	client *http.Client,
	websiteURL string,
) (data []byte, contentType string, err error) {
	candidates := make([]string, 0, 3)

	pageInfo, parseErr := webinspect.Parse(ctx, client, websiteURL)
	if parseErr == nil {
		if logoURL, logoErr := webinspect.FindLogoURL(pageInfo); logoErr == nil {
			candidates = append(candidates, logoURL)
		}
	}

	if parsed, parseURLErr := url.Parse(websiteURL); parseURLErr == nil && parsed.Host != "" {
		base := url.URL{Scheme: parsed.Scheme, Host: parsed.Host}
		if base.Scheme == "" {
			base.Scheme = "https"
		}

		candidates = append(
			candidates,
			base.ResolveReference(&url.URL{Path: "/apple-touch-icon.png"}).String(),
			base.ResolveReference(&url.URL{Path: "/favicon.ico"}).String(),
		)
	}

	for _, candidate := range candidates {
		data, contentType, err = downloadImage(ctx, client, candidate)
		if err == nil {
			return data, contentType, nil
		}
	}

	return nil, "", fmt.Errorf("cannot fetch logo for %s", websiteURL)
}

// downloadImage fetches a single candidate URL and returns its bytes when
// the response is a non-empty image within the size cap.
func downloadImage(
	ctx context.Context,
	client *http.Client,
	rawURL string,
) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create logo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot fetch logo: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("cannot fetch logo: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	contentType = strings.TrimSpace(contentType)

	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("logo response is not an image: %q", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLogoSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("cannot read logo body: %w", err)
	}

	if len(body) > maxLogoSize {
		return nil, "", fmt.Errorf("logo response exceeds max size %d bytes", maxLogoSize)
	}

	if len(body) == 0 {
		return nil, "", fmt.Errorf("logo response is empty")
	}

	return body, contentType, nil
}
