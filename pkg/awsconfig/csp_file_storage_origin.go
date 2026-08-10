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

package awsconfig

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"go.probo.inc/probo/pkg/baseurl"
)

// CSPFileStorageOrigin returns the http(s) origin the browser follows after a
// private file 307 to object storage, safe to embed as a CSP source.
//
// Matches aws-sdk-go-v2 S3 addressing: path-style (or a non-virtual-hostable
// bucket) uses the endpoint / regional S3 host; otherwise the bucket is
// prefixed as a subdomain.
func CSPFileStorageOrigin(
	endpoint string,
	region string,
	bucket string,
	usePathStyle bool,
) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	region = strings.TrimSpace(region)
	bucket = strings.TrimSpace(bucket)

	if bucket == "" {
		return "", fmt.Errorf("CSP file storage origin requires a bucket")
	}

	if region == "" {
		region = DefaultRegion
	}

	if err := validateBucketForCSP(bucket); err != nil {
		return "", err
	}

	raw, err := fileStorageURL(endpoint, region, bucket, usePathStyle)
	if err != nil {
		return "", err
	}

	origin, err := baseurl.CSPOrigin(raw)
	if err != nil {
		return "", fmt.Errorf("invalid CSP file storage origin: %w", err)
	}

	return origin, nil
}

func fileStorageURL(
	endpoint string,
	region string,
	bucket string,
	usePathStyle bool,
) (string, error) {
	if endpoint == "" {
		// Native AWS endpoints are always HTTPS, so dotted buckets are not
		// virtual-hostable and the SDK falls back to the regional host.
		if usePathStyle || !bucketIsVirtualHostable(bucket, true) {
			return "https://s3." + region + ".amazonaws.com", nil
		}

		return "https://" + bucket + ".s3." + region + ".amazonaws.com", nil
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid CSP file storage endpoint: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("CSP file storage endpoint must include scheme and host")
	}

	if parsed.User != nil {
		return "", fmt.Errorf("CSP file storage endpoint must not include userinfo")
	}

	// IP hosts, explicit path-style, and non-virtual-hostable buckets all keep
	// the bucket in the path — the browser follows the endpoint origin.
	if usePathStyle ||
		hostIsIP(parsed.Hostname()) ||
		!bucketIsVirtualHostable(bucket, parsed.Scheme == "https") {
		return parsed.Scheme + "://" + parsed.Host, nil
	}

	return parsed.Scheme + "://" + bucket + "." + parsed.Host, nil
}

func hostIsIP(hostname string) bool {
	return net.ParseIP(hostname) != nil
}

// bucketIsVirtualHostable mirrors aws-sdk-go-v2's IsVirtualHostableS3Bucket:
// HTTPS forbids dots (TLS wildcard/cert issues); HTTP may use dotted labels
// when each label is a valid 3–63 char host label.
func bucketIsVirtualHostable(bucket string, https bool) bool {
	if net.ParseIP(bucket) != nil {
		return false
	}

	var labels []string
	if https {
		labels = []string{bucket}
	} else {
		labels = strings.Split(bucket, ".")
	}

	for _, label := range labels {
		if l := len(label); l < 3 || l > 63 {
			return false
		}

		for _, r := range label {
			if r >= 'A' && r <= 'Z' {
				return false
			}
		}

		if !smithyhttp.ValidHostLabel(label) {
			return false
		}
	}

	return true
}

func validateBucketForCSP(bucket string) error {
	if strings.ContainsAny(bucket, " \t\r\n;,\"'\\*/?#@:") {
		return fmt.Errorf("CSP file storage bucket contains invalid characters")
	}

	if bucket == "." || bucket == ".." {
		return fmt.Errorf("CSP file storage bucket is invalid")
	}

	return nil
}
