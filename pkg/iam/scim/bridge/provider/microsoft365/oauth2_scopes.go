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

package microsoft365

var (
	// OAuth2Scopes are the Microsoft Graph permission scopes required by the
	// SCIM provisioning bridge. The bridge reads users (including the
	// extended profile fields populated below) and their managers from the
	// Microsoft Graph API.
	//
	// - openid / profile: identifies the consenting admin.
	// - offline_access: required to receive a refresh token.
	// - User.Read.All: read all user profiles in the directory.
	// - Directory.Read.All: needed to read manager relationships and
	//   organizational data on every user without per-user consent.
	OAuth2Scopes = []string{
		"openid",
		"profile",
		"offline_access",
		"https://graph.microsoft.com/User.Read.All",
		"https://graph.microsoft.com/Directory.Read.All",
	}
)
