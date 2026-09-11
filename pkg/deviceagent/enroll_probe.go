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

package deviceagent

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type EnrollmentTrust string

const (
	TrustProboCloud EnrollmentTrust = "probo_cloud"
	TrustUnverified EnrollmentTrust = "unverified"
	TrustInsecure   EnrollmentTrust = "insecure"
	TrustUnknown    EnrollmentTrust = "unknown"
)

const enrollmentProbeTimeout = 5 * time.Second

// ProbeEnrollmentTrust classifies serverURL for the browser-enrollment
// confirm dialog. Probo hosts are trusted only when the same pinned
// HTTP client used for API calls can complete a TLS handshake.
func ProbeEnrollmentTrust(ctx context.Context, serverURL string) EnrollmentTrust {
	normalized, err := NormalizeServerURL(serverURL)
	if err != nil {
		return TrustUnknown
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return TrustUnknown
	}

	if parsed.Scheme == "http" {
		if isLoopbackHost(parsed.Hostname()) {
			return TrustUnknown
		}

		return TrustInsecure
	}

	if !IsProboServers(normalized) {
		return TrustUnverified
	}

	probeCtx, cancel := context.WithTimeout(ctx, enrollmentProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodHead, parsed.String(), nil)
	if err != nil {
		return TrustUnknown
	}

	client := NewClient(normalized, "", "probo-agent/enroll-probe")
	client.HTTP.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return TrustUnknown
	}

	_ = resp.Body.Close()

	return TrustProboCloud
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}
