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

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// retoolCloudHost is the shared gateway every Retool Cloud organization is
	// reached through; the API token selects the organization.
	retoolCloudHost = "https://api.retool.com"

	// retoolAPIPath is the path every Retool API v2 base URL ends in, on the
	// cloud gateway and on a self-hosted instance alike.
	retoolAPIPath = "/api/v2"
)

// retoolRegistration wires the Retool access-review connector.
//
// Retool authenticates with an API token presented as Authorization: Bearer,
// the default scheme, and the token is bound to one organization, so there is
// nothing to pick (Pattern 3). Retool Cloud routes every organization through
// the shared api.retool.com gateway — the token itself selects the
// organization — so a cloud customer supplies no URL and Endpoints.APIBase
// serves them. A self-hosted instance has its own domain, which is what the
// optional baseUrl setting carries.
//
// The token needs the users:read scope. Retool refuses a token without it with
// 403 rather than 401, which the framework already reports as the operation
// being refused rather than the token being dead.
//
// No NewNameResolver: GET /organization/ returns settings, not a name, so the
// source keeps its generic name.
func retoolRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderRetool,
		DisplayName:      "Retool",
		DocumentationURL: accessReviewDocsURL("retool"),
		APIKey: &APIKeyConfig{
			ExtraSettings: []ExtraSetting{
				{Key: "baseUrl", Label: "Base URL (leave empty for Retool Cloud)", Required: false},
			},
			// No KeyFormat: Retool's tokens carry a retool_ prefix in
			// practice, but neither its API reference nor its authentication
			// guide documents one, and a shape inferred from a sample would
			// reject valid keys the day it mints a different one.
		},
		Endpoints: Endpoints{
			APIBase: retoolCloudHost + retoolAPIPath,
		},
		BuildProbeURL: buildRetoolProbeURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			apiBase, err := retoolAPIBase(conn, ep)
			if err != nil {
				return nil, fmt.Errorf("cannot create retool driver: %w", err)
			}

			return drivers.NewRetoolDriver(c, apiBase), nil
		},
	}
}

// retoolAPIBase resolves the API root a connector points at: the customer's
// self-hosted instance when they named one, and the cloud gateway in
// ep.APIBase otherwise.
//
// It composes the self-hosted root from the instance origin plus Retool's own
// /api/v2 prefix, so the two cases produce the same shape and the driver has
// one kind of base URL to join onto.
func retoolAPIBase(conn *coredata.Connector, ep Endpoints) (string, error) {
	settings, err := coredata.ConnectorSettings[coredata.RetoolConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read retool connector settings: %w", err)
	}

	if settings.BaseURL == "" {
		return ep.APIBase, nil
	}

	instanceURL, err := normalizeSelfHostedBaseURL(settings.BaseURL)
	if err != nil {
		return "", fmt.Errorf("cannot resolve retool base URL: %w", err)
	}

	parsed, err := url.Parse(instanceURL)
	if err != nil {
		return "", fmt.Errorf("cannot resolve retool base URL: %w", err)
	}

	// Reduce to the origin plus a clean path before deciding anything about
	// the suffix. Comparing against the raw path instead lets shapes a person
	// really does paste slip past: a "/api/v2/." keeps the prefix and gets a
	// second one appended, and a bare "/?" survives as ForceQuery and swallows
	// the prefix into the query string.
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawPath = ""

	cleaned := path.Clean("/" + strings.Trim(parsed.Path, "/"))
	if cleaned == "/" {
		cleaned = ""
	}

	// Retool's own docs write the base URL as https://retool.example.com/api/v2,
	// so a self-hosted admin copying it lands here with the prefix already on.
	// Appending unconditionally would build /api/v2/api/v2/users, which answers
	// 404 and reads as an unreachable instance rather than a mistyped setting.
	parsed.Path = strings.TrimSuffix(cleaned, retoolAPIPath) + retoolAPIPath

	return parsed.String(), nil
}
