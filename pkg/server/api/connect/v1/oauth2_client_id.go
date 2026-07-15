// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package connect_v1

import (
	"net/url"
)

const oauth2AuthorizeFullPath = "/api/connect/v1" + oauth2AuthorizePath

func oauth2ClientIDFromContinueURL(continueURL string) string {
	parsed, err := url.Parse(continueURL)
	if err != nil {
		return ""
	}

	// Only read client_id from the real authorize endpoint. The continue URL
	// itself is already validated by saferedirect before this runs.
	if parsed.Path != oauth2AuthorizeFullPath {
		return ""
	}

	return parsed.Query().Get("client_id")
}
