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

package cmdutil

import "errors"

// OrganizationRequiredMessage is returned when a command needs an organization
// and neither --org nor the default saved by `prb auth login` is set.
const OrganizationRequiredMessage = "organization is required; pass --org or --organization, or set a default with 'prb auth login'"

// OrganizationRequiredMessageOrgOnly is OrganizationRequiredMessage without
// the --organization hint, for commands where --organization already means
// something else, such as connector connect's boolean that connects a cloud
// organization instead of supplying an org ID.
const OrganizationRequiredMessageOrgOnly = "organization is required; pass --org, or set a default with 'prb auth login'"

// ResolveOrganization returns the organization a command should use. An
// explicit flag wins; otherwise the host's saved default is used.
func ResolveOrganization(flag, configured string) (string, error) {
	return resolveOrganization(flag, configured, OrganizationRequiredMessage)
}

// ResolveOrganizationOrgOnly is ResolveOrganization, except the error omits
// the --organization hint. Use it on commands where --organization does not
// supply an org ID, such as connector connect.
func ResolveOrganizationOrgOnly(flag, configured string) (string, error) {
	return resolveOrganization(flag, configured, OrganizationRequiredMessageOrgOnly)
}

func resolveOrganization(flag, configured, missing string) (string, error) {
	if flag == "" {
		flag = configured
	}

	if flag == "" {
		return "", errors.New(missing)
	}

	return flag, nil
}
