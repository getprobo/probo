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

func ParseEnrollURL(raw string) (serverURL string, enrollmentToken string, err error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", errors.New("enrollment URL is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("cannot parse enrollment URL: %w", err)
	}

	if parsed.Scheme != "probo" {
		return "", "", errors.New("enrollment URL must use probo scheme")
	}

	if parsed.Host != "enroll" {
		return "", "", errors.New("enrollment URL must be probo://enroll")
	}

	query := parsed.Query()

	serverURL, err = NormalizeServerURL(query.Get("server"))
	if err != nil {
		return "", "", fmt.Errorf("invalid server in enrollment URL: %w", err)
	}

	enrollmentToken = strings.TrimSpace(query.Get("token"))
	if enrollmentToken == "" {
		return "", "", errors.New("enrollment token is missing")
	}

	return serverURL, enrollmentToken, nil
}
