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

package trust

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

type (
	compliancePageData struct {
		OrgName       string
		Description   string
		Details       []compliancePageDetail
		Frameworks    []compliancePageFramework
		Documents     []compliancePageDocument
		Audits        []compliancePageAudit
		Vendors       []compliancePageVendor
		References    []compliancePageReference
		ExternalLinks []compliancePageExternalLink
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

	compliancePageVendor struct {
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

	compliancePageExternalLink struct {
		Name string
		URL  string
	}
)

func (s *Service) RenderCompliancePageMarkdown(
	ctx context.Context,
	w io.Writer,
	trustCenterID gid.GID,
	tenantID gid.TenantID,
) error {
	org, err := s.GetOrganizationByTrustCenterID(ctx, trustCenterID)
	if err != nil {
		return fmt.Errorf("cannot load organization for compliance page: %w", err)
	}

	tenantSvc := s.WithTenant(tenantID)

	data := &compliancePageData{
		OrgName: org.Name,
	}

	if org.Description != nil && *org.Description != "" {
		data.Description = *org.Description
	}

	if org.WebsiteURL != nil && *org.WebsiteURL != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Website", Value: *org.WebsiteURL})
	}
	if org.Email != nil && *org.Email != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Email", Value: *org.Email})
	}
	if org.HeadquarterAddress != nil && *org.HeadquarterAddress != "" {
		data.Details = append(data.Details, compliancePageDetail{Label: "Headquarters", Value: *org.HeadquarterAddress})
	}

	data.Frameworks = s.fetchComplianceFrameworks(ctx, tenantSvc, trustCenterID)
	data.Documents = s.fetchDocuments(ctx, tenantSvc, org.ID)
	data.Audits = s.fetchAudits(ctx, tenantSvc, org.ID)
	data.Vendors = s.fetchVendors(ctx, tenantSvc, org.ID)
	data.References = s.fetchReferences(ctx, tenantSvc, trustCenterID)
	data.ExternalLinks = s.fetchExternalLinks(ctx, tenantSvc, trustCenterID)

	if err := complianceTmpl.Execute(w, data); err != nil {
		return fmt.Errorf("cannot render compliance page markdown: %w", err)
	}

	return nil
}

func (s *Service) fetchComplianceFrameworks(ctx context.Context, tenantSvc *TenantService, trustCenterID gid.GID) []compliancePageFramework {
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

		result, err := tenantSvc.ComplianceFrameworks.ListByTrustCenterID(ctx, trustCenterID, cursor)
		if err != nil {
			break
		}

		for _, cf := range result.Data {
			if cf.Visibility != coredata.ComplianceFrameworkVisibilityPublic {
				continue
			}

			fw, err := tenantSvc.Frameworks.Get(ctx, cf.FrameworkID)
			if err != nil {
				continue
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

	return frameworks
}

func (s *Service) fetchDocuments(ctx context.Context, tenantSvc *TenantService, orgID gid.GID) []compliancePageDocument {
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

		result, err := tenantSvc.Documents.ListForOrganizationId(ctx, orgID, cursor)
		if err != nil {
			break
		}

		for _, doc := range result.Data {
			if doc.TrustCenterVisibility != coredata.TrustCenterVisibilityPublic {
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

	return docs
}

func (s *Service) fetchAudits(ctx context.Context, tenantSvc *TenantService, orgID gid.GID) []compliancePageAudit {
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

		result, err := tenantSvc.Audits.ListForOrganizationId(ctx, orgID, cursor)
		if err != nil {
			break
		}

		for _, audit := range result.Data {
			if audit.TrustCenterVisibility != coredata.TrustCenterVisibilityPublic {
				continue
			}

			frameworkName := ""
			fw, err := tenantSvc.Frameworks.Get(ctx, audit.FrameworkID)
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

	return audits
}

func (s *Service) fetchVendors(ctx context.Context, tenantSvc *TenantService, orgID gid.GID) []compliancePageVendor {
	var vendors []compliancePageVendor

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.VendorOrderField]{
				Field:     coredata.VendorOrderFieldName,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := tenantSvc.Vendors.ListForOrganizationId(ctx, orgID, cursor)
		if err != nil {
			break
		}

		for _, v := range result.Data {
			var countries []string
			for _, c := range v.Countries {
				countries = append(countries, c.String())
			}

			vendors = append(
				vendors,
				compliancePageVendor{
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
		ck := last.CursorKey(coredata.VendorOrderFieldName)
		cursorKey = &ck
	}

	return vendors
}

func (s *Service) fetchReferences(ctx context.Context, tenantSvc *TenantService, trustCenterID gid.GID) []compliancePageReference {
	var refs []compliancePageReference

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.TrustCenterReferenceOrderField]{
				Field:     coredata.TrustCenterReferenceOrderFieldRank,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := tenantSvc.TrustCenterReferences.ListForTrustCenterID(ctx, trustCenterID, cursor)
		if err != nil {
			break
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
		ck := last.CursorKey(coredata.TrustCenterReferenceOrderFieldRank)
		cursorKey = &ck
	}

	return refs
}

func (s *Service) fetchExternalLinks(ctx context.Context, tenantSvc *TenantService, trustCenterID gid.GID) []compliancePageExternalLink {
	var links []compliancePageExternalLink

	var cursorKey *page.CursorKey
	for {
		cursor := page.NewCursor(
			page.MaxCursorSize,
			cursorKey,
			page.Head,
			page.OrderBy[coredata.ComplianceExternalURLOrderField]{
				Field:     coredata.ComplianceExternalURLOrderFieldRank,
				Direction: page.OrderDirectionAsc,
			},
		)

		result, err := tenantSvc.ComplianceExternalURLs.ListForTrustCenterID(ctx, trustCenterID, cursor)
		if err != nil {
			break
		}

		for _, l := range result.Data {
			links = append(
				links,
				compliancePageExternalLink{
					Name: l.Name,
					URL:  l.URL,
				},
			)
		}

		if !result.Info.HasNext {
			break
		}

		last := result.Data[len(result.Data)-1]
		ck := last.CursorKey(coredata.ComplianceExternalURLOrderFieldRank)
		cursorKey = &ck
	}

	return links
}
