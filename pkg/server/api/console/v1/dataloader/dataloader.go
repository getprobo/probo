// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package dataloader

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/vikstrous/dataloadgen"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/thirdparty"
)

type (
	ctxKey struct{ name string }

	// AuthorizeKey identifies an authorize call. ResourceAttributes is the
	// canonical (sorted-key) JSON encoding of the attributes passed via
	// authz.WithAttr, so calls with the same logical inputs share a key and
	// batch together.
	AuthorizeKey struct {
		ResourceID          gid.GID
		Action              string
		ResourceAttributes  string
		DryRun              bool
		SkipAssumptionCheck bool
	}

	AuthorizeResult struct {
		Scope *coredata.Scope
	}

	Loaders struct {
		Organization         *dataloadgen.Loader[gid.GID, *coredata.Organization]
		Framework            *dataloadgen.Loader[gid.GID, *coredata.Framework]
		Control              *dataloadgen.Loader[gid.GID, *coredata.Control]
		ThirdParty           *dataloadgen.Loader[gid.GID, *coredata.ThirdParty]
		Document             *dataloadgen.Loader[gid.GID, *coredata.Document]
		Profile              *dataloadgen.Loader[gid.GID, *coredata.MembershipProfile]
		Risk                 *dataloadgen.Loader[gid.GID, *coredata.Risk]
		Measure              *dataloadgen.Loader[gid.GID, *coredata.Measure]
		Task                 *dataloadgen.Loader[gid.GID, *coredata.Task]
		File                 *dataloadgen.Loader[gid.GID, *coredata.File]
		CookieBanner         *dataloadgen.Loader[gid.GID, *coredata.CookieBanner]
		CookieCategory       *dataloadgen.Loader[gid.GID, *coredata.CookieCategory]
		CommonTrackerPattern *dataloadgen.Loader[gid.GID, *coredata.CommonTrackerPattern]
		CommonThirdParty     *dataloadgen.Loader[gid.GID, *coredata.CommonThirdParty]
		Authorize            *dataloadgen.Loader[AuthorizeKey, AuthorizeResult]
	}

	batchFetcher struct {
		probo        *probo.Service
		iam          *iam.Service
		cookieBanner *cookiebanner.Service
		thirdParty   *thirdparty.Service
	}
)

var loadersKey = &ctxKey{name: "dataloaders"}

func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func NewMiddleware(proboSvc *probo.Service, iamSvc *iam.Service, cookieBannerSvc *cookiebanner.Service, thirdPartySvc *thirdparty.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				f := &batchFetcher{
					probo:        proboSvc,
					iam:          iamSvc,
					cookieBanner: cookieBannerSvc,
					thirdParty:   thirdPartySvc,
				}
				loaders := f.newLoaders()
				ctx := context.WithValue(r.Context(), loadersKey, loaders)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func (f *batchFetcher) newLoaders() *Loaders {
	return &Loaders{
		Organization:         dataloadgen.NewMappedLoader(f.fetchOrganizations),
		Framework:            dataloadgen.NewMappedLoader(f.fetchFrameworks),
		Control:              dataloadgen.NewMappedLoader(f.fetchControls),
		ThirdParty:           dataloadgen.NewMappedLoader(f.fetchThirdParties),
		Document:             dataloadgen.NewMappedLoader(f.fetchDocuments),
		Profile:              dataloadgen.NewMappedLoader(f.fetchProfiles),
		Risk:                 dataloadgen.NewMappedLoader(f.fetchRisks),
		Measure:              dataloadgen.NewMappedLoader(f.fetchMeasures),
		Task:                 dataloadgen.NewMappedLoader(f.fetchTasks),
		File:                 dataloadgen.NewMappedLoader(f.fetchFiles),
		CookieBanner:         dataloadgen.NewMappedLoader(f.fetchCookieBanners),
		CookieCategory:       dataloadgen.NewMappedLoader(f.fetchCookieCategories),
		CommonTrackerPattern: dataloadgen.NewMappedLoader(f.fetchCommonTrackerPatterns),
		CommonThirdParty:     dataloadgen.NewMappedLoader(f.fetchCommonThirdParties),
		Authorize: dataloadgen.NewMappedLoader(
			f.fetchAuthorizes,
			dataloadgen.WithoutCache(),
		),
	}
}

func (f *batchFetcher) fetchOrganizations(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Organization, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	orgs, err := f.probo.Organizations.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load organizations: %w", err)
	}

	result := make(map[gid.GID]*coredata.Organization, len(orgs))
	for _, org := range orgs {
		result[org.ID] = org
	}

	return result, nil
}

func (f *batchFetcher) fetchFrameworks(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Framework, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	frameworks, err := f.probo.Frameworks.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load frameworks: %w", err)
	}

	result := make(map[gid.GID]*coredata.Framework, len(frameworks))
	for _, v := range frameworks {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchControls(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Control, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	controls, err := f.probo.Controls.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load controls: %w", err)
	}

	result := make(map[gid.GID]*coredata.Control, len(controls))
	for _, v := range controls {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchThirdParties(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.ThirdParty, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	thirdParties, err := f.probo.ThirdParties.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load thirdParties: %w", err)
	}

	result := make(map[gid.GID]*coredata.ThirdParty, len(thirdParties))
	for _, v := range thirdParties {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchDocuments(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Document, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	documents, err := f.probo.Documents.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load documents: %w", err)
	}

	result := make(map[gid.GID]*coredata.Document, len(documents))
	for _, v := range documents {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchProfiles(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.MembershipProfile, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	profiles, err := f.iam.OrganizationService.GetProfilesByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load profiles: %w", err)
	}

	result := make(map[gid.GID]*coredata.MembershipProfile, len(profiles))
	for _, v := range profiles {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchRisks(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Risk, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	risks, err := f.probo.Risks.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load risks: %w", err)
	}

	result := make(map[gid.GID]*coredata.Risk, len(risks))
	for _, v := range risks {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchMeasures(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Measure, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	measures, err := f.probo.Measures.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load measures: %w", err)
	}

	result := make(map[gid.GID]*coredata.Measure, len(measures))
	for _, v := range measures {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchTasks(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Task, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	tasks, err := f.probo.Tasks.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load tasks: %w", err)
	}

	result := make(map[gid.GID]*coredata.Task, len(tasks))
	for _, v := range tasks {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchFiles(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.File, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	files, err := f.probo.Files.GetByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load files: %w", err)
	}

	result := make(map[gid.GID]*coredata.File, len(files))
	for _, v := range files {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchCookieBanners(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.CookieBanner, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	banners, err := f.cookieBanner.GetCookieBannersByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load cookie banners: %w", err)
	}

	result := make(map[gid.GID]*coredata.CookieBanner, len(banners))
	for _, v := range banners {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchCookieCategories(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.CookieCategory, error) {
	scope := coredata.NewScopeFromObjectID(keys[0])

	categories, err := f.cookieBanner.GetCookieCategoriesByIDs(ctx, scope, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load cookie categories: %w", err)
	}

	result := make(map[gid.GID]*coredata.CookieCategory, len(categories))
	for _, v := range categories {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchCommonTrackerPatterns(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.CommonTrackerPattern, error) {
	patterns, err := f.cookieBanner.GetCommonTrackerPatternsByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load common tracker patterns: %w", err)
	}

	result := make(map[gid.GID]*coredata.CommonTrackerPattern, len(patterns))
	for _, v := range patterns {
		result[v.ID] = v
	}

	return result, nil
}

func (f *batchFetcher) fetchCommonThirdParties(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.CommonThirdParty, error) {
	parties, err := f.thirdParty.GetCommonThirdPartiesByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load common third parties: %w", err)
	}

	result := make(map[gid.GID]*coredata.CommonThirdParty, len(parties))
	for _, v := range parties {
		result[v.ID] = v
	}

	return result, nil
}

// fetchAuthorizes evaluates the batch with a single AuthorizeMulti call and
// surfaces per-key denials via dataloadgen.MappedFetchError. When
// AuthorizeMulti cannot evaluate the batch as a whole (e.g. mixed
// organizations), we fall back to per-item Authorize so every key still
// gets a scope or iam error.
//
// The Authorize loader is created with WithoutCache() so repeated calls
// with the same (resource, action) within a single request still produce
// one audit log entry per call.
func (f *batchFetcher) fetchAuthorizes(
	ctx context.Context,
	keys []AuthorizeKey,
) (map[AuthorizeKey]AuthorizeResult, error) {
	identity := authn.IdentityFromContext(ctx)
	if identity == nil {
		return nil, fmt.Errorf("cannot authorize without an identity in context")
	}

	session := authn.SessionFromContext(ctx)

	items := make([]iam.MultiAuthorizeItem, 0, len(keys))
	for _, key := range keys {
		attrs, err := decodeAuthorizeKeyAttributes(key.ResourceAttributes)
		if err != nil {
			return nil, fmt.Errorf("cannot decode authorize key attributes: %w", err)
		}

		items = append(items, iam.MultiAuthorizeItem{
			Resource:            key.ResourceID,
			Action:              key.Action,
			ResourceAttributes:  attrs,
			DryRun:              key.DryRun,
			SkipAssumptionCheck: key.SkipAssumptionCheck,
		})
	}

	multiParams := iam.AuthorizeMultiParams{
		Principal: identity.ID,
		Items:     items,
	}
	if session != nil {
		multiParams.Session = &session.ID
	}

	scope, decisions, err := f.iam.Authorizer.AuthorizeMulti(ctx, multiParams)
	if err != nil {
		return f.fetchAuthorizesIndividually(ctx, keys, identity.ID, session)
	}

	result := make(map[AuthorizeKey]AuthorizeResult, len(keys))
	perKeyErrs := make(dataloadgen.MappedFetchError[AuthorizeKey])

	for i, key := range keys {
		if decisions[i] != nil {
			perKeyErrs[key] = decisions[i]
			continue
		}

		result[key] = AuthorizeResult{Scope: scope}
	}

	if len(perKeyErrs) > 0 {
		return result, perKeyErrs
	}

	return result, nil
}

// fetchAuthorizesIndividually is the per-item fallback used when
// AuthorizeMulti cannot evaluate the batch as a whole.
func (f *batchFetcher) fetchAuthorizesIndividually(
	ctx context.Context,
	keys []AuthorizeKey,
	principalID gid.GID,
	session *coredata.Session,
) (map[AuthorizeKey]AuthorizeResult, error) {
	result := make(map[AuthorizeKey]AuthorizeResult, len(keys))
	perKeyErrs := make(dataloadgen.MappedFetchError[AuthorizeKey])

	for _, key := range keys {
		attrs, err := decodeAuthorizeKeyAttributes(key.ResourceAttributes)
		if err != nil {
			return nil, fmt.Errorf("cannot decode authorize key attributes: %w", err)
		}

		params := iam.AuthorizeParams{
			Principal:           principalID,
			Resource:            key.ResourceID,
			Action:              key.Action,
			ResourceAttributes:  make(map[string]string, len(attrs)),
			DryRun:              key.DryRun,
			SkipAssumptionCheck: key.SkipAssumptionCheck,
		}
		maps.Copy(params.ResourceAttributes, attrs)

		if session != nil {
			params.Session = &session.ID
		}

		scope, err := f.iam.Authorizer.Authorize(ctx, params)
		if err != nil {
			perKeyErrs[key] = err
			continue
		}

		result[key] = AuthorizeResult{Scope: scope}
	}

	if len(perKeyErrs) > 0 {
		return result, perKeyErrs
	}

	return result, nil
}

func EncodeAuthorizeKeyAttributes(attrs policy.Attributes) string {
	if len(attrs) == 0 {
		return ""
	}

	b, _ := json.Marshal(attrs)

	return string(b)
}

func decodeAuthorizeKeyAttributes(s string) (policy.Attributes, error) {
	attrs := policy.Attributes{}
	if s == "" {
		return attrs, nil
	}

	if err := json.Unmarshal([]byte(s), &attrs); err != nil {
		return nil, err
	}

	return attrs, nil
}
