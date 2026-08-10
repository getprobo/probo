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

	"go.probo.inc/probo/pkg/baseurl"
)

// CSPFileStorageOrigin returns the http(s) origin the browser follows after a
// private file 307 to object storage, safe to embed as a CSP source.
//
// Custom endpoint + path-style (or IP host): endpoint origin.
// Custom endpoint + virtual-hosted hostname: scheme://{bucket}.{endpoint-host}.
// Empty endpoint + virtual-hosted: https://{bucket}.s3.{region}.amazonaws.com.
// Empty endpoint + path-style: https://s3.{region}.amazonaws.com.
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
		if usePathStyle {
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

	// The AWS SDK forces path-style addressing for IP endpoints, so the
	// browser follows the endpoint origin even when UsePathStyle is false.
	if usePathStyle || hostIsIP(parsed.Hostname()) {
		return parsed.Scheme + "://" + parsed.Host, nil
	}

	return parsed.Scheme + "://" + bucket + "." + parsed.Host, nil
}

func hostIsIP(hostname string) bool {
	return net.ParseIP(hostname) != nil
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
