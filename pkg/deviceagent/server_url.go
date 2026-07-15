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
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	USConsoleHost = "us.probo.com"
	EUConsoleHost = "eu.probo.com"
)

const (
	USConsoleURL = "https://" + USConsoleHost
	EUConsoleURL = "https://" + EUConsoleHost
)

const DefaultServerURL = USConsoleURL

func NormalizeServerURL(host string) (string, error) {
	raw := strings.TrimSpace(host)
	if raw == "" {
		return "", errors.New("server URL is required")
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("cannot parse server URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", fmt.Errorf("unsupported server URL scheme %q", parsed.Scheme)
	}

	if parsed.Hostname() == "" {
		return "", errors.New("server URL must include a hostname")
	}

	if parsed.User != nil {
		return "", errors.New("server URL must not include user credentials")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("server URL must not include a path")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("server URL must not include query parameters or fragments")
	}

	return (&url.URL{Scheme: scheme, Host: strings.ToLower(parsed.Host)}).String(), nil
}

// ConsoleEnrollURL returns the browser enrollment page URL for serverURL.
func ConsoleEnrollURL(serverURL string) (string, error) {
	normalized, err := NormalizeServerURL(serverURL)
	if err != nil {
		return "", err
	}

	enrollURL, err := url.JoinPath(normalized, "enroll")
	if err != nil {
		return "", fmt.Errorf("cannot build console enroll URL: %w", err)
	}

	return enrollURL, nil
}
