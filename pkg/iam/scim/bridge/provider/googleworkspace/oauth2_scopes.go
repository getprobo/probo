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
