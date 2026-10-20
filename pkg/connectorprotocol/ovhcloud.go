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

package connectorprotocol

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	OVHcloudRegionEU = "EU"
	OVHcloudRegionUS = "US"
)

// ovhcloudRegion is one OVHcloud API region. Each runs its own identity
// service, so an OAuth2 client and a credential minted in one region are
// unknown to the others.
type ovhcloudRegion struct {
	authHost string
	apiHost  string
}

var ovhcloudRegions = map[string]ovhcloudRegion{
	OVHcloudRegionEU: {authHost: "www.ovh.com", apiHost: "eu.api.ovh.com"},
	OVHcloudRegionUS: {authHost: "us.ovhcloud.com", apiHost: "api.us.ovhcloud.com"},
}

// NormalizeOVHcloudRegion returns the canonical region code for raw.
func NormalizeOVHcloudRegion(raw string) (string, error) {
	region := strings.ToUpper(strings.TrimSpace(raw))
	if _, ok := ovhcloudRegions[region]; !ok {
		return "", fmt.Errorf("cannot resolve ovhcloud region: must be EU or US")
	}

	return region, nil
}

func lookupOVHcloudRegion(raw string) (ovhcloudRegion, error) {
	region, err := NormalizeOVHcloudRegion(raw)
	if err != nil {
		return ovhcloudRegion{}, err
	}

	return ovhcloudRegions[region], nil
}

// OVHcloudAuthorizeURL returns the OAuth2 authorize endpoint of region.
func OVHcloudAuthorizeURL(region string) (string, error) {
	r, err := lookupOVHcloudRegion(region)
	if err != nil {
		return "", fmt.Errorf("cannot build authorize URL: %w", err)
	}

	u := url.URL{Scheme: "https", Host: r.authHost, Path: "/auth/oauth2/authorize"}

	return u.String(), nil
}

// OVHcloudTokenURL returns the OAuth2 token endpoint of region.
func OVHcloudTokenURL(region string) (string, error) {
	r, err := lookupOVHcloudRegion(region)
	if err != nil {
		return "", fmt.Errorf("cannot build token URL: %w", err)
	}

	u := url.URL{Scheme: "https", Host: r.authHost, Path: "/auth/oauth2/token"}

	return u.String(), nil
}

// OVHcloudAPIBase returns the versioned API root of region.
func OVHcloudAPIBase(region string) (string, error) {
	r, err := lookupOVHcloudRegion(region)
	if err != nil {
		return "", fmt.Errorf("cannot build API base: %w", err)
	}

	u := url.URL{Scheme: "https", Host: r.apiHost, Path: "/1.0"}

	return u.String(), nil
}
