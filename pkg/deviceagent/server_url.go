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

const (
	USConsoleHost = "us.console.getprobo.com"
	EUConsoleHost = "eu.console.getprobo.com"
)

const (
	USConsoleURL = "https://" + USConsoleHost
	EUConsoleURL = "https://" + EUConsoleHost
)

const DefaultServerURL = USConsoleURL

func NormalizeServerURL(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("server URL is required")
	}

	lower := strings.ToLower(host)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		host = "https://" + host
	}

	parsed, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("cannot parse server URL: %w", err)
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("unsupported server URL scheme %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", errors.New("server URL must include a hostname")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("server URL must not include a path")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("server URL must not include query parameters or fragments")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
