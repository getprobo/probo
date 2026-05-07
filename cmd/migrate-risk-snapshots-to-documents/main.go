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

// Command migrate-risk-snapshots-to-documents creates documents and document
// versions from existing risk snapshots. For each organization that has risk
// snapshots, it generates a risk list document using the same ProseMirror
// builder as the publish flow.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
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

type orgWithRiskSnapshots struct {
	organizationID   gid.GID
	tenantID         gid.TenantID
	organizationName string
}

type riskSnapshot struct {
	snapshotID  string
	publishedAt time.Time
}

func migrate(ctx context.Context, pgClient *pg.Client, dryRun bool) error {
	var orgs []orgWithRiskSnapshots
	err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error
		orgs, err = loadOrgsWithRiskSnapshots(ctx, conn)
		return err
	})
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("no organizations with risk snapshots to migrate")
		return nil
	}

	var stats struct {
		documents, versions, failed int
	}

	for _, org := range orgs {
		if dryRun {
			var count int
			err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
				snapshots, err := loadRiskSnapshots(ctx, conn, org.organizationID)
				count = len(snapshots)
				return err
			})
			if err != nil {
				return err
			}
			fmt.Printf("would migrate org %s (%s) — %d risk snapshot(s)\n",
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

func migrateOrg(ctx context.Context, tx pg.Tx, org orgWithRiskSnapshots) error {
	snapshots, err := loadRiskSnapshots(ctx, tx, org.organizationID)
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
		`INSERT INTO generated_documents (organization_id, tenant_id, risks_document_id, created_at, updated_at)
VALUES (@organization_id, @tenant_id, @risks_document_id, @created_at, @updated_at)
ON CONFLICT (organization_id) DO UPDATE SET risks_document_id = @risks_document_id, updated_at = @updated_at`,
		pgx.NamedArgs{
			"organization_id":   org.organizationID,
			"tenant_id":         org.tenantID,
			"risks_document_id": documentID,
			"created_at":        now,
			"updated_at":        now,
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
				"title":           "Risks",
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

func loadOrgsWithRiskSnapshots(ctx context.Context, conn pg.Querier) ([]orgWithRiskSnapshots, error) {
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
        WHERE gd.organization_id = o.id AND gd.risks_document_id IS NOT NULL
    )
    AND EXISTS (
        SELECT 1 FROM snapshots s
        WHERE s.organization_id = o.id AND s.type = 'RISKS'
    )
ORDER BY o.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query organizations with risk snapshots: %w", err)
	}
	defer rows.Close()

	var result []orgWithRiskSnapshots
	for rows.Next() {
		var o orgWithRiskSnapshots
		var createdAt time.Time
		if err := rows.Scan(&o.organizationID, &o.tenantID, &o.organizationName, &createdAt); err != nil {
			return nil, fmt.Errorf("cannot scan organization: %w", err)
		}
		result = append(result, o)
	}

	return result, rows.Err()
}

func loadRiskSnapshots(ctx context.Context, conn pg.Querier, organizationID gid.GID) ([]riskSnapshot, error) {
	rows, err := conn.Query(
		ctx,
		`
SELECT DISTINCT
    s.id,
    s.created_at
FROM snapshots s
WHERE s.organization_id = @organization_id
    AND s.type = 'RISKS'
ORDER BY s.created_at ASC;
`,
		pgx.NamedArgs{"organization_id": organizationID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query risk snapshots for org %s: %w", organizationID, err)
	}
	defer rows.Close()

	var result []riskSnapshot
	for rows.Next() {
		var s riskSnapshot
		if err := rows.Scan(&s.snapshotID, &s.publishedAt); err != nil {
			return nil, fmt.Errorf("cannot scan snapshot: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

type riskInfo struct {
	id                 string
	name               string
	description        *string
	category           string
	treatment          string
	note               string
	ownerName          string
	inherentLikelihood int
	inherentImpact     int
	inherentRiskScore  int
	residualLikelihood int
	residualImpact     int
	residualRiskScore  int
}

func buildSnapshotContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, error) {
	riskRows, err := tx.Query(
		ctx,
		`
SELECT
    r.id,
    r.name,
    r.description,
    r.category,
    r.treatment::text,
    r.note,
    COALESCE(NULLIF(p.full_name, ''), 'Not assigned'),
    r.inherent_likelihood,
    r.inherent_impact,
    r.inherent_risk_score,
    r.residual_likelihood,
    r.residual_impact,
    r.residual_risk_score
FROM risks r
LEFT JOIN iam_membership_profiles p ON p.id = r.owner_profile_id
WHERE r.snapshot_id = @snapshot_id
ORDER BY r.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot risks: %w", err)
	}
	defer riskRows.Close()

	var risks []riskInfo
	for riskRows.Next() {
		var r riskInfo
		if err := riskRows.Scan(
			&r.id, &r.name, &r.description, &r.category, &r.treatment, &r.note,
			&r.ownerName,
			&r.inherentLikelihood, &r.inherentImpact, &r.inherentRiskScore,
			&r.residualLikelihood, &r.residualImpact, &r.residualRiskScore,
		); err != nil {
			return "", fmt.Errorf("cannot scan risk: %w", err)
		}
		risks = append(risks, r)
	}
	if err := riskRows.Err(); err != nil {
		return "", err
	}

	rows := make([]docgen.RiskListRow, 0, len(risks))
	for _, r := range risks {
		row := docgen.RiskListRow{
			Name:                    r.name,
			Description:             derefOrNotSpecified(r.description),
			Category:                stringOrNotSpecified(r.category),
			Treatment:               formatTreatment(r.treatment),
			Owner:                   r.ownerName,
			InherentLikelihood:      r.inherentLikelihood,
			InherentLikelihoodLabel: riskLikelihoodLabel(r.inherentLikelihood),
			InherentImpact:          r.inherentImpact,
			InherentImpactLabel:     riskImpactLabel(r.inherentImpact),
			InherentRiskScore:       r.inherentRiskScore,
			InherentSeverity:        riskSeverityLabel(r.inherentRiskScore),
			ResidualLikelihood:      r.residualLikelihood,
			ResidualLikelihoodLabel: riskLikelihoodLabel(r.residualLikelihood),
			ResidualImpact:          r.residualImpact,
			ResidualImpactLabel:     riskImpactLabel(r.residualImpact),
			ResidualRiskScore:       r.residualRiskScore,
			ResidualSeverity:        riskSeverityLabel(r.residualRiskScore),
			Note:                    stringOrNotSpecified(r.note),
		}
		rows = append(rows, row)
	}

	docData := docgen.RiskListData{
		Title:            "Risks",
		OrganizationName: orgName,
		CreatedAt:        publishedAt,
		TotalRisks:       len(rows),
		Rows:             rows,
	}

	return probo.BuildRiskListDocument(docData)
}

func derefOrNotSpecified(s *string) string {
	if s == nil || *s == "" {
		return "Not specified"
	}
	return *s
}

func stringOrNotSpecified(s string) string {
	if s == "" {
		return "Not specified"
	}
	return s
}

func formatTreatment(t string) string {
	switch t {
	case "MITIGATED":
		return "Mitigated"
	case "ACCEPTED":
		return "Accepted"
	case "AVOIDED":
		return "Avoided"
	case "TRANSFERRED":
		return "Transferred"
	default:
		return stringOrNotSpecified(t)
	}
}

func riskLikelihoodLabel(v int) string {
	switch v {
	case 1:
		return "Improbable"
	case 2:
		return "Remote"
	case 3:
		return "Occasional"
	case 4:
		return "Probable"
	case 5:
		return "Frequent"
	default:
		return "Unknown"
	}
}

func riskImpactLabel(v int) string {
	switch v {
	case 1:
		return "Negligible"
	case 2:
		return "Low"
	case 3:
		return "Moderate"
	case 4:
		return "Significant"
	case 5:
		return "Catastrophic"
	default:
		return "Unknown"
	}
}

func riskSeverityLabel(score int) string {
	switch {
	case score >= 15:
		return "Critical"
	case score >= 5:
		return "High"
	default:
		return "Low"
	}
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN")
	}

	var opts []pg.Option

	if u.Host != "" {
		host := u.Host
		if u.Port() == "" {
			host = net.JoinHostPort(u.Hostname(), "5432")
		}
		opts = append(opts, pg.WithAddr(host))
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
