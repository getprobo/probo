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

package gcp

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	errInvalidProviderResource    = errors.New("workloadIdentityProvider is not a workload identity provider resource")
	errInvalidServiceAccountEmail = errors.New("serviceAccountEmail is not a service account email")

	providerResourcePattern = regexp.MustCompile(
		`^(?:(?:https:)?//([^/]+)/)?` +
			`/?projects/([1-9][0-9]*)` +
			`/locations/global` +
			`/workloadIdentityPools/([a-z0-9][a-z0-9-]{2,30}[a-z0-9])` +
			`/providers/([a-z0-9][a-z0-9-]{2,30}[a-z0-9])/?$`,
	)
	serviceAccountEmailPattern = regexp.MustCompile(
		`^[a-z][a-z0-9-]{4,28}[a-z0-9]@[a-z][a-z0-9-]{4,28}[a-z0-9](?:\.s3ns)?\.iam\.gserviceaccount\.com$`,
	)
)

type (
	// ConnectorSettings is the customer-supplied WIF provider resource and the
	// service account Probo impersonates. Both values are public knowledge.
	ConnectorSettings struct {
		WorkloadIdentityProvider string
		ServiceAccountEmail      string
	}

	providerResource struct {
		iamHost       string
		projectNumber string
		poolID        string
		providerID    string
	}
)

// NewConnectorSettings validates the customer-supplied provider resource and
// service account email. Values are stored trimmed in canonical form. Returned
// errors are safe to show a client: they never echo the resource or email.
func NewConnectorSettings(providerResource, serviceAccountEmail string) (ConnectorSettings, error) {
	parsed, err := parseProviderResource(providerResource)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create gcp connector: %w", err)
	}

	email, err := parseServiceAccountEmail(serviceAccountEmail)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create gcp connector: %w", err)
	}

	if _, err := inferUniverse(parsed, email); err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create gcp connector: %w", err)
	}

	return ConnectorSettings{
		WorkloadIdentityProvider: canonicalProviderResource(parsed),
		ServiceAccountEmail:      email,
	}, nil
}

func parseProviderResource(raw string) (providerResource, error) {
	matches := providerResourcePattern.FindStringSubmatch(strings.TrimSpace(raw))
	if matches == nil {
		return providerResource{}, errInvalidProviderResource
	}

	host := matches[1]
	if host != "" && !supportedIAMHost(host) {
		return providerResource{}, errUnsupportedUniverse
	}

	return providerResource{
		iamHost:       host,
		projectNumber: matches[2],
		poolID:        matches[3],
		providerID:    matches[4],
	}, nil
}

func parseServiceAccountEmail(raw string) (string, error) {
	email := strings.TrimSpace(raw)
	if !serviceAccountEmailPattern.MatchString(email) {
		return "", errInvalidServiceAccountEmail
	}

	return email, nil
}

func canonicalProviderResource(p providerResource) string {
	return strings.Join(
		[]string{
			"projects",
			p.projectNumber,
			"locations",
			"global",
			"workloadIdentityPools",
			p.poolID,
			"providers",
			p.providerID,
		},
		"/",
	)
}
