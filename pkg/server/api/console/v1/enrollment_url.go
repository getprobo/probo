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

package console_v1

import (
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/baseurl"
)

// enrollmentURLs holds the public API origin and probo:// deep link issued
// when a device enrollment token is created.
type enrollmentURLs struct {
	ServerURL     string
	EnrollmentURL string
}

// buildEnrollmentURLs derives the agent server origin and deep link from the
// deployment base URL and a one-shot enrollment token.
func buildEnrollmentURLs(baseURL *baseurl.BaseURL, enrollmentToken string) (enrollmentURLs, error) {
	if baseURL == nil {
		return enrollmentURLs{}, fmt.Errorf("base URL is required")
	}

	if enrollmentToken == "" {
		return enrollmentURLs{}, fmt.Errorf("enrollment token is required")
	}

	serverURL := (&url.URL{
		Scheme: baseURL.Scheme(),
		Host:   baseURL.Host(),
	}).String()

	enrollURL := &url.URL{
		Scheme: "probo",
		Host:   "enroll",
	}
	query := enrollURL.Query()
	query.Set("server", serverURL)
	query.Set("token", enrollmentToken)
	enrollURL.RawQuery = query.Encode()

	return enrollmentURLs{
		ServerURL:     serverURL,
		EnrollmentURL: enrollURL.String(),
	}, nil
}
