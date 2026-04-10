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

// Command migrate-soa-snapshots-to-documents creates documents, document
// versions, approval quorums, and approval decisions from existing SOA
// snapshots. For each snapshot it generates the ProseMirror content using
// the same builder as the publish flow.
package main

import (
	"context"
	"flag"
	"fmt"
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

	return pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return migrate(ctx, tx, dryRun)
	})
}

type originalSOA struct {
	id             string
	tenantID       gid.TenantID
	organizationID gid.GID
	name           string
	ownerProfileID *string
}

type snapshotSOA struct {
	id          string
	snapshotID  string
	publishedAt time.Time
}

func migrate(ctx context.Context, tx pg.Tx, dryRun bool) error {
	originals, err := loadOriginalSOAs(ctx, tx)
	if err != nil {
		return err
	}

	if len(originals) == 0 {
		fmt.Println("no SOAs with snapshots to migrate")
		return nil
	}

	var stats struct {
		documents, versions, quorums, decisions int
	}

	for _, orig := range originals {
		snapshots, err := loadSnapshots(ctx, tx, orig.id)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("would migrate SOA %s (%s) — %d snapshot(s)\n", orig.id, orig.name, len(snapshots))
			continue
		}

		documentID := gid.New(orig.tenantID, coredata.DocumentEntityType)
		now := time.Now()

		_, err = tx.Exec(
			ctx,
			`
INSERT INTO documents (
    id, tenant_id, organization_id, content_source,
    current_published_major, current_published_minor,
    trust_center_visibility, status, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id,
    'GENERATED'::document_content_source,
    @current_published_major, 0,
    'NONE'::trust_center_visibility,
    'ACTIVE'::document_status,
    @created_at, @updated_at
)`,
			pgx.NamedArgs{
				"id":                      documentID,
				"tenant_id":               orig.tenantID,
				"organization_id":         orig.organizationID,
				"current_published_major": len(snapshots),
				"created_at":              now,
				"updated_at":              now,
			},
		)
		if err != nil {
			return fmt.Errorf("cannot insert document for SOA %s: %w", orig.id, err)
		}
		stats.documents++

		_, err = tx.Exec(
			ctx,
			`UPDATE statements_of_applicability SET document_id = @document_id WHERE id = @id`,
			pgx.NamedArgs{
				"document_id": documentID,
				"id":          orig.id,
			},
		)
		if err != nil {
			return fmt.Errorf("cannot link document to SOA %s: %w", orig.id, err)
		}

		for major, snap := range snapshots {
			content, err := buildSnapshotContent(ctx, tx, snap.id, orig.name)
			if err != nil {
				return fmt.Errorf("cannot build content for snapshot %s of SOA %s: %w", snap.snapshotID, orig.id, err)
			}

			versionID := gid.New(orig.tenantID, coredata.DocumentVersionEntityType)

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
    'STATEMENT_OF_APPLICABILITY'::document_type,
    @content, '',
    'PUBLISHED'::document_version_status,
    'LANDSCAPE'::document_version_orientation,
    @published_at, @published_at, @published_at
)`,
				pgx.NamedArgs{
					"id":              versionID,
					"tenant_id":       orig.tenantID,
					"organization_id": orig.organizationID,
					"document_id":     documentID,
					"title":           orig.name,
					"major":           major + 1,
					"content":         content,
					"published_at":    snap.publishedAt,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot insert version for snapshot %s: %w", snap.snapshotID, err)
			}
			stats.versions++

			if orig.ownerProfileID != nil {
				quorumID := gid.New(orig.tenantID, coredata.DocumentVersionApprovalQuorumEntityType)

				_, err = tx.Exec(
					ctx,
					`
INSERT INTO document_version_approval_quorums (
    id, tenant_id, organization_id, version_id, status, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id, @version_id,
    'APPROVED'::document_version_approval_quorum_status,
    @created_at, @created_at
)`,
					pgx.NamedArgs{
						"id":              quorumID,
						"tenant_id":       orig.tenantID,
						"organization_id": orig.organizationID,
						"version_id":      versionID,
						"created_at":      snap.publishedAt,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot insert quorum for snapshot %s: %w", snap.snapshotID, err)
				}
				stats.quorums++

				decisionID := gid.New(orig.tenantID, coredata.DocumentVersionApprovalDecisionEntityType)

				_, err = tx.Exec(
					ctx,
					`
INSERT INTO document_version_approval_decisions (
    id, tenant_id, organization_id, quorum_id,
    approver_id, state, decided_at, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id, @quorum_id,
    @approver_id,
    'APPROVED'::document_version_approval_decision_state,
    @decided_at, @decided_at, @decided_at
)`,
					pgx.NamedArgs{
						"id":              decisionID,
						"tenant_id":       orig.tenantID,
						"organization_id": orig.organizationID,
						"quorum_id":       quorumID,
						"approver_id":     *orig.ownerProfileID,
						"decided_at":      snap.publishedAt,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot insert decision for snapshot %s: %w", snap.snapshotID, err)
				}
				stats.decisions++
			}
		}

		fmt.Printf("migrated SOA %s (%s) — %d version(s)\n", orig.id, orig.name, len(snapshots))
	}

	if dryRun {
		fmt.Printf("\n%d SOA(s) would be migrated\n", len(originals))
		return nil
	}

	fmt.Printf("\ncreated %d document(s), %d version(s), %d quorum(s), %d decision(s)\n",
		stats.documents, stats.versions, stats.quorums, stats.decisions)

	return nil
}

func loadOriginalSOAs(ctx context.Context, tx pg.Tx) ([]originalSOA, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    soa.id,
    soa.tenant_id,
    soa.organization_id,
    soa.name,
    soa.owner_profile_id
FROM statements_of_applicability soa
WHERE soa.snapshot_id IS NULL
    AND soa.document_id IS NULL
    AND EXISTS (
        SELECT 1 FROM statements_of_applicability snap
        WHERE snap.source_id = soa.id AND snap.snapshot_id IS NOT NULL
    )
ORDER BY soa.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query original SOAs: %w", err)
	}
	defer rows.Close()

	var result []originalSOA
	for rows.Next() {
		var o originalSOA
		if err := rows.Scan(&o.id, &o.tenantID, &o.organizationID, &o.name, &o.ownerProfileID); err != nil {
			return nil, fmt.Errorf("cannot scan original SOA: %w", err)
		}
		result = append(result, o)
	}

	return result, rows.Err()
}

func loadSnapshots(ctx context.Context, tx pg.Tx, originalSOAID string) ([]snapshotSOA, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    snap.id,
    snap.snapshot_id,
    snap_record.created_at
FROM statements_of_applicability snap
JOIN snapshots snap_record ON snap_record.id = snap.snapshot_id
WHERE snap.source_id = @source_id
    AND snap.snapshot_id IS NOT NULL
ORDER BY snap_record.created_at ASC;
`,
		pgx.NamedArgs{"source_id": originalSOAID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query snapshots for SOA %s: %w", originalSOAID, err)
	}
	defer rows.Close()

	var result []snapshotSOA
	for rows.Next() {
		var s snapshotSOA
		if err := rows.Scan(&s.id, &s.snapshotID, &s.publishedAt); err != nil {
			return nil, fmt.Errorf("cannot scan snapshot: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

type snapshotControl struct {
	frameworkName               string
	sectionTitle                string
	controlName                 string
	applicability               bool
	justification               *string
	bestPractice                bool
	implemented                 string
	notImplementedJustification *string
	legalCount                  int
	contractualCount            int
	hasRisk                     bool
}

func buildSnapshotContent(ctx context.Context, tx pg.Tx, snapshotSOAID string, soaName string) (string, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    f.name AS framework_name,
    c.section_title,
    c.name AS control_name,
    stmt.applicability,
    stmt.justification,
    c.best_practice,
    c.implemented,
    c.not_implemented_justification,
    (SELECT COUNT(*) FROM control_obligations co
     JOIN obligations o ON o.id = co.obligation_id
     WHERE co.control_id = c.id AND o.type = 'LEGAL') AS legal_count,
    (SELECT COUNT(*) FROM control_obligations co
     JOIN obligations o ON o.id = co.obligation_id
     WHERE co.control_id = c.id AND o.type = 'CONTRACTUAL') AS contractual_count,
    EXISTS (
        SELECT 1 FROM control_risks cr WHERE cr.control_id = c.id
    ) AS has_risk
FROM applicability_statements stmt
JOIN controls c ON c.id = stmt.control_id
JOIN frameworks f ON f.id = c.framework_id
WHERE stmt.statement_of_applicability_id = @soa_id
ORDER BY f.name, c.section_title;
`,
		pgx.NamedArgs{"soa_id": snapshotSOAID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot controls: %w", err)
	}
	defer rows.Close()

	frameworkControlsMap := make(map[string][]docgen.ControlData)
	var frameworkOrder []string

	for rows.Next() {
		var sc snapshotControl
		if err := rows.Scan(
			&sc.frameworkName,
			&sc.sectionTitle,
			&sc.controlName,
			&sc.applicability,
			&sc.justification,
			&sc.bestPractice,
			&sc.implemented,
			&sc.notImplementedJustification,
			&sc.legalCount,
			&sc.contractualCount,
			&sc.hasRisk,
		); err != nil {
			return "", fmt.Errorf("cannot scan control: %w", err)
		}

		if _, exists := frameworkControlsMap[sc.frameworkName]; !exists {
			frameworkOrder = append(frameworkOrder, sc.frameworkName)
		}

		var regulatory, contractual, bestPractice, riskAssessment *bool

		if sc.applicability {
			regulatory = new(sc.legalCount > 0)
			contractual = new(sc.contractualCount > 0)
			riskAssessment = new(sc.hasRisk)
			bestPractice = new(sc.bestPractice)
		}

		applicability := sc.applicability
		implemented := sc.implemented

		frameworkControlsMap[sc.frameworkName] = append(
			frameworkControlsMap[sc.frameworkName],
			docgen.ControlData{
				FrameworkName: sc.frameworkName,
				SectionTitle:  sc.sectionTitle,
				Name:          sc.controlName,
				Applicability: &applicability,
				Justification: sc.justification,
				BestPractice:  bestPractice,
				Implemented:   &implemented,
				NotImplementedJustification: func() *string {
					if sc.implemented == "IMPLEMENTED" {
						return nil
					}
					return sc.notImplementedJustification
				}(),
				Regulatory:     regulatory,
				Contractual:    contractual,
				RiskAssessment: riskAssessment,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	totalControls := 0
	frameworkGroups := make([]docgen.FrameworkControlGroup, len(frameworkOrder))
	for i, name := range frameworkOrder {
		frameworkGroups[i] = docgen.FrameworkControlGroup{
			FrameworkName: name,
			Controls:      frameworkControlsMap[name],
		}
		totalControls += len(frameworkControlsMap[name])
	}

	data := docgen.StatementOfApplicabilityData{
		Title:           soaName,
		TotalControls:   totalControls,
		FrameworkGroups: frameworkGroups,
	}

	return probo.BuildStatementOfApplicabilityProseMirrorDocument(data)
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
