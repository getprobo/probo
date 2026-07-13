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
	"fmt"
	"net/url"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/discovery/vfs"
	"go.probo.inc/probo/pkg/discovery/vfs/gitfs"
)

func (s *discoveryScanner) buildWorkspace(
	ctx context.Context,
	repos []repoListItem,
) (vfs.FS, []string) {
	token, ok := oauthAccessToken(s.conn)
	if !ok {
		return s.apiFallbackFS(), []string{
			"git clone unavailable without oauth access token; falling back to github API file reads",
		}
	}

	workspace := gitfs.NewWorkspace()
	auth := &http.BasicAuth{Username: "x-access-token", Password: token}

	var limitations []string

	for _, repo := range repos {
		if err := ctx.Err(); err != nil {
			return s.apiFallbackFS(), append(limitations, "git clone interrupted")
		}

		repoURL, err := githubCloneURL(s.org, repo.Name)
		if err != nil {
			limitations = append(
				limitations,
				fmt.Sprintf("cannot build clone URL for repository %s", repo.Name),
			)

			continue
		}

		fs, err := gitfs.CloneRepo(ctx, repoURL, auth, repo.DefaultBranch)
		if err != nil {
			limitations = append(
				limitations,
				fmt.Sprintf("cannot clone repository %s: %v", repo.Name, err),
			)

			continue
		}

		workspace.AddRepo(repo.Name, fs)
	}

	if workspace.RepoCount() == 0 {
		return s.apiFallbackFS(), append(
			limitations,
			"no repositories cloned via git; falling back to github API file reads",
		)
	}

	if workspace.RepoCount() < len(repos) {
		limitations = append(
			limitations,
			fmt.Sprintf(
				"cloned %d of %d repositories via git; remaining repos use API fallbacks only",
				workspace.RepoCount(),
				len(repos),
			),
		)
	}

	return workspace, limitations
}

func (s *discoveryScanner) apiFallbackFS() vfs.FS {
	return newGitHubFS(s.api, s.org)
}

func oauthAccessToken(conn connector.Connection) (string, bool) {
	oauth2Conn, ok := conn.(*connector.OAuth2Connection)
	if !ok || oauth2Conn.AccessToken == "" {
		return "", false
	}

	return oauth2Conn.AccessToken, true
}

func githubCloneURL(org, repo string) (string, error) {
	return url.JoinPath("https://github.com", org, repo+".git")
}
