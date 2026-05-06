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

package googleworkspace

var (
	// OAuth2Scopes are the OAuth2 scopes required by the SCIM provisioning
	// bridge to read users from the Admin Directory API. The scopes are
	// intentionally limited to what the bridge actually consumes so the
	// integration also works for Google Cloud Identity (Free or Premium)
	// customers, not only Google Workspace customers. In particular,
	// admin.directory.userschema is a Workspace-only entitlement (custom
	// user fields are not available on Cloud Identity) and must not be
	// requested here, otherwise Cloud Identity-only admins cannot complete
	// the OAuth consent flow.
	OAuth2Scopes = []string{
		"https://www.googleapis.com/auth/admin.directory.user.readonly",
	}
)
