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

package probo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

var trackerPolicyTemplate = template.Must(
	template.New("tracker_policy.md.tmpl").
		Funcs(template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("January 2, 2006")
			},
		}).
		ParseFS(Templates, "templates/tracker_policy.md.tmpl"),
)

// BuildTrackerPolicyDocument renders the tracker policy markdown template for
// the given data and converts it into the ProseMirror JSON expected by
// DocumentVersion.Content.
func BuildTrackerPolicyDocument(data docgen.TrackerPolicyData) (string, error) {
	var buf bytes.Buffer
	if err := trackerPolicyTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute tracker policy template: %w", err)
	}

	node, err := prosemirror.ParseMarkdown(buf.String())
	if err != nil {
		return "", fmt.Errorf("cannot convert tracker policy markdown: %w", err)
	}

	out, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal tracker policy prosemirror node: %w", err)
	}

	return string(out), nil
}

// PublishTrackerPolicy generates (or regenerates) the cookie and tracking
// technologies policy document for a banner from its latest published version
// snapshot. The document is stored as a GENERATED document that is PRIVATE in
// the trust center by default, and is linked to the banner through
// cookie_banners.policy_document_id.
func (s *GeneratedDocumentService) PublishTrackerPolicy(
	ctx context.Context,
	scope coredata.Scoper,
	cookieBannerID gid.GID,
) error {
	// Phase 1: collect data and render the prosemirror document outside any
	// write transaction. Template rendering, markdown parsing, and JSON
	// marshaling are CPU work that should not hold a write transaction open.
	var (
		organizationID gid.GID
		bannerOrigin   string
		documentData   docgen.TrackerPolicyData
	)

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		banner := &coredata.CookieBanner{}
		if err := banner.LoadByID(ctx, conn, scope, cookieBannerID); err != nil {
			return fmt.Errorf("cannot load cookie banner: %w", err)
		}

		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, conn, scope, banner.OrganizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		var err error

		documentData, err = s.buildTrackerPolicyDocumentData(ctx, scope, conn, organization, banner)
		if err != nil {
			return fmt.Errorf("cannot build document data: %w", err)
		}

		organizationID = banner.OrganizationID
		bannerOrigin = banner.Origin

		return nil
	})
	if err != nil {
		return err
	}

	prosemirrorJSON, err := BuildTrackerPolicyDocument(documentData)
	if err != nil {
		return fmt.Errorf("cannot build prosemirror document: %w", err)
	}

	// Phase 2: persist the document and version in a write transaction.
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			banner := &coredata.CookieBanner{}
			if err := banner.LoadByID(ctx, tx, scope, cookieBannerID); err != nil {
				return fmt.Errorf("cannot reload cookie banner: %w", err)
			}

			now := time.Now()

			var (
				document    *coredata.Document
				existingDoc *coredata.Document
			)

			if banner.PolicyDocumentID != nil {
				doc := &coredata.Document{}

				err = doc.LoadByID(ctx, tx, scope, *banner.PolicyDocumentID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load tracker policy document: %w", err)
				}

				if err == nil && doc.ArchivedAt == nil {
					existingDoc = doc
				} else {
					banner.PolicyDocumentID = nil
					banner.UpdatedAt = now

					if err := banner.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot clear tracker policy document reference: %w", err)
					}
				}
			}

			if existingDoc == nil {
				documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)

				document = &coredata.Document{
					ID:                    documentID,
					OrganizationID:        organizationID,
					WriteMode:             coredata.DocumentWriteModeGenerated,
					TrustCenterVisibility: coredata.TrustCenterVisibilityPrivate,
					Status:                coredata.DocumentStatusActive,
					CreatedAt:             now,
					UpdatedAt:             now,
				}

				if err := document.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document: %w", err)
				}

				banner.PolicyDocumentID = &documentID
				banner.UpdatedAt = now

				if err := banner.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update tracker policy document reference: %w", err)
				}
			} else {
				document = existingDoc
			}

			documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)
			documentVersion := &coredata.DocumentVersion{
				ID:             documentVersionID,
				OrganizationID: organizationID,
				DocumentID:     document.ID,
				Title:          fmt.Sprintf("Cookie and Tracking Technologies Policy — %s", bannerOrigin),
				Content:        prosemirrorJSON,
				Classification: coredata.DocumentClassificationPublic,
				DocumentType:   coredata.DocumentTypePolicy,
				Orientation:    coredata.DocumentVersionOrientationPortrait,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return s.publishOrRequestApproval(ctx, scope, tx, document, documentVersion, organizationID, nil, false, now)
		},
	)
}

func (s *GeneratedDocumentService) buildTrackerPolicyDocumentData(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	organization *coredata.Organization,
	banner *coredata.CookieBanner,
) (docgen.TrackerPolicyData, error) {
	version := &coredata.CookieBannerVersion{}
	if err := version.LoadLatestPublishedByCookieBannerID(ctx, conn, scope, banner.ID); err != nil {
		return docgen.TrackerPolicyData{}, fmt.Errorf("cannot load latest published version: %w", err)
	}

	snapshot, err := version.GetSnapshot()
	if err != nil {
		return docgen.TrackerPolicyData{}, fmt.Errorf("cannot decode snapshot: %w", err)
	}

	categories := make([]docgen.TrackerPolicyCategory, 0, len(snapshot.Categories))
	for _, c := range snapshot.Categories {
		trackers := make([]docgen.TrackerPolicyTracker, 0, len(c.Cookies))
		for _, cookie := range c.Cookies {
			trackers = append(trackers, docgen.TrackerPolicyTracker{
				Name:     sanitizeTrackerCell(cookie.Name),
				Type:     cookie.TrackerType.Label(),
				Purpose:  trackerPurpose(cookie.Description),
				Duration: cookie.HumanizedDuration(),
			})
		}

		categories = append(categories, docgen.TrackerPolicyCategory{
			Name:        strings.TrimSpace(c.Name),
			Description: strings.TrimSpace(c.Description),
			Necessary:   c.Kind == coredata.CookieCategoryKindNecessary,
			Trackers:    trackers,
		})
	}

	thirdParties, err := s.buildTrackerPolicyThirdParties(ctx, scope, conn, banner.ID)
	if err != nil {
		return docgen.TrackerPolicyData{}, err
	}

	privacyPolicyURL := ""
	if snapshot.PrivacyPolicyURL != nil {
		privacyPolicyURL = strings.TrimSpace(*snapshot.PrivacyPolicyURL)
	}

	return docgen.TrackerPolicyData{
		OrganizationName:  organization.Name,
		WebsiteOrigin:     banner.Origin,
		GeneratedAt:       time.Now(),
		PrivacyPolicyURL:  privacyPolicyURL,
		ConsentExpiryDays: snapshot.ConsentExpiryDays,
		Categories:        categories,
		ThirdParties:      thirdParties,
	}, nil
}

func (s *GeneratedDocumentService) buildTrackerPolicyThirdParties(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	cookieBannerID gid.GID,
) ([]docgen.TrackerPolicyThirdParty, error) {
	var patterns coredata.TrackerPatterns

	thirdPartyIDs, err := patterns.LoadDistinctThirdPartyIDsByCookieBannerID(ctx, conn, scope, cookieBannerID)
	if err != nil {
		return nil, fmt.Errorf("cannot load distinct third party ids: %w", err)
	}

	if len(thirdPartyIDs) == 0 {
		return nil, nil
	}

	var thirdParties coredata.ThirdParties
	if err := thirdParties.LoadByIDs(ctx, conn, scope, thirdPartyIDs); err != nil {
		return nil, fmt.Errorf("cannot load third parties: %w", err)
	}

	rows := make([]docgen.TrackerPolicyThirdParty, 0, len(thirdParties))
	for _, tp := range thirdParties {
		row := docgen.TrackerPolicyThirdParty{Name: strings.TrimSpace(tp.Name)}

		if tp.Description != nil {
			row.Description = collapseWhitespace(*tp.Description)
		}

		if tp.PrivacyPolicyURL != nil {
			row.PrivacyPolicyURL = strings.TrimSpace(*tp.PrivacyPolicyURL)
		}

		rows = append(rows, row)
	}

	// LoadByIDs returns rows in an unspecified order (no ORDER BY), so sort
	// here to keep the generated policy document deterministic across
	// regenerations.
	slices.SortFunc(rows, func(a, b docgen.TrackerPolicyThirdParty) int {
		if c := strings.Compare(a.Name, b.Name); c != 0 {
			return c
		}

		if c := strings.Compare(a.Description, b.Description); c != 0 {
			return c
		}

		return strings.Compare(a.PrivacyPolicyURL, b.PrivacyPolicyURL)
	})

	return rows, nil
}

// trackerPurpose returns a table-safe purpose string for a tracker, falling
// back to a neutral label when no enriched description is available.
func trackerPurpose(description string) string {
	cell := sanitizeTrackerCell(description)
	if cell == "" {
		return "Not specified"
	}

	return cell
}

// sanitizeTrackerCell makes free-form text safe to embed in a markdown table
// cell: it collapses whitespace (including newlines) and escapes pipe
// characters so they do not break the column layout.
func sanitizeTrackerCell(s string) string {
	return strings.ReplaceAll(collapseWhitespace(s), "|", "\\|")
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
