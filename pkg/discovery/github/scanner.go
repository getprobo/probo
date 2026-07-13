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

package github

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/discovery/vfs"
)

type discoveryScanner struct {
	api                 *apiClient
	fs                  vfs.FS
	org                 string
	conn                connector.Connection
	logger              *log.Logger
	scopes              scopeSet
	repoClassifier      RepoClassifier
	repoClassifications map[string]RepoClassification
}

func newDiscoveryScanner(
	httpClient *http.Client,
	org string,
	conn connector.Connection,
	logger *log.Logger,
	repoClassifier RepoClassifier,
) *discoveryScanner {
	api := newAPIClient(httpClient)

	if repoClassifier == nil {
		repoClassifier = DefaultRepoClassifier()
	}

	return &discoveryScanner{
		api:            api,
		fs:             newGitHubFS(api, org),
		org:            org,
		conn:           conn,
		logger:         logger,
		scopes:         newScopeSet(conn),
		repoClassifier: repoClassifier,
	}
}

func (s *discoveryScanner) scan(ctx context.Context) (*FactSheet, error) {
	sheet := &FactSheet{
		GitHubOrg: s.org,
	}

	if err := s.scanOrgSettings(ctx, sheet); err != nil {
		return nil, err
	}

	s.scanGovernance(ctx, sheet)
	s.scanOrgProfile(ctx, sheet)
	s.scanRepos(ctx, sheet)

	return sheet, nil
}
