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

package oidc

import (
	"context"
	"net/url"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	iamoauth2 "go.probo.inc/probo/pkg/iam/oauth2"
)

const (
	oauth2AuthorizePath          = "/api/connect/v1/oauth2/authorize"
	signInSourceQueryKey         = "source"
	signInSourceCompliancePortal = "compliance-portal"
)

// portalAuthorizeStateID returns the authorize `state` when continueURL is a
// Connect authorize request for a CIMD client stamped as a compliance-portal
// sign-in. Console paths, GID clients, and connector CIMD clients fail.
func portalAuthorizeStateID(continueURL string) (string, bool) {
	parsed, err := url.Parse(continueURL)
	if err != nil {
		return "", false
	}

	if parsed.Path != oauth2AuthorizePath {
		return "", false
	}

	query := parsed.Query()
	if query.Get(signInSourceQueryKey) != signInSourceCompliancePortal {
		return "", false
	}

	if !iamoauth2.IsCIMDClientID(query.Get("client_id")) {
		return "", false
	}

	stateID := query.Get("state")
	if stateID == "" {
		return "", false
	}

	return stateID, true
}

// allowsPersonalAccounts reports whether this OIDC login is finishing a
// compliance-portal authorize. The continue URL must look like a portal
// authorize request and its `state` must be a COMPLIANCE_PORTAL oidc state
// (created by portal /initiate), so a crafted console `continue` query
// cannot open the personal-account exception.
func (s *Service) allowsPersonalAccounts(ctx context.Context, continueURL string) bool {
	stateID, ok := portalAuthorizeStateID(continueURL)
	if !ok {
		return false
	}

	var portalState coredata.OIDCState

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return portalState.LoadByID(ctx, conn, stateID)
		},
	)
	if err != nil {
		return false
	}

	return portalState.Provider == coredata.OIDCProviderCompliancePortal
}
