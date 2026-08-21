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

package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

const (
	githubCredentialTypePAT          = "personal access token"
	githubCredentialTypeOAuthApp     = "OAuth app token"
	githubCredentialTypeGitHubApp    = "GitHub app token"
	githubInstallationExternalPrefix = "installation:"
	githubPATExternalPrefix          = "pat:"
	githubCredentialExternalPrefix   = "credential:"
	githubDeployKeyExternalPrefix    = "deploy-key:"
)

type (
	githubInstallationsPage struct {
		Installations []githubInstallation `json:"installations"`
	}

	githubInstallation struct {
		ID          int64             `json:"id"`
		AppSlug     string            `json:"app_slug"`
		Permissions map[string]string `json:"permissions"`
		CreatedAt   string            `json:"created_at"`
		SuspendedAt *string           `json:"suspended_at"`
	}

	githubFineGrainedPAT struct {
		ID              int64   `json:"id"`
		TokenID         int64   `json:"token_id"`
		TokenName       string  `json:"token_name"`
		TokenExpired    bool    `json:"token_expired"`
		TokenLastUsedAt *string `json:"token_last_used_at"`
		AccessGrantedAt string  `json:"access_granted_at"`
		Owner           struct {
			Login string `json:"login"`
		} `json:"owner"`
		Permissions struct {
			Organization map[string]string `json:"organization"`
			Repository   map[string]string `json:"repository"`
		} `json:"permissions"`
	}

	githubCredentialAuthorization struct {
		Login                     string  `json:"login"`
		CredentialID              int64   `json:"credential_id"`
		CredentialType            string  `json:"credential_type"`
		CredentialAuthorizedAt    string  `json:"credential_authorized_at"`
		CredentialAccessedAt      *string `json:"credential_accessed_at"`
		AuthorizedCredentialID    *int64  `json:"authorized_credential_id"`
		AuthorizedCredentialTitle *string `json:"authorized_credential_title"`
		AuthorizedCredentialNote  *string `json:"authorized_credential_note"`
	}

	gitHubDeployKeysRequest struct {
		Query     string                    `json:"query"`
		Variables gitHubDeployKeysVariables `json:"variables"`
	}

	gitHubDeployKeysVariables struct {
		Org   string  `json:"org"`
		After *string `json:"after"`
	}

	gitHubDeployKeysResponse struct {
		Data struct {
			Organization *struct {
				Repositories struct {
					Nodes []struct {
						Name       string `json:"name"`
						DeployKeys *struct {
							Nodes []githubGraphQLDeployKey `json:"nodes"`
						} `json:"deployKeys"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"organization"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	githubGraphQLDeployKey struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		CreatedAt string `json:"createdAt"`
		ReadOnly  bool   `json:"readOnly"`
	}

	githubDeployKeyRecord struct {
		Repo string
		Key  githubGraphQLDeployKey
	}
)

// githubDeployKeysQuery batches repository deploy keys. REST GET
// /repos/{owner}/{repo}/keys is one call per repo and would blow the
// campaign fetch timeout on a large org. Nested pagination stops at 100
// keys per repo; orgs with more than that on a single repo are rare.
const githubDeployKeysQuery = `query AccessReviewGitHubDeployKeys($org: String!, $after: String) { organization(login: $org) { repositories(first: 100, after: $after) { nodes { name deployKeys(first: 100) { nodes { id title createdAt readOnly } } } pageInfo { hasNextPage endCursor } } } }`

func (d *GitHubDriver) appendServiceAccounts(ctx context.Context, records []AccountRecord) []AccountRecord {
	if err := ctx.Err(); err != nil {
		d.logger.WarnCtx(ctx, "cannot fetch github service accounts", log.Error(err))

		return records
	}

	installations, err := d.fetchAllInstallations(ctx)
	if err != nil {
		d.logger.WarnCtx(ctx, "cannot fetch github app installations", log.Error(err))
	} else {
		for _, inst := range installations {
			if rec, ok := githubInstallationRecord(inst); ok {
				records = append(records, rec)
			}
		}
	}

	seenPATs := make(map[string]struct{})

	pats, err := d.fetchAllFineGrainedPATs(ctx)
	if err != nil {
		d.logger.WarnCtx(ctx, "cannot fetch github fine-grained personal access tokens", log.Error(err))
	} else {
		for _, pat := range pats {
			rec, ok := githubFineGrainedPATRecord(pat)
			if !ok {
				continue
			}

			seenPATs[rec.ExternalID] = struct{}{}
			githubRememberPATIDs(seenPATs, pat)
			records = append(records, rec)
		}
	}

	credentials, err := d.fetchAllCredentialAuthorizations(ctx)
	if err != nil {
		d.logger.WarnCtx(ctx, "cannot fetch github credential authorizations", log.Error(err))
	} else {
		for _, cred := range credentials {
			if !githubCredentialIsToken(cred.CredentialType) {
				continue
			}

			if githubCredentialDuplicatesPAT(seenPATs, cred) {
				continue
			}

			if rec, ok := githubCredentialRecord(cred); ok {
				records = append(records, rec)
			}
		}
	}

	deployKeys, err := d.fetchAllDeployKeys(ctx)
	if err != nil {
		d.logger.WarnCtx(ctx, "cannot fetch github deploy keys", log.Error(err))
	} else {
		for _, item := range deployKeys {
			if rec, ok := githubDeployKeyAccountRecord(item); ok {
				records = append(records, rec)
			}
		}
	}

	return records
}

func (d *GitHubDriver) fetchAllInstallations(ctx context.Context) ([]githubInstallation, error) {
	endpoint, err := githubOrgCollectionURL(d.baseURL, d.org, githubInstallationsSegment)
	if err != nil {
		return nil, err
	}

	var all []githubInstallation

	for range maxPaginationPages {
		var page githubInstallationsPage

		nextURL, err := d.getJSON(ctx, endpoint, "app installations", &page)
		if err != nil {
			return nil, err
		}

		all = append(all, page.Installations...)

		if nextURL == "" {
			return all, nil
		}

		endpoint = nextURL
	}

	return nil, fmt.Errorf("cannot list all github app installations: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchAllFineGrainedPATs(ctx context.Context) ([]githubFineGrainedPAT, error) {
	endpoint, err := githubOrgCollectionURL(d.baseURL, d.org, githubPersonalAccessTokensSegment)
	if err != nil {
		return nil, err
	}

	return githubListPages[githubFineGrainedPAT](ctx, d, endpoint, "fine-grained personal access tokens")
}

func (d *GitHubDriver) fetchAllCredentialAuthorizations(ctx context.Context) ([]githubCredentialAuthorization, error) {
	endpoint, err := githubOrgCollectionURL(d.baseURL, d.org, githubCredentialAuthorizationsSegment)
	if err != nil {
		return nil, err
	}

	return githubListPages[githubCredentialAuthorization](ctx, d, endpoint, "credential authorizations")
}

func (d *GitHubDriver) fetchAllDeployKeys(ctx context.Context) ([]githubDeployKeyRecord, error) {
	var (
		keys  []githubDeployKeyRecord
		after *string
	)

	for range maxPaginationPages {
		page, next, err := d.fetchDeployKeysPage(ctx, after)
		if err != nil {
			return nil, err
		}

		keys = append(keys, page...)

		if next == nil {
			return keys, nil
		}

		after = next
	}

	return nil, fmt.Errorf("cannot list all github deploy keys: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchDeployKeysPage(
	ctx context.Context,
	after *string,
) ([]githubDeployKeyRecord, *string, error) {
	endpoint, err := url.JoinPath(d.baseURL, githubGraphqlPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot build github graphql URL: %w", err)
	}

	payload, err := json.Marshal(
		gitHubDeployKeysRequest{
			Query: githubDeployKeysQuery,
			Variables: gitHubDeployKeysVariables{
				Org:   d.org,
				After: after,
			},
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot marshal github deploy keys query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create github deploy keys request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot execute github deploy keys request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("cannot fetch github deploy keys: unexpected status %d", httpResp.StatusCode)
	}

	var resp gitHubDeployKeysResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, nil, fmt.Errorf("cannot decode github deploy keys response: %w", err)
	}

	if resp.Data.Organization == nil {
		if len(resp.Errors) > 0 {
			return nil, nil, fmt.Errorf("cannot fetch github deploy keys: graphql error")
		}

		return nil, nil, nil
	}

	var keys []githubDeployKeyRecord

	for _, repo := range resp.Data.Organization.Repositories.Nodes {
		if repo.DeployKeys == nil {
			continue
		}

		for _, key := range repo.DeployKeys.Nodes {
			keys = append(
				keys,
				githubDeployKeyRecord{
					Repo: repo.Name,
					Key:  key,
				},
			)
		}
	}

	pageInfo := resp.Data.Organization.Repositories.PageInfo
	if !pageInfo.HasNextPage || pageInfo.EndCursor == "" {
		return keys, nil, nil
	}

	next := pageInfo.EndCursor

	return keys, &next, nil
}

func githubOrgCollectionURL(baseURL, org, segment string) (string, error) {
	u, err := url.JoinPath(baseURL, githubOrgsSegment, url.PathEscape(org), segment)
	if err != nil {
		return "", fmt.Errorf("cannot build github %s URL: %w", segment, err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("cannot parse github %s URL: %w", segment, err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

func githubListPages[T any](
	ctx context.Context,
	d *GitHubDriver,
	endpoint, what string,
) ([]T, error) {
	var all []T

	for range maxPaginationPages {
		var page []T

		nextURL, err := d.getJSON(ctx, endpoint, what, &page)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)

		if nextURL == "" {
			return all, nil
		}

		endpoint = nextURL
	}

	return nil, fmt.Errorf("cannot list all github %s: %w", what, ErrPaginationLimitReached)
}

func (d *GitHubDriver) getJSON(ctx context.Context, pageURL, what string, dst any) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create github %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute github %s request: %w", what, err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch github %s: unexpected status %d", what, httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(dst); err != nil {
		return "", fmt.Errorf("cannot decode github %s response: %w", what, err)
	}

	nextURL := rfc5988.FindByRel(httpResp.Header.Get("Link"), "next")
	if nextURL == "" {
		return "", nil
	}

	nextURL, err = sameHostNextPageURL("github", d.baseURL, nextURL)
	if err != nil {
		return "", err
	}

	return nextURL, nil
}

func githubInstallationRecord(inst githubInstallation) (AccountRecord, bool) {
	if inst.ID == 0 {
		return AccountRecord{}, false
	}

	fullName := strings.TrimSpace(inst.AppSlug)
	if fullName == "" {
		fullName = "GitHub App"
	}

	suspended := inst.SuspendedAt != nil && strings.TrimSpace(*inst.SuspendedAt) != ""

	return AccountRecord{
		FullName:    fullName,
		Roles:       []string{"github-app"},
		Active:      new(!suspended),
		IsAdmin:     new(inst.Permissions["administration"] == "write"),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodOAuth2,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		CreatedAt:   parseRFC3339Ptr(inst.CreatedAt),
		ExternalID:  githubInstallationExternalPrefix + strconv.FormatInt(inst.ID, 10),
	}, true
}

func githubFineGrainedPATRecord(pat githubFineGrainedPAT) (AccountRecord, bool) {
	id := pat.TokenID
	if id == 0 {
		id = pat.ID
	}

	if id == 0 {
		return AccountRecord{}, false
	}

	fullName := strings.TrimSpace(pat.TokenName)
	if fullName == "" {
		owner := strings.TrimSpace(pat.Owner.Login)
		if owner != "" {
			fullName = "Fine-grained PAT (" + owner + ")"
		} else {
			fullName = "Fine-grained PAT"
		}
	}

	return AccountRecord{
		FullName:    fullName,
		Roles:       []string{"fine-grained PAT"},
		Active:      new(!pat.TokenExpired),
		IsAdmin:     new(githubPATIsAdmin(pat)),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		LastLogin:   parseRFC3339Ptr(ref.UnrefOrZero(pat.TokenLastUsedAt)),
		CreatedAt:   parseRFC3339Ptr(pat.AccessGrantedAt),
		ExternalID:  githubPATExternalPrefix + strconv.FormatInt(id, 10),
	}, true
}

func githubPATIsAdmin(pat githubFineGrainedPAT) bool {
	return pat.Permissions.Organization["administration"] == "write" ||
		pat.Permissions.Repository["administration"] == "write"
}

func githubRememberPATIDs(seen map[string]struct{}, pat githubFineGrainedPAT) {
	if pat.ID != 0 {
		seen[githubPATExternalPrefix+strconv.FormatInt(pat.ID, 10)] = struct{}{}
	}

	if pat.TokenID != 0 {
		seen[githubPATExternalPrefix+strconv.FormatInt(pat.TokenID, 10)] = struct{}{}
	}
}

func githubCredentialIsToken(credentialType string) bool {
	switch credentialType {
	case githubCredentialTypePAT, githubCredentialTypeOAuthApp, githubCredentialTypeGitHubApp:
		return true
	default:
		return false
	}
}

func githubCredentialAuthMethod(credentialType string) coredata.AccessReviewEntryAuthMethod {
	switch credentialType {
	case githubCredentialTypeOAuthApp, githubCredentialTypeGitHubApp:
		return coredata.AccessReviewEntryAuthMethodOAuth2
	default:
		return coredata.AccessReviewEntryAuthMethodAPIKey
	}
}

func githubCredentialDuplicatesPAT(seen map[string]struct{}, cred githubCredentialAuthorization) bool {
	if cred.CredentialType != githubCredentialTypePAT {
		return false
	}

	if cred.CredentialID != 0 {
		if _, ok := seen[githubPATExternalPrefix+strconv.FormatInt(cred.CredentialID, 10)]; ok {
			return true
		}
	}

	if cred.AuthorizedCredentialID != nil && *cred.AuthorizedCredentialID != 0 {
		if _, ok := seen[githubPATExternalPrefix+strconv.FormatInt(*cred.AuthorizedCredentialID, 10)]; ok {
			return true
		}
	}

	return false
}

func githubCredentialRecord(cred githubCredentialAuthorization) (AccountRecord, bool) {
	if cred.CredentialID == 0 {
		return AccountRecord{}, false
	}

	fullName := strings.TrimSpace(ref.UnrefOrZero(cred.AuthorizedCredentialTitle))
	if fullName == "" {
		fullName = strings.TrimSpace(ref.UnrefOrZero(cred.AuthorizedCredentialNote))
	}

	if fullName == "" {
		kind := strings.TrimSpace(cred.CredentialType)
		owner := strings.TrimSpace(cred.Login)
		switch {
		case kind != "" && owner != "":
			fullName = kind + " (" + owner + ")"
		case kind != "":
			fullName = kind
		default:
			fullName = "GitHub token"
		}
	}

	role := strings.TrimSpace(cred.CredentialType)
	roles := []string{}
	if role != "" {
		roles = []string{role}
	}

	return AccountRecord{
		FullName:    fullName,
		Roles:       roles,
		Active:      new(true),
		IsAdmin:     new(false),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  githubCredentialAuthMethod(cred.CredentialType),
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		LastLogin:   parseRFC3339Ptr(ref.UnrefOrZero(cred.CredentialAccessedAt)),
		CreatedAt:   parseRFC3339Ptr(cred.CredentialAuthorizedAt),
		ExternalID:  githubCredentialExternalPrefix + strconv.FormatInt(cred.CredentialID, 10),
	}, true
}

func githubDeployKeyAccountRecord(item githubDeployKeyRecord) (AccountRecord, bool) {
	id := strings.TrimSpace(item.Key.ID)
	if id == "" {
		return AccountRecord{}, false
	}

	title := strings.TrimSpace(item.Key.Title)
	if title == "" {
		title = "Deploy key"
	}

	repo := strings.TrimSpace(item.Repo)
	fullName := title
	if repo != "" {
		fullName = title + " (" + repo + ")"
	}

	role := "deploy-key"
	if !item.Key.ReadOnly {
		role = "deploy-key (read-write)"
	}

	return AccountRecord{
		FullName:    fullName,
		Roles:       []string{role},
		Active:      new(true),
		IsAdmin:     new(false),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodSSH,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		CreatedAt:   parseRFC3339Ptr(item.Key.CreatedAt),
		ExternalID:  githubDeployKeyExternalPrefix + id,
	}, true
}
