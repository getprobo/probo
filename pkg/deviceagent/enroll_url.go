// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package deviceagent

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func ParseEnrollURL(raw string) (serverURL string, token string, err error) {
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

	if parsed.Host != "enroll" && parsed.Path != "/enroll" {
		return "", "", errors.New("enrollment URL path must be enroll")
	}

	query := parsed.Query()

	serverURL, err = NormalizeServerURL(query.Get("server"))
	if err != nil {
		return "", "", fmt.Errorf("invalid server in enrollment URL: %w", err)
	}

	token = strings.TrimSpace(query.Get("token"))
	if token == "" {
		return "", "", errors.New("enrollment token is missing")
	}

	return serverURL, token, nil
}
