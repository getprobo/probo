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

package visitor

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

//go:embed compliance.md.tmpl
var complianceTmplContent string

//go:embed sitemap.xml.tmpl
var sitemapTmplContent string

//go:embed robots.txt.tmpl
var robotsTmplContent string

var complianceTmpl = template.Must(
	template.New("compliance").
		Funcs(template.FuncMap{
			"cell": func(s string) string {
				s = strings.ReplaceAll(s, `|`, `\|`)
				s = strings.ReplaceAll(s, "\n", " ")
				s = strings.ReplaceAll(s, "\r", "")

				return s
			},
		}).
		Parse(complianceTmplContent),
)

var sitemapTmpl = template.Must(
	template.New("sitemap").Parse(sitemapTmplContent),
)

var robotsTmpl = template.Must(
	template.New("robots").Parse(robotsTmplContent),
)

type (
	compliancePageData struct {
		OrgName      string
		Description  string
		Details      []compliancePageDetail
		Frameworks   []compliancePageFramework
		Documents    []compliancePageDocument
		Audits       []compliancePageAudit
		ThirdParties []compliancePageThirdParty
		References   []compliancePageReference
		CustomLinks  []compliancePageCustomLink
	}

	compliancePageDetail struct {
		Label string
		Value string
	}

	compliancePageFramework struct {
		Name        string
		Description string
	}

	compliancePageDocument struct {
		Title string
		Type  string
	}

	compliancePageAudit struct {
		Name       string
		Framework  string
		ValidFrom  string
		ValidUntil string
	}

	compliancePageThirdParty struct {
		Name      string
		Category  string
		Countries string
		Website   string
	}

	compliancePageReference struct {
		Name        string
		Description string
		Website     string
	}

	compliancePageCustomLink struct {
		Name string
		URL  string
	}
)

func (s *Service) RenderCompliancePortalMarkdown(
	ctx context.Context,
	w io.Writer,
	compliancePageID gid.GID,
	scope coredata.Scoper,
) error {
	org, err := s.GetPortalOrganization(ctx, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot load organization for compliance page: %w", err)
	}

	compliancePage, err := s.GetPortalByID(ctx, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot load compliance page: %w", err)
	}

	data := &compliancePageData{
		OrgName: org.Name,
	}

	if compliancePage.Description != nil && *compliancePage.Description != "" {
		data.Description = *compliancePage.Description
	}

	if compliancePage.WebsiteURL != nil && *compliancePage.WebsiteURL != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Website", Value: *compliancePage.WebsiteURL})
	}

	if compliancePage.Email != nil && *compliancePage.Email != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Email", Value: *compliancePage.Email})
	}

	if compliancePage.HeadquarterAddress != nil && *compliancePage.HeadquarterAddress != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Headquarters", Value: *compliancePage.HeadquarterAddress})
	}

	data.Frameworks, err = s.fetchComplianceFrameworks(ctx, scope, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot fetch compliance frameworks: %w", err)
	}

	data.Documents, err = s.fetchDocuments(ctx, scope, org.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch documents: %w", err)
	}

	data.Audits, err = s.fetchAudits(ctx, scope, org.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch audits: %w", err)
	}

	data.ThirdParties, err = s.fetchThirdParties(ctx, scope, org.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch thirdParties: %w", err)
	}

	data.References, err = s.fetchReferences(ctx, scope, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot fetch references: %w", err)
	}

	data.CustomLinks, err = s.fetchCustomLinks(ctx, scope, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot fetch external links: %w", err)
	}

	if err := complianceTmpl.Execute(w, data); err != nil {
		return fmt.Errorf("cannot render compliance page markdown: %w", err)
	}

	return nil
}

type (
	sitemapData struct {
		BaseURL   string
		Documents []string
	}

	robotsData struct {
		Indexable bool
		BaseURL   string
	}
)

func (s *Service) RenderSitemap(
	ctx context.Context,
	w io.Writer,
	compliancePageID gid.GID,
	scope coredata.Scoper,
	baseURL string,
) error {
	org, err := s.GetPortalOrganization(ctx, compliancePageID)
	if err != nil {
		return fmt.Errorf("cannot load organization for sitemap: %w", err)
	}

	data := &sitemapData{
		BaseURL: baseURL,
	}

	data.Documents, err = s.fetchDocumentIDs(ctx, scope, org.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch document IDs for sitemap: %w", err)
	}

	if err := sitemapTmpl.Execute(w, data); err != nil {
		return fmt.Errorf("cannot render sitemap: %w", err)
	}

	return nil
}

func (s *Service) RenderRobotsTxt(
	ctx context.Context,
	w io.Writer,
	searchEngineIndexing coredata.SearchEngineIndexing,
	baseURL string,
) error {
	data := &robotsData{
		Indexable: searchEngineIndexing == coredata.SearchEngineIndexingIndexable,
		BaseURL:   baseURL,
	}

	if err := robotsTmpl.Execute(w, data); err != nil {
		return fmt.Errorf("cannot render robots.txt: %w", err)
	}

	return nil
}

func (s *Service) fetchDocumentIDs(
	ctx context.Context,
	scope coredata.Scoper,
	orgID gid.GID,
) ([]string, error) {
	seen := make(map[gid.GID]struct{})

	var resourceIDs []gid.GID

	appendResourceID := func(id gid.GID) {
		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}
		resourceIDs = append(resourceIDs, id)
	}

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.DocumentOrderField]{
				Field:     coredata.DocumentOrderFieldTitle,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListDocumentsForOrganizationID(ctx, scope, orgID, cursor, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot list documents: %w", err)
		}

		for _, doc := range result.Data {
			if doc.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			appendResourceID(doc.ID)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.DocumentOrderFieldTitle)
		cursorKey = &ck
	}

	cursorKey = nil
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.CompliancePortalFileOrderField]{
				Field:     coredata.CompliancePortalFileOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListPortalFilesForOrganizationID(
			ctx,
			scope,
			orgID,
			cursor,
			coredata.NewCompliancePortalFileFilter(),
		)
		if err != nil {
			return nil, fmt.Errorf("cannot list compliance page files: %w", err)
		}

		for _, file := range result.Data {
			if file.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			appendResourceID(file.ID)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.CompliancePortalFileOrderFieldCreatedAt)
		cursorKey = &ck
	}

	cursorKey = nil
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.AuditOrderField]{
				Field:     coredata.AuditOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListAuditsForOrganizationID(ctx, scope, orgID, cursor, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot list audits: %w", err)
		}

		for _, audit := range result.Data {
			if audit.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			if audit.ReportFileID == nil {
				continue
			}

			appendResourceID(*audit.ReportFileID)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.AuditOrderFieldCreatedAt)
		cursorKey = &ck
	}

	aliases, err := s.resourceAlias.LoadByResourceIDs(ctx, scope, resourceIDs)
	if err != nil {
		return nil, fmt.Errorf("cannot load resource aliases: %w", err)
	}

	paths := make([]string, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		if alias, ok := aliases[resourceID]; ok {
			paths = append(paths, alias)
			continue
		}

		paths = append(paths, resourceID.String())
	}

	return paths, nil
}

func (s *Service) fetchComplianceFrameworks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) ([]compliancePageFramework, error) {
	var frameworks []compliancePageFramework

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.ComplianceFrameworkOrderField]{
				Field:     coredata.ComplianceFrameworkOrderFieldRank,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListComplianceFrameworksByPortalID(ctx, scope, compliancePageID, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot list compliance frameworks: %w", err)
		}

		for _, cf := range result.Data {
			if cf.Visibility != coredata.ComplianceFrameworkVisibilityPublic {
				continue
			}

			fw, err := s.GetFramework(ctx, scope, cf.FrameworkID)
			if err != nil {
				return nil, fmt.Errorf("cannot get framework %s: %w", cf.FrameworkID, err)
			}

			fi := compliancePageFramework{Name: fw.Name}
			if fw.Description != nil {
				fi.Description = *fw.Description
			}

			frameworks = append(frameworks, fi)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.ComplianceFrameworkOrderFieldRank)
		cursorKey = &ck
	}

	return frameworks, nil
}

func (s *Service) fetchDocuments(
	ctx context.Context,
	scope coredata.Scoper,
	orgID gid.GID,
) ([]compliancePageDocument, error) {
	var docs []compliancePageDocument

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.DocumentOrderField]{
				Field:     coredata.DocumentOrderFieldTitle,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListDocumentsForOrganizationID(ctx, scope, orgID, cursor, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot list documents: %w", err)
		}

		for _, doc := range result.Data {
			if doc.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			docs = append(
				docs,
				compliancePageDocument{
					Title: doc.Title,
					Type:  doc.DocumentType.String(),
				},
			)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.DocumentOrderFieldTitle)
		cursorKey = &ck
	}

	return docs, nil
}

func (s *Service) fetchAudits(
	ctx context.Context,
	scope coredata.Scoper,
	orgID gid.GID,
) ([]compliancePageAudit, error) {
	var audits []compliancePageAudit

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.AuditOrderField]{
				Field:     coredata.AuditOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListAuditsForOrganizationID(ctx, scope, orgID, cursor, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot list audits: %w", err)
		}

		for _, audit := range result.Data {
			if audit.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
				continue
			}

			frameworkName := ""

			fw, err := s.GetFramework(ctx, scope, audit.FrameworkID)
			if err == nil {
				frameworkName = fw.Name
			}

			ai := compliancePageAudit{
				Name:      ref.UnrefOrZero(audit.Name),
				Framework: frameworkName,
			}
			if audit.ValidFrom != nil {
				ai.ValidFrom = audit.ValidFrom.Format("2006-01-02")
			}

			if audit.ValidUntil != nil {
				ai.ValidUntil = audit.ValidUntil.Format("2006-01-02")
			}

			audits = append(audits, ai)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.AuditOrderFieldCreatedAt)
		cursorKey = &ck
	}

	return audits, nil
}

func (s *Service) fetchThirdParties(
	ctx context.Context,
	scope coredata.Scoper,
	orgID gid.GID,
) ([]compliancePageThirdParty, error) {
	var thirdParties []compliancePageThirdParty

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.ThirdPartyOrderField]{
				Field:     coredata.ThirdPartyOrderFieldName,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListThirdPartiesForOrganizationID(ctx, scope, orgID, cursor, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot list thirdParties: %w", err)
		}

		for _, v := range result.Data {
			var countries []string
			for _, c := range v.Countries {
				countries = append(countries, c.String())
			}

			thirdParties = append(
				thirdParties,
				compliancePageThirdParty{
					Name:      v.Name,
					Category:  v.Category.String(),
					Countries: strings.Join(countries, ", "),
					Website:   ref.UnrefOrZero(v.WebsiteURL),
				},
			)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.ThirdPartyOrderFieldName)
		cursorKey = &ck
	}

	return thirdParties, nil
}

func (s *Service) fetchReferences(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) ([]compliancePageReference, error) {
	var refs []compliancePageReference

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.CompliancePortalReferenceOrderField]{
				Field:     coredata.CompliancePortalReferenceOrderFieldRank,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListPortalReferencesForPortalID(ctx, scope, compliancePageID, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot list references: %w", err)
		}

		for _, r := range result.Data {
			ri := compliancePageReference{
				Name:    r.Name,
				Website: r.WebsiteURL,
			}
			if r.Description != nil {
				ri.Description = *r.Description
			}

			refs = append(refs, ri)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.CompliancePortalReferenceOrderFieldRank)
		cursorKey = &ck
	}

	return refs, nil
}

func (s *Service) fetchCustomLinks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) ([]compliancePageCustomLink, error) {
	var links []compliancePageCustomLink

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.ComplianceCustomLinkOrderField]{
				Field:     coredata.ComplianceCustomLinkOrderFieldRank,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := s.ListCustomLinksForPortalID(ctx, scope, compliancePageID, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot list custom links: %w", err)
		}

		for _, l := range result.Data {
			links = append(
				links,
				compliancePageCustomLink{
					Name: l.Name,
					URL:  l.URL,
				},
			)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.ComplianceCustomLinkOrderFieldRank)
		cursorKey = &ck
	}

	return links, nil
}
