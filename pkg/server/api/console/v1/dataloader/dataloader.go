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
	"fmt"
	"net/http"

	"github.com/vikstrous/dataloadgen"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
)

type (
	ctxKey struct{ name string }

	Loaders struct {
		Organization *dataloadgen.Loader[gid.GID, *coredata.Organization]
		Framework    *dataloadgen.Loader[gid.GID, *coredata.Framework]
		Control      *dataloadgen.Loader[gid.GID, *coredata.Control]
		Vendor       *dataloadgen.Loader[gid.GID, *coredata.Vendor]
		Document     *dataloadgen.Loader[gid.GID, *coredata.Document]
		Profile      *dataloadgen.Loader[gid.GID, *coredata.MembershipProfile]
		Risk         *dataloadgen.Loader[gid.GID, *coredata.Risk]
		Measure      *dataloadgen.Loader[gid.GID, *coredata.Measure]
		Task         *dataloadgen.Loader[gid.GID, *coredata.Task]
		File         *dataloadgen.Loader[gid.GID, *coredata.File]
		Report       *dataloadgen.Loader[gid.GID, *coredata.Report]
	}

	batchFetcher struct {
		probo *probo.Service
		iam   *iam.Service
	}
)

var loadersKey = &ctxKey{name: "dataloaders"}

func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func NewMiddleware(proboSvc *probo.Service, iamSvc *iam.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				f := &batchFetcher{probo: proboSvc, iam: iamSvc}
				loaders := f.newLoaders()
				ctx := context.WithValue(r.Context(), loadersKey, loaders)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func (f *batchFetcher) newLoaders() *Loaders {
	return &Loaders{
		Organization: dataloadgen.NewMappedLoader(f.fetchOrganizations),
		Framework:    dataloadgen.NewMappedLoader(f.fetchFrameworks),
		Control:      dataloadgen.NewMappedLoader(f.fetchControls),
		Vendor:       dataloadgen.NewMappedLoader(f.fetchVendors),
		Document:     dataloadgen.NewMappedLoader(f.fetchDocuments),
		Profile:      dataloadgen.NewMappedLoader(f.fetchProfiles),
		Risk:         dataloadgen.NewMappedLoader(f.fetchRisks),
		Measure:      dataloadgen.NewMappedLoader(f.fetchMeasures),
		Task:         dataloadgen.NewMappedLoader(f.fetchTasks),
		File:         dataloadgen.NewMappedLoader(f.fetchFiles),
		Report:       dataloadgen.NewMappedLoader(f.fetchReports),
	}
}

func (f *batchFetcher) fetchOrganizations(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Organization, error) {
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	orgs, err := tenantSvc.Organizations.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	frameworks, err := tenantSvc.Frameworks.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	controls, err := tenantSvc.Controls.GetByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load controls: %w", err)
	}

	result := make(map[gid.GID]*coredata.Control, len(controls))
	for _, v := range controls {
		result[v.ID] = v
	}
	return result, nil
}

func (f *batchFetcher) fetchVendors(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Vendor, error) {
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	vendors, err := tenantSvc.Vendors.GetByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load vendors: %w", err)
	}

	result := make(map[gid.GID]*coredata.Vendor, len(vendors))
	for _, v := range vendors {
		result[v.ID] = v
	}
	return result, nil
}

func (f *batchFetcher) fetchDocuments(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Document, error) {
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	documents, err := tenantSvc.Documents.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	risks, err := tenantSvc.Risks.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	measures, err := tenantSvc.Measures.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	tasks, err := tenantSvc.Tasks.GetByIDs(ctx, keys...)
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
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	files, err := tenantSvc.Files.GetByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load files: %w", err)
	}

	result := make(map[gid.GID]*coredata.File, len(files))
	for _, v := range files {
		result[v.ID] = v
	}
	return result, nil
}

func (f *batchFetcher) fetchReports(ctx context.Context, keys []gid.GID) (map[gid.GID]*coredata.Report, error) {
	tenantSvc := f.probo.WithTenant(keys[0].TenantID())

	reports, err := tenantSvc.Reports.GetByIDs(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load reports: %w", err)
	}

	result := make(map[gid.GID]*coredata.Report, len(reports))
	for _, v := range reports {
		result[v.ID] = v
	}
	return result, nil
}
