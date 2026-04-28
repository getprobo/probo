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

// Command migrate-vendor-snapshots-to-documents creates documents and document
// versions from existing vendor snapshots. For each organization that has vendor
// snapshots, it generates a vendor list document using the same ProseMirror
// builder as the publish flow.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		pgDSN  string
		dryRun bool
	)

	flag.StringVar(
		&pgDSN,
		"pg-dsn",
		os.Getenv("DATABASE_URL"),
		"PostgreSQL connection URL (default: DATABASE_URL env)",
	)
	flag.BoolVar(&dryRun, "dry-run", false, "show what would be done without writing")
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	ctx := context.Background()

	pgClient, err := newPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	return migrate(ctx, pgClient, dryRun)
}

type orgWithVendorSnapshots struct {
	organizationID   gid.GID
	tenantID         gid.TenantID
	organizationName string
}

type vendorSnapshot struct {
	snapshotID  string
	publishedAt time.Time
}

func migrate(ctx context.Context, pgClient *pg.Client, dryRun bool) error {
	var orgs []orgWithVendorSnapshots
	err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error
		orgs, err = loadOrgsWithVendorSnapshots(ctx, conn)
		return err
	})
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("no organizations with vendor snapshots to migrate")
		return nil
	}

	var stats struct {
		documents, versions, failed int
	}

	for _, org := range orgs {
		if dryRun {
			var count int
			err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
				snapshots, err := loadVendorSnapshots(ctx, conn, org.organizationID)
				count = len(snapshots)
				return err
			})
			if err != nil {
				return err
			}
			fmt.Printf("would migrate org %s (%s) — %d vendor snapshot(s)\n",
				org.organizationID, org.organizationName, count)
			continue
		}

		err := pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return migrateOrg(ctx, tx, org)
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL org %s (%s): %v\n",
				org.organizationID, org.organizationName, err)
			stats.failed++
			continue
		}

		stats.documents++
	}

	if dryRun {
		fmt.Printf("\n%d organization(s) would be migrated\n", len(orgs))
		return nil
	}

	fmt.Printf("\nmigrated %d organization(s), %d failed\n",
		stats.documents, stats.failed)

	return nil
}

func migrateOrg(ctx context.Context, tx pg.Tx, org orgWithVendorSnapshots) error {
	snapshots, err := loadVendorSnapshots(ctx, tx, org.organizationID)
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		return nil
	}

	documentID := gid.New(org.tenantID, coredata.DocumentEntityType)
	now := time.Now()

	_, err = tx.Exec(
		ctx,
		`
INSERT INTO documents (
    id, tenant_id, organization_id, write_mode,
    current_published_major, current_published_minor,
    trust_center_visibility, status, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id,
    'GENERATED'::document_write_mode,
    @current_published_major, 0,
    'NONE'::trust_center_visibility,
    'ACTIVE'::document_status,
    @created_at, @updated_at
)`,
		pgx.NamedArgs{
			"id":                      documentID,
			"tenant_id":               org.tenantID,
			"organization_id":         org.organizationID,
			"current_published_major": len(snapshots),
			"created_at":              now,
			"updated_at":              now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot insert document: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO generated_documents (organization_id, tenant_id, vendors_document_id, created_at, updated_at)
VALUES (@organization_id, @tenant_id, @vendors_document_id, @created_at, @updated_at)
ON CONFLICT (organization_id) DO UPDATE SET vendors_document_id = @vendors_document_id, updated_at = @updated_at`,
		pgx.NamedArgs{
			"organization_id":     org.organizationID,
			"tenant_id":           org.tenantID,
			"vendors_document_id": documentID,
			"created_at":          now,
			"updated_at":          now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot link document: %w", err)
	}

	for major, snap := range snapshots {
		content, err := buildSnapshotContent(ctx, tx, snap.snapshotID, org.organizationName, snap.publishedAt)
		if err != nil {
			return fmt.Errorf("cannot build content for snapshot %s: %w", snap.snapshotID, err)
		}

		versionID := gid.New(org.tenantID, coredata.DocumentVersionEntityType)

		_, err = tx.Exec(
			ctx,
			`
INSERT INTO document_versions (
    id, tenant_id, organization_id, document_id,
    title, major, minor, classification, document_type,
    content, changelog, status, orientation,
    pdf_attempt_count,
    published_at, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id, @document_id,
    @title, @major, 0,
    'CONFIDENTIAL'::document_classification,
    'REGISTER'::document_type,
    @content, '',
    'PUBLISHED'::document_version_status,
    'PORTRAIT'::document_version_orientation,
    0,
    @published_at, @published_at, @published_at
)`,
			pgx.NamedArgs{
				"id":              versionID,
				"tenant_id":       org.tenantID,
				"organization_id": org.organizationID,
				"document_id":     documentID,
				"title":           "Vendors",
				"major":           major + 1,
				"content":         content,
				"published_at":    snap.publishedAt,
			},
		)
		if err != nil {
			return fmt.Errorf("cannot insert version for snapshot %s: %w", snap.snapshotID, err)
		}
	}

	fmt.Printf("OK   org %s (%s) — %d version(s)\n",
		org.organizationID, org.organizationName, len(snapshots))

	return nil
}

func loadOrgsWithVendorSnapshots(ctx context.Context, conn pg.Querier) ([]orgWithVendorSnapshots, error) {
	rows, err := conn.Query(
		ctx,
		`
SELECT DISTINCT
    o.id,
    o.tenant_id,
    o.name,
    o.created_at
FROM organizations o
WHERE NOT EXISTS (
        SELECT 1 FROM generated_documents gd
        WHERE gd.organization_id = o.id AND gd.vendors_document_id IS NOT NULL
    )
    AND EXISTS (
        SELECT 1 FROM snapshots s
        WHERE s.organization_id = o.id AND s.type = 'VENDORS'
    )
ORDER BY o.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query organizations with vendor snapshots: %w", err)
	}
	defer rows.Close()

	var result []orgWithVendorSnapshots
	for rows.Next() {
		var o orgWithVendorSnapshots
		var createdAt time.Time
		if err := rows.Scan(&o.organizationID, &o.tenantID, &o.organizationName, &createdAt); err != nil {
			return nil, fmt.Errorf("cannot scan organization: %w", err)
		}
		result = append(result, o)
	}

	return result, rows.Err()
}

func loadVendorSnapshots(ctx context.Context, conn pg.Querier, organizationID gid.GID) ([]vendorSnapshot, error) {
	rows, err := conn.Query(
		ctx,
		`
SELECT DISTINCT
    s.id,
    s.created_at
FROM snapshots s
WHERE s.organization_id = @organization_id
    AND s.type = 'VENDORS'
ORDER BY s.created_at ASC;
`,
		pgx.NamedArgs{"organization_id": organizationID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query vendor snapshots for org %s: %w", organizationID, err)
	}
	defer rows.Close()

	var result []vendorSnapshot
	for rows.Next() {
		var s vendorSnapshot
		if err := rows.Scan(&s.snapshotID, &s.publishedAt); err != nil {
			return nil, fmt.Errorf("cannot scan snapshot: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

type vendorInfo struct {
	id       string
	name     string
	category string

	legalName                     *string
	description                   *string
	headquarterAddress            *string
	websiteURL                    *string
	privacyPolicyURL              *string
	serviceLevelAgreementURL      *string
	dataProcessingAgreementURL    *string
	businessAssociateAgreementURL *string
	subprocessorsListURL          *string
	statusPageURL                 *string
	termsOfServiceURL             *string
	securityPageURL               *string
	trustPageURL                  *string
	certifications                []string
	countries                     []string
	businessOwnerName             string
	securityOwnerName             string
}

func buildSnapshotContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, error) {
	vendorRows, err := tx.Query(
		ctx,
		`
SELECT
    v.id,
    v.name,
    v.category,
    v.legal_name,
    v.description,
    v.headquarter_address,
    v.website_url,
    v.privacy_policy_url,
    v.service_level_agreement_url,
    v.data_processing_agreement_url,
    v.business_associate_agreement_url,
    v.subprocessors_list_url,
    v.status_page_url,
    v.terms_of_service_url,
    v.security_page_url,
    v.trust_page_url,
    v.certifications,
    v.countries,
    COALESCE(bo.full_name, 'Not assigned'),
    COALESCE(so.full_name, 'Not assigned')
FROM vendors v
LEFT JOIN iam_membership_profiles bo ON bo.id = v.business_owner_profile_id
LEFT JOIN iam_membership_profiles so ON so.id = v.security_owner_profile_id
WHERE v.snapshot_id = @snapshot_id
ORDER BY v.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot vendors: %w", err)
	}
	defer vendorRows.Close()

	var vendors []vendorInfo
	for vendorRows.Next() {
		var v vendorInfo
		if err := vendorRows.Scan(
			&v.id, &v.name, &v.category,
			&v.legalName, &v.description, &v.headquarterAddress,
			&v.websiteURL, &v.privacyPolicyURL, &v.serviceLevelAgreementURL,
			&v.dataProcessingAgreementURL, &v.businessAssociateAgreementURL,
			&v.subprocessorsListURL, &v.statusPageURL, &v.termsOfServiceURL,
			&v.securityPageURL, &v.trustPageURL,
			&v.certifications, &v.countries,
			&v.businessOwnerName, &v.securityOwnerName,
		); err != nil {
			return "", fmt.Errorf("cannot scan vendor: %w", err)
		}
		vendors = append(vendors, v)
	}
	if err := vendorRows.Err(); err != nil {
		return "", err
	}

	vendorIDs := make([]string, len(vendors))
	for i, v := range vendors {
		vendorIDs[i] = v.id
	}

	servicesByVendor, err := loadSnapshotServices(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	contactsByVendor, err := loadSnapshotContacts(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	assessmentsByVendor, err := loadSnapshotRiskAssessments(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	reportsByVendor, err := loadSnapshotComplianceReports(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	baaByVendor, err := loadSnapshotBAAs(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	dpaByVendor, err := loadSnapshotDPAs(ctx, tx, snapshotID, vendorIDs)
	if err != nil {
		return "", err
	}

	rows := make([]docgen.VendorListRow, 0, len(vendors))
	for _, v := range vendors {
		row := docgen.VendorListRow{
			Name:                          v.name,
			LegalName:                     deref(v.legalName),
			Description:                   deref(v.description),
			Category:                      formatCategory(v.category),
			HeadquarterAddress:            deref(v.headquarterAddress),
			WebsiteURL:                    deref(v.websiteURL),
			PrivacyPolicyURL:              deref(v.privacyPolicyURL),
			ServiceLevelAgreementURL:      deref(v.serviceLevelAgreementURL),
			DataProcessingAgreementURL:    deref(v.dataProcessingAgreementURL),
			BusinessAssociateAgreementURL: deref(v.businessAssociateAgreementURL),
			SubprocessorsListURL:          deref(v.subprocessorsListURL),
			StatusPageURL:                 deref(v.statusPageURL),
			TermsOfServiceURL:             deref(v.termsOfServiceURL),
			SecurityPageURL:               deref(v.securityPageURL),
			TrustPageURL:                  deref(v.trustPageURL),
			Certifications:                joinOrDefault(v.certifications),
			Countries:                     joinOrDefault(v.countries),
			BusinessOwner:                 v.businessOwnerName,
			SecurityOwner:                 v.securityOwnerName,
			Services:                      servicesByVendor[v.id],
			Contacts:                      contactsByVendor[v.id],
			RiskAssessments:               assessmentsByVendor[v.id],
			ComplianceReports:             reportsByVendor[v.id],
			BusinessAssociateAgreement:    baaByVendor[v.id],
			DataPrivacyAgreement:          dpaByVendor[v.id],
		}
		rows = append(rows, row)
	}

	docData := docgen.VendorListData{
		Title:            "Vendors",
		OrganizationName: orgName,
		CreatedAt:        publishedAt,
		TotalVendors:     len(rows),
		Rows:             rows,
	}

	return probo.BuildVendorListDocument(docData)
}

func loadSnapshotServices(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string][]docgen.VendorListService, error) {
	rows, err := tx.Query(ctx,
		`SELECT vs.vendor_id, vs.name, COALESCE(vs.description, 'Not specified')
		FROM vendor_services vs
		WHERE vs.snapshot_id = @snapshot_id AND vs.vendor_id = ANY(@vendor_ids)
		ORDER BY vs.vendor_id, vs.name ASC`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot services: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]docgen.VendorListService)
	for rows.Next() {
		var vendorID, name, desc string
		if err := rows.Scan(&vendorID, &name, &desc); err != nil {
			return nil, fmt.Errorf("cannot scan service: %w", err)
		}
		result[vendorID] = append(result[vendorID], docgen.VendorListService{Name: name, Description: desc})
	}
	return result, rows.Err()
}

func loadSnapshotContacts(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string][]docgen.VendorListContact, error) {
	rows, err := tx.Query(ctx,
		`SELECT vc.vendor_id,
		    COALESCE(vc.full_name, 'Not specified'),
		    COALESCE(vc.email, 'Not specified'),
		    COALESCE(vc.phone, 'Not specified'),
		    COALESCE(vc.role, 'Not specified')
		FROM vendor_contacts vc
		WHERE vc.snapshot_id = @snapshot_id AND vc.vendor_id = ANY(@vendor_ids)
		ORDER BY vc.vendor_id, vc.full_name ASC`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot contacts: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]docgen.VendorListContact)
	for rows.Next() {
		var vendorID, name, email, phone, role string
		if err := rows.Scan(&vendorID, &name, &email, &phone, &role); err != nil {
			return nil, fmt.Errorf("cannot scan contact: %w", err)
		}
		result[vendorID] = append(result[vendorID], docgen.VendorListContact{
			FullName: name, Email: email, Phone: phone, Role: role,
		})
	}
	return result, rows.Err()
}

func loadSnapshotRiskAssessments(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string][]docgen.VendorListRiskAssessment, error) {
	rows, err := tx.Query(ctx,
		`SELECT vra.vendor_id, vra.created_at, vra.expires_at, vra.data_sensitivity, vra.business_impact, COALESCE(vra.notes, 'Not specified')
		FROM vendor_risk_assessments vra
		WHERE vra.snapshot_id = @snapshot_id AND vra.vendor_id = ANY(@vendor_ids)
		ORDER BY vra.vendor_id, vra.created_at DESC`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot risk assessments: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]docgen.VendorListRiskAssessment)
	for rows.Next() {
		var vendorID, sensitivity, impact, notes string
		var assessedAt, expiresAt time.Time
		if err := rows.Scan(&vendorID, &assessedAt, &expiresAt, &sensitivity, &impact, &notes); err != nil {
			return nil, fmt.Errorf("cannot scan risk assessment: %w", err)
		}
		result[vendorID] = append(result[vendorID], docgen.VendorListRiskAssessment{
			AssessedAt:      assessedAt.Format("2006-01-02"),
			ExpiresAt:       expiresAt.Format("2006-01-02"),
			DataSensitivity: sensitivity,
			BusinessImpact:  impact,
			Notes:           notes,
		})
	}
	return result, rows.Err()
}

func loadSnapshotComplianceReports(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string][]docgen.VendorListComplianceReport, error) {
	rows, err := tx.Query(ctx,
		`SELECT vcr.vendor_id, vcr.report_name, vcr.report_date, vcr.valid_until
		FROM vendor_compliance_reports vcr
		WHERE vcr.snapshot_id = @snapshot_id AND vcr.vendor_id = ANY(@vendor_ids)
		ORDER BY vcr.vendor_id, vcr.report_date DESC`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot compliance reports: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]docgen.VendorListComplianceReport)
	for rows.Next() {
		var vendorID, name string
		var reportDate time.Time
		var validUntil *time.Time
		if err := rows.Scan(&vendorID, &name, &reportDate, &validUntil); err != nil {
			return nil, fmt.Errorf("cannot scan compliance report: %w", err)
		}
		vu := "Not specified"
		if validUntil != nil {
			vu = validUntil.Format("2006-01-02")
		}
		result[vendorID] = append(result[vendorID], docgen.VendorListComplianceReport{
			ReportName: name, ReportDate: reportDate.Format("2006-01-02"), ValidUntil: vu,
		})
	}
	return result, rows.Err()
}

func loadSnapshotBAAs(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string]*docgen.VendorListAgreement, error) {
	rows, err := tx.Query(ctx,
		`SELECT vbaa.vendor_id, vbaa.valid_from, vbaa.valid_until
		FROM vendor_business_associate_agreements vbaa
		WHERE vbaa.snapshot_id = @snapshot_id AND vbaa.vendor_id = ANY(@vendor_ids)`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot BAAs: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*docgen.VendorListAgreement)
	for rows.Next() {
		var vendorID string
		var validFrom, validUntil *time.Time
		if err := rows.Scan(&vendorID, &validFrom, &validUntil); err != nil {
			return nil, fmt.Errorf("cannot scan BAA: %w", err)
		}
		result[vendorID] = &docgen.VendorListAgreement{
			ValidFrom: fmtTime(validFrom), ValidUntil: fmtTime(validUntil),
		}
	}
	return result, rows.Err()
}

func loadSnapshotDPAs(ctx context.Context, tx pg.Tx, snapshotID string, vendorIDs []string) (map[string]*docgen.VendorListAgreement, error) {
	rows, err := tx.Query(ctx,
		`SELECT vdpa.vendor_id, vdpa.valid_from, vdpa.valid_until
		FROM vendor_data_privacy_agreements vdpa
		WHERE vdpa.snapshot_id = @snapshot_id AND vdpa.vendor_id = ANY(@vendor_ids)`,
		pgx.NamedArgs{"snapshot_id": snapshotID, "vendor_ids": vendorIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot DPAs: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*docgen.VendorListAgreement)
	for rows.Next() {
		var vendorID string
		var validFrom, validUntil *time.Time
		if err := rows.Scan(&vendorID, &validFrom, &validUntil); err != nil {
			return nil, fmt.Errorf("cannot scan DPA: %w", err)
		}
		result[vendorID] = &docgen.VendorListAgreement{
			ValidFrom: fmtTime(validFrom), ValidUntil: fmtTime(validUntil),
		}
	}
	return result, rows.Err()
}

func deref(s *string) string {
	if s == nil || *s == "" {
		return "Not specified"
	}
	return *s
}

func joinOrDefault(items []string) string {
	if len(items) == 0 {
		return "Not specified"
	}
	return strings.Join(items, ", ")
}

func fmtTime(t *time.Time) string {
	if t == nil {
		return "Not specified"
	}
	return t.Format("2006-01-02")
}

func formatCategory(c string) string {
	switch c {
	case "ANALYTICS":
		return "Analytics"
	case "CLOUD_MONITORING":
		return "Cloud Monitoring"
	case "CLOUD_PROVIDER":
		return "Cloud Provider"
	case "COLLABORATION":
		return "Collaboration"
	case "CUSTOMER_SUPPORT":
		return "Customer Support"
	case "DATA_STORAGE_AND_PROCESSING":
		return "Data Storage and Processing"
	case "DOCUMENT_MANAGEMENT":
		return "Document Management"
	case "EMPLOYEE_MANAGEMENT":
		return "Employee Management"
	case "ENGINEERING":
		return "Engineering"
	case "FINANCE":
		return "Finance"
	case "IDENTITY_PROVIDER":
		return "Identity Provider"
	case "IT":
		return "IT"
	case "MARKETING":
		return "Marketing"
	case "OFFICE_OPERATIONS":
		return "Office Operations"
	case "OTHER":
		return "Other"
	case "PASSWORD_MANAGEMENT":
		return "Password Management"
	case "PRODUCT_AND_DESIGN":
		return "Product and Design"
	case "PROFESSIONAL_SERVICES":
		return "Professional Services"
	case "RECRUITING":
		return "Recruiting"
	case "SALES":
		return "Sales"
	case "SECURITY":
		return "Security"
	case "VERSION_CONTROL":
		return "Version Control"
	default:
		return c
	}
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN")
	}

	var opts []pg.Option

	if u.Host != "" {
		opts = append(opts, pg.WithAddr(u.Host))
	}

	if u.User != nil {
		opts = append(opts, pg.WithUser(u.User.Username()))
		if password, ok := u.User.Password(); ok {
			opts = append(opts, pg.WithPassword(password))
		}
	}

	if len(u.Path) > 1 {
		opts = append(opts, pg.WithDatabase(u.Path[1:]))
	}

	return pg.NewClient(opts...)
}
