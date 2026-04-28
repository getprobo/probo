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

// Command migrate-processing-activity-snapshots-to-documents creates documents
// and document versions from existing processing activity snapshots. For each
// organization that has PROCESSING_ACTIVITIES snapshots, it produces three
// register documents — Processing Activities, Data Protection Impact
// Assessments, and Transfer Impact Assessments — using the same ProseMirror
// builders as the publish flow, with one version per snapshot ordered by date.
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

	return pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return migrate(ctx, tx, dryRun)
	})
}

type orgWithSnapshots struct {
	organizationID   gid.GID
	tenantID         gid.TenantID
	organizationName string
}

type processingActivitySnapshot struct {
	snapshotID  string
	publishedAt time.Time
}

type kind struct {
	name    string
	title   string
	column  string
	buildFn func(ctx context.Context, tx pg.Tx, snapshotID string, orgName string, publishedAt time.Time) (string, int, error)
}

func migrate(ctx context.Context, tx pg.Tx, dryRun bool) error {
	orgs, err := loadOrgsWithSnapshots(ctx, tx)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("no organizations with processing activity snapshots to migrate")
		return nil
	}

	kinds := []kind{
		{name: "processing-activity", title: "Processing Activities", column: "processing_activities_document_id", buildFn: buildProcessingActivityContent},
		{name: "dpia", title: "Data Protection Impact Assessments", column: "data_protection_impact_assessments_document_id", buildFn: buildDPIAContent},
		{name: "tia", title: "Transfer Impact Assessments", column: "transfer_impact_assessments_document_id", buildFn: buildTIAContent},
	}

	var stats struct {
		documents, versions int
	}

	for _, org := range orgs {
		snapshots, err := loadSnapshots(ctx, tx, org.organizationID)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("would migrate org %s (%s) — %d snapshot(s) × %d kind(s)\n",
				org.organizationID, org.organizationName, len(snapshots), len(kinds))
			continue
		}

		for _, k := range kinds {
			documentID := gid.New(org.tenantID, coredata.DocumentEntityType)
			now := time.Now()

			versionsInserted := 0
			for major, snap := range snapshots {
				content, count, err := k.buildFn(ctx, tx, snap.snapshotID, org.organizationName, snap.publishedAt)
				if err != nil {
					return fmt.Errorf("cannot build %s content for snapshot %s of org %s: %w",
						k.name, snap.snapshotID, org.organizationID, err)
				}
				if count == 0 {
					continue
				}

				versionID := gid.New(org.tenantID, coredata.DocumentVersionEntityType)

				_, err = tx.Exec(
					ctx,
					`
INSERT INTO document_versions (
    id, tenant_id, organization_id, document_id,
    title, major, minor, classification, document_type,
    content, changelog, status, orientation,
    published_at, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id, @document_id,
    @title, @major, 0,
    'CONFIDENTIAL'::document_classification,
    'REGISTER'::document_type,
    @content, '',
    'PUBLISHED'::document_version_status,
    'PORTRAIT'::document_version_orientation,
    @published_at, @published_at, @published_at
)`,
					pgx.NamedArgs{
						"id":              versionID,
						"tenant_id":       org.tenantID,
						"organization_id": org.organizationID,
						"document_id":     documentID,
						"title":           k.title,
						"major":           major + 1,
						"content":         content,
						"published_at":    snap.publishedAt,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot insert %s version for snapshot %s: %w", k.name, snap.snapshotID, err)
				}
				versionsInserted++
			}

			if versionsInserted == 0 {
				continue
			}

			_, err := tx.Exec(
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
					"current_published_major": versionsInserted,
					"created_at":              now,
					"updated_at":              now,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot insert %s document for org %s: %w", k.name, org.organizationID, err)
			}
			stats.documents++
			stats.versions += versionsInserted

			_, err = tx.Exec(
				ctx,
				fmt.Sprintf(`
INSERT INTO generated_documents (organization_id, tenant_id, %s, created_at, updated_at)
VALUES (@organization_id, @tenant_id, @document_id, @created_at, @updated_at)
ON CONFLICT (organization_id) DO UPDATE SET %s = @document_id, updated_at = @updated_at`, k.column, k.column),
				pgx.NamedArgs{
					"organization_id": org.organizationID,
					"tenant_id":       org.tenantID,
					"document_id":     documentID,
					"created_at":      now,
					"updated_at":      now,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot link %s document to org %s: %w", k.name, org.organizationID, err)
			}
		}

		fmt.Printf("migrated org %s (%s) — %d snapshot(s)\n",
			org.organizationID, org.organizationName, len(snapshots))
	}

	if dryRun {
		fmt.Printf("\n%d organization(s) would be migrated\n", len(orgs))
		return nil
	}

	fmt.Printf("\ncreated %d document(s), %d version(s)\n", stats.documents, stats.versions)
	return nil
}

func loadOrgsWithSnapshots(ctx context.Context, tx pg.Tx) ([]orgWithSnapshots, error) {
	rows, err := tx.Query(
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
        WHERE gd.organization_id = o.id
          AND gd.processing_activities_document_id IS NOT NULL
    )
    AND EXISTS (
        SELECT 1 FROM snapshots s
        WHERE s.organization_id = o.id AND s.type = 'PROCESSING_ACTIVITIES'
    )
ORDER BY o.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query organizations with snapshots: %w", err)
	}
	defer rows.Close()

	var result []orgWithSnapshots
	for rows.Next() {
		var o orgWithSnapshots
		var createdAt time.Time
		if err := rows.Scan(&o.organizationID, &o.tenantID, &o.organizationName, &createdAt); err != nil {
			return nil, fmt.Errorf("cannot scan organization: %w", err)
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func loadSnapshots(ctx context.Context, tx pg.Tx, organizationID gid.GID) ([]processingActivitySnapshot, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    s.id,
    s.created_at
FROM snapshots s
WHERE s.organization_id = @organization_id
    AND s.type = 'PROCESSING_ACTIVITIES'
ORDER BY s.created_at ASC;
`,
		pgx.NamedArgs{"organization_id": organizationID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query snapshots for org %s: %w", organizationID, err)
	}
	defer rows.Close()

	var result []processingActivitySnapshot
	for rows.Next() {
		var s processingActivitySnapshot
		if err := rows.Scan(&s.snapshotID, &s.publishedAt); err != nil {
			return nil, fmt.Errorf("cannot scan snapshot: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func buildProcessingActivityContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, int, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    pa.id,
    pa.name,
    pa.purpose,
    pa.data_subject_category,
    pa.personal_data_category,
    pa.special_or_criminal_data,
    pa.consent_evidence_link,
    pa.lawful_basis,
    pa.recipients,
    pa.location,
    pa.international_transfers,
    pa.transfer_safeguards,
    pa.retention_period,
    pa.security_measures,
    pa.data_protection_impact_assessment_needed,
    pa.transfer_impact_assessment_needed,
    pa.last_review_date,
    pa.next_review_date,
    pa.role,
    COALESCE(p.full_name, '')
FROM processing_activities pa
LEFT JOIN iam_membership_profiles p ON p.id = pa.dpo_profile_id
WHERE pa.snapshot_id = @snapshot_id
ORDER BY pa.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", 0, fmt.Errorf("cannot load snapshot processing activities: %w", err)
	}
	defer rows.Close()

	type paRow struct {
		id                                   gid.GID
		name                                 string
		purpose                              *string
		dataSubjectCategory                  *string
		personalDataCategory                 *string
		specialOrCriminalData                string
		consentEvidenceLink                  *string
		lawfulBasis                          string
		recipients                           *string
		location                             *string
		internationalTransfers               bool
		transferSafeguards                   *string
		retentionPeriod                      *string
		securityMeasures                     *string
		dataProtectionImpactAssessmentNeeded string
		transferImpactAssessmentNeeded       string
		lastReviewDate                       *time.Time
		nextReviewDate                       *time.Time
		role                                 string
		dpoName                              string
	}

	var pas []paRow
	for rows.Next() {
		var p paRow
		if err := rows.Scan(
			&p.id, &p.name, &p.purpose, &p.dataSubjectCategory, &p.personalDataCategory,
			&p.specialOrCriminalData, &p.consentEvidenceLink, &p.lawfulBasis,
			&p.recipients, &p.location, &p.internationalTransfers, &p.transferSafeguards,
			&p.retentionPeriod, &p.securityMeasures,
			&p.dataProtectionImpactAssessmentNeeded, &p.transferImpactAssessmentNeeded,
			&p.lastReviewDate, &p.nextReviewDate, &p.role, &p.dpoName,
		); err != nil {
			return "", 0, fmt.Errorf("cannot scan PA: %w", err)
		}
		pas = append(pas, p)
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}

	if len(pas) == 0 {
		return "", 0, nil
	}

	vendorMap, err := loadVendorsForSnapshot(ctx, tx, snapshotID)
	if err != nil {
		return "", 0, err
	}

	listRows := make([]docgen.ProcessingActivityListRow, len(pas))
	for i, p := range pas {
		dpo := "Not assigned"
		if p.dpoName != "" {
			dpo = p.dpoName
		}

		vendors := "None"
		if v, ok := vendorMap[p.id]; ok && len(v) > 0 {
			vendors = strings.Join(v, ", ")
		}

		listRows[i] = docgen.ProcessingActivityListRow{
			Name:                                 p.name,
			Purpose:                              derefOrNotSpecified(p.purpose),
			Role:                                 formatRoleString(p.role),
			DataSubjectCategory:                  derefOrNotSpecified(p.dataSubjectCategory),
			PersonalDataCategory:                 derefOrNotSpecified(p.personalDataCategory),
			SpecialOrCriminalData:                formatSpecialOrCriminalDataString(p.specialOrCriminalData),
			LawfulBasis:                          formatLawfulBasisString(p.lawfulBasis),
			ConsentEvidenceLink:                  derefOrNotSpecified(p.consentEvidenceLink),
			Recipients:                           derefOrNotSpecified(p.recipients),
			Location:                             derefOrNotSpecified(p.location),
			InternationalTransfers:               yesNoLabel(p.internationalTransfers),
			TransferSafeguards:                   formatTransferSafeguardString(p.transferSafeguards),
			RetentionPeriod:                      derefOrNotSpecified(p.retentionPeriod),
			SecurityMeasures:                     derefOrNotSpecified(p.securityMeasures),
			DataProtectionImpactAssessmentNeeded: formatYesNoString(p.dataProtectionImpactAssessmentNeeded),
			TransferImpactAssessmentNeeded:       formatYesNoString(p.transferImpactAssessmentNeeded),
			LastReviewDate:                       formatDateOrNotSpecified(p.lastReviewDate),
			NextReviewDate:                       formatDateOrNotSpecified(p.nextReviewDate),
			DataProtectionOfficer:                dpo,
			Vendors:                              vendors,
		}
	}

	content, err := probo.BuildProcessingActivityListDocument(docgen.ProcessingActivityListData{
		Title:                     "Processing Activities",
		OrganizationName:          orgName,
		CreatedAt:                 publishedAt,
		TotalProcessingActivities: len(listRows),
		Rows:                      listRows,
	})
	if err != nil {
		return "", 0, err
	}
	return content, len(listRows), nil
}

func loadVendorsForSnapshot(ctx context.Context, tx pg.Tx, snapshotID string) (map[gid.GID][]string, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT pav.processing_activity_id, v.name
FROM processing_activity_vendors pav
INNER JOIN vendors v ON v.id = pav.vendor_id
WHERE pav.snapshot_id = @snapshot_id
ORDER BY pav.processing_activity_id, v.name;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load snapshot vendors: %w", err)
	}
	defer rows.Close()

	result := make(map[gid.GID][]string)
	for rows.Next() {
		var paID gid.GID
		var name string
		if err := rows.Scan(&paID, &name); err != nil {
			return nil, fmt.Errorf("cannot scan vendor row: %w", err)
		}
		result[paID] = append(result[paID], name)
	}
	return result, rows.Err()
}

func buildDPIAContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, int, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    pa.name,
    dpia.description,
    dpia.necessity_and_proportionality,
    dpia.potential_risk,
    dpia.mitigations,
    dpia.residual_risk
FROM processing_activity_data_protection_impact_assessments dpia
INNER JOIN processing_activities pa ON pa.id = dpia.processing_activity_id
WHERE dpia.snapshot_id = @snapshot_id
ORDER BY pa.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", 0, fmt.Errorf("cannot load snapshot DPIAs: %w", err)
	}
	defer rows.Close()

	var listRows []docgen.DataProtectionImpactAssessmentListRow
	for rows.Next() {
		var name string
		var description, necessity, potentialRisk, mitigations *string
		var residualRisk *string
		if err := rows.Scan(&name, &description, &necessity, &potentialRisk, &mitigations, &residualRisk); err != nil {
			return "", 0, fmt.Errorf("cannot scan DPIA: %w", err)
		}
		listRows = append(listRows, docgen.DataProtectionImpactAssessmentListRow{
			ProcessingActivityName:      name,
			Description:                 derefOrNotSpecified(description),
			NecessityAndProportionality: derefOrNotSpecified(necessity),
			PotentialRisk:               derefOrNotSpecified(potentialRisk),
			Mitigations:                 derefOrNotSpecified(mitigations),
			ResidualRisk:                formatResidualRiskString(residualRisk),
		})
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}

	if len(listRows) == 0 {
		return "", 0, nil
	}

	content, err := probo.BuildDataProtectionImpactAssessmentListDocument(docgen.DataProtectionImpactAssessmentListData{
		Title:                                "Data Protection Impact Assessments",
		OrganizationName:                     orgName,
		CreatedAt:                            publishedAt,
		TotalDataProtectionImpactAssessments: len(listRows),
		Rows:                                 listRows,
	})
	if err != nil {
		return "", 0, err
	}
	return content, len(listRows), nil
}

func buildTIAContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, int, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    pa.name,
    tia.data_subjects,
    tia.legal_mechanism,
    tia.transfer,
    tia.local_law_risk,
    tia.supplementary_measures
FROM processing_activity_transfer_impact_assessments tia
INNER JOIN processing_activities pa ON pa.id = tia.processing_activity_id
WHERE tia.snapshot_id = @snapshot_id
ORDER BY pa.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", 0, fmt.Errorf("cannot load snapshot TIAs: %w", err)
	}
	defer rows.Close()

	var listRows []docgen.TransferImpactAssessmentListRow
	for rows.Next() {
		var name string
		var dataSubjects, legalMechanism, transfer, localLawRisk, supplementary *string
		if err := rows.Scan(&name, &dataSubjects, &legalMechanism, &transfer, &localLawRisk, &supplementary); err != nil {
			return "", 0, fmt.Errorf("cannot scan TIA: %w", err)
		}
		listRows = append(listRows, docgen.TransferImpactAssessmentListRow{
			ProcessingActivityName: name,
			DataSubjects:           derefOrNotSpecified(dataSubjects),
			LegalMechanism:         derefOrNotSpecified(legalMechanism),
			Transfer:               derefOrNotSpecified(transfer),
			LocalLawRisk:           derefOrNotSpecified(localLawRisk),
			SupplementaryMeasures:  derefOrNotSpecified(supplementary),
		})
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}

	if len(listRows) == 0 {
		return "", 0, nil
	}

	content, err := probo.BuildTransferImpactAssessmentListDocument(docgen.TransferImpactAssessmentListData{
		Title:                          "Transfer Impact Assessments",
		OrganizationName:               orgName,
		CreatedAt:                      publishedAt,
		TotalTransferImpactAssessments: len(listRows),
		Rows:                           listRows,
	})
	if err != nil {
		return "", 0, err
	}
	return content, len(listRows), nil
}

func derefOrNotSpecified(s *string) string {
	if s == nil || *s == "" {
		return "Not specified"
	}
	return *s
}

func formatDateOrNotSpecified(t *time.Time) string {
	if t == nil {
		return "Not specified"
	}
	return t.Format("January 2, 2006")
}

func yesNoLabel(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatYesNoString(s string) string {
	switch s {
	case "NEEDED":
		return "Yes"
	case "NOT_NEEDED":
		return "No"
	default:
		return s
	}
}

func formatRoleString(role string) string {
	switch role {
	case "CONTROLLER":
		return "Controller"
	case "PROCESSOR":
		return "Processor"
	default:
		return role
	}
}

func formatLawfulBasisString(b string) string {
	switch b {
	case "CONSENT":
		return "Consent"
	case "CONTRACTUAL_NECESSITY":
		return "Contractual Necessity"
	case "LEGAL_OBLIGATION":
		return "Legal Obligation"
	case "LEGITIMATE_INTEREST":
		return "Legitimate Interest"
	case "PUBLIC_TASK":
		return "Public Task"
	case "VITAL_INTERESTS":
		return "Vital Interests"
	default:
		return b
	}
}

func formatSpecialOrCriminalDataString(s string) string {
	switch s {
	case "YES":
		return "Yes"
	case "NO":
		return "No"
	case "POSSIBLE":
		return "Possible"
	default:
		return s
	}
}

func formatTransferSafeguardString(s *string) string {
	if s == nil {
		return "Not specified"
	}
	switch *s {
	case "STANDARD_CONTRACTUAL_CLAUSES":
		return "Standard Contractual Clauses"
	case "BINDING_CORPORATE_RULES":
		return "Binding Corporate Rules"
	case "ADEQUACY_DECISION":
		return "Adequacy Decision"
	case "DEROGATIONS":
		return "Derogations"
	case "CODES_OF_CONDUCT":
		return "Codes of Conduct"
	case "CERTIFICATION_MECHANISMS":
		return "Certification Mechanisms"
	default:
		return *s
	}
}

func formatResidualRiskString(s *string) string {
	if s == nil {
		return "Not specified"
	}
	switch *s {
	case "LOW":
		return "Low"
	case "MEDIUM":
		return "Medium"
	case "HIGH":
		return "High"
	default:
		return *s
	}
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
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
