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

// Command migrate-obligation-snapshots-to-documents creates documents and document
// versions from existing obligation snapshots. For each organization that has
// obligation snapshots, it generates an obligation register document using the same
// ProseMirror builder as the publish flow.
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

type orgWithObligationSnapshots struct {
	organizationID   gid.GID
	tenantID         gid.TenantID
	organizationName string
}

type obligationSnapshot struct {
	snapshotID  string
	publishedAt time.Time
}

func migrate(ctx context.Context, tx pg.Tx, dryRun bool) error {
	orgs, err := loadOrgsWithObligationSnapshots(ctx, tx)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("no organizations with obligation snapshots to migrate")
		return nil
	}

	var stats struct {
		documents, versions int
	}

	for _, org := range orgs {
		snapshots, err := loadObligationSnapshots(ctx, tx, org.organizationID)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("would migrate org %s (%s) — %d obligation snapshot(s)\n",
				org.organizationID, org.organizationName, len(snapshots))
			continue
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
			return fmt.Errorf("cannot insert document for org %s: %w", org.organizationID, err)
		}
		stats.documents++

		_, err = tx.Exec(
			ctx,
			`INSERT INTO generated_documents (organization_id, tenant_id, obligations_document_id, created_at, updated_at)
VALUES (@organization_id, @tenant_id, @obligations_document_id, @created_at, @updated_at)
ON CONFLICT (organization_id) DO UPDATE SET obligations_document_id = @obligations_document_id, updated_at = @updated_at`,
			pgx.NamedArgs{
				"organization_id":         org.organizationID,
				"tenant_id":               org.tenantID,
				"obligations_document_id": documentID,
				"created_at":              now,
				"updated_at":              now,
			},
		)
		if err != nil {
			return fmt.Errorf("cannot link document to org %s: %w", org.organizationID, err)
		}

		for major, snap := range snapshots {
			content, err := buildSnapshotContent(ctx, tx, snap.snapshotID, org.organizationName, snap.publishedAt)
			if err != nil {
				return fmt.Errorf("cannot build content for snapshot %s of org %s: %w",
					snap.snapshotID, org.organizationID, err)
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
					"title":           "Obligation Register",
					"major":           major + 1,
					"content":         content,
					"published_at":    snap.publishedAt,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot insert version for snapshot %s: %w", snap.snapshotID, err)
			}
			stats.versions++
		}

		fmt.Printf("migrated org %s (%s) — %d version(s)\n",
			org.organizationID, org.organizationName, len(snapshots))
	}

	if dryRun {
		fmt.Printf("\n%d organization(s) would be migrated\n", len(orgs))
		return nil
	}

	fmt.Printf("\ncreated %d document(s), %d version(s)\n", stats.documents, stats.versions)

	return nil
}

func loadOrgsWithObligationSnapshots(ctx context.Context, tx pg.Tx) ([]orgWithObligationSnapshots, error) {
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
        WHERE gd.organization_id = o.id AND gd.obligations_document_id IS NOT NULL
    )
    AND EXISTS (
        SELECT 1 FROM obligations ob
        WHERE ob.organization_id = o.id AND ob.snapshot_id IS NOT NULL
    )
ORDER BY o.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query organizations with obligation snapshots: %w", err)
	}
	defer rows.Close()

	var result []orgWithObligationSnapshots
	for rows.Next() {
		var o orgWithObligationSnapshots
		var createdAt time.Time
		if err := rows.Scan(&o.organizationID, &o.tenantID, &o.organizationName, &createdAt); err != nil {
			return nil, fmt.Errorf("cannot scan organization: %w", err)
		}
		result = append(result, o)
	}

	return result, rows.Err()
}

func loadObligationSnapshots(ctx context.Context, tx pg.Tx, organizationID gid.GID) ([]obligationSnapshot, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT DISTINCT
    s.id,
    s.created_at
FROM snapshots s
WHERE s.organization_id = @organization_id
    AND s.type = 'OBLIGATIONS'
ORDER BY s.created_at ASC;
`,
		pgx.NamedArgs{"organization_id": organizationID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query obligation snapshots for org %s: %w", organizationID, err)
	}
	defer rows.Close()

	var result []obligationSnapshot
	for rows.Next() {
		var s obligationSnapshot
		if err := rows.Scan(&s.snapshotID, &s.publishedAt); err != nil {
			return nil, fmt.Errorf("cannot scan snapshot: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func buildSnapshotContent(
	ctx context.Context,
	tx pg.Tx,
	snapshotID string,
	orgName string,
	publishedAt time.Time,
) (string, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT
    ob.area,
    ob.source,
    ob.requirement,
    ob.status,
    ob.type,
    ob.regulator,
    ob.due_date,
    COALESCE(p.full_name, '-')
FROM obligations ob
LEFT JOIN iam_membership_profiles p ON p.id = ob.owner_profile_id
WHERE ob.snapshot_id = @snapshot_id
ORDER BY ob.created_at ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot obligations: %w", err)
	}
	defer rows.Close()

	type obligationInfo struct {
		area        *string
		source      *string
		requirement *string
		status      string
		oblType     string
		regulator   *string
		dueDate     *time.Time
		ownerName   string
	}

	var obligations []obligationInfo
	for rows.Next() {
		var o obligationInfo
		if err := rows.Scan(&o.area, &o.source, &o.requirement, &o.status, &o.oblType, &o.regulator, &o.dueDate, &o.ownerName); err != nil {
			return "", fmt.Errorf("cannot scan obligation: %w", err)
		}
		obligations = append(obligations, o)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	obligationRows := make([]docgen.ObligationListRow, len(obligations))
	for i, o := range obligations {
		area := "-"
		if o.area != nil && *o.area != "" {
			area = *o.area
		}

		source := "-"
		if o.source != nil && *o.source != "" {
			source = *o.source
		}

		requirement := "-"
		if o.requirement != nil && *o.requirement != "" {
			requirement = *o.requirement
		}

		regulator := "-"
		if o.regulator != nil && *o.regulator != "" {
			regulator = *o.regulator
		}

		dueDate := "-"
		if o.dueDate != nil {
			dueDate = o.dueDate.Format("2006-01-02")
		}

		obligationRows[i] = docgen.ObligationListRow{
			Area:        area,
			Source:      source,
			Requirement: requirement,
			Status:      formatObligationStatusString(o.status),
			Type:        formatObligationTypeString(o.oblType),
			Regulator:   regulator,
			Owner:       o.ownerName,
			DueDate:     dueDate,
		}
	}

	docData := docgen.ObligationListData{
		Title:            "Obligation Register",
		OrganizationName: orgName,
		CreatedAt:        publishedAt,
		TotalObligations: len(obligationRows),
		Rows:             obligationRows,
	}

	return probo.BuildObligationListDocument(docData)
}

func formatObligationStatusString(s string) string {
	switch s {
	case "NON_COMPLIANT":
		return "Non Compliant"
	case "PARTIALLY_COMPLIANT":
		return "Partially Compliant"
	case "COMPLIANT":
		return "Compliant"
	default:
		return s
	}
}

func formatObligationTypeString(t string) string {
	switch t {
	case "LEGAL":
		return "Legal"
	case "CONTRACTUAL":
		return "Contractual"
	default:
		return t
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
