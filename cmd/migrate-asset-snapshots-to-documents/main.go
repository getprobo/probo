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

// Command migrate-asset-snapshots-to-documents creates documents and document
// versions from existing asset snapshots. For each organization that has asset
// snapshots, it generates an asset list document using the same ProseMirror
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

	return pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return migrate(ctx, tx, dryRun)
	})
}

type orgWithAssetSnapshots struct {
	organizationID   gid.GID
	tenantID         gid.TenantID
	organizationName string
}

type assetSnapshot struct {
	snapshotID  string
	publishedAt time.Time
}

func migrate(ctx context.Context, tx pg.Tx, dryRun bool) error {
	orgs, err := loadOrgsWithAssetSnapshots(ctx, tx)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("no organizations with asset snapshots to migrate")
		return nil
	}

	var stats struct {
		documents, versions int
	}

	for _, org := range orgs {
		snapshots, err := loadAssetSnapshots(ctx, tx, org.organizationID)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("would migrate org %s (%s) — %d asset snapshot(s)\n",
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
			`INSERT INTO generated_documents (organization_id, tenant_id, asset_list_document_id, created_at, updated_at)
VALUES (@organization_id, @tenant_id, @asset_list_document_id, @created_at, @updated_at)
ON CONFLICT (organization_id) DO UPDATE SET asset_list_document_id = @asset_list_document_id, updated_at = @updated_at`,
			pgx.NamedArgs{
				"organization_id":        org.organizationID,
				"tenant_id":              org.tenantID,
				"asset_list_document_id": documentID,
				"created_at":             now,
				"updated_at":             now,
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
					"title":           "Asset List",
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

func loadOrgsWithAssetSnapshots(ctx context.Context, tx pg.Tx) ([]orgWithAssetSnapshots, error) {
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
        WHERE gd.organization_id = o.id AND gd.asset_list_document_id IS NOT NULL
    )
    AND EXISTS (
        SELECT 1 FROM assets a
        WHERE a.organization_id = o.id AND a.snapshot_id IS NOT NULL
    )
ORDER BY o.created_at;
`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query organizations with asset snapshots: %w", err)
	}
	defer rows.Close()

	var result []orgWithAssetSnapshots
	for rows.Next() {
		var o orgWithAssetSnapshots
		var createdAt time.Time
		if err := rows.Scan(&o.organizationID, &o.tenantID, &o.organizationName, &createdAt); err != nil {
			return nil, fmt.Errorf("cannot scan organization: %w", err)
		}
		result = append(result, o)
	}

	return result, rows.Err()
}

func loadAssetSnapshots(ctx context.Context, tx pg.Tx, organizationID gid.GID) ([]assetSnapshot, error) {
	rows, err := tx.Query(
		ctx,
		`
SELECT DISTINCT
    s.id,
    s.created_at
FROM snapshots s
WHERE s.organization_id = @organization_id
    AND s.type = 'ASSETS'
ORDER BY s.created_at ASC;
`,
		pgx.NamedArgs{"organization_id": organizationID},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query asset snapshots for org %s: %w", organizationID, err)
	}
	defer rows.Close()

	var result []assetSnapshot
	for rows.Next() {
		var s assetSnapshot
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
    a.id,
    a.name,
    a.asset_type,
    a.amount,
    a.data_types_stored,
    COALESCE(p.full_name, '-')
FROM assets a
LEFT JOIN iam_membership_profiles p ON p.id = a.owner_profile_id
WHERE a.snapshot_id = @snapshot_id
ORDER BY a.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot assets: %w", err)
	}
	defer rows.Close()

	type assetInfo struct {
		id              string
		name            string
		assetType       string
		amount          int
		dataTypesStored string
		ownerName       string
	}

	var assets []assetInfo
	for rows.Next() {
		var a assetInfo
		if err := rows.Scan(&a.id, &a.name, &a.assetType, &a.amount, &a.dataTypesStored, &a.ownerName); err != nil {
			return "", fmt.Errorf("cannot scan asset: %w", err)
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	thirdPartyRows, err := tx.Query(
		ctx,
		`
SELECT
    av.asset_id,
    v.name
FROM asset_third_parties av
JOIN third_parties v ON v.id = av.third_party_id
WHERE av.snapshot_id = @snapshot_id
ORDER BY v.name ASC;
`,
		pgx.NamedArgs{"snapshot_id": snapshotID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load snapshot asset third_parties: %w", err)
	}
	defer thirdPartyRows.Close()

	thirdPartiesByAsset := make(map[string][]string)
	for thirdPartyRows.Next() {
		var assetID, thirdPartyName string
		if err := thirdPartyRows.Scan(&assetID, &thirdPartyName); err != nil {
			return "", fmt.Errorf("cannot scan thirdParty: %w", err)
		}
		thirdPartiesByAsset[assetID] = append(thirdPartiesByAsset[assetID], thirdPartyName)
	}
	if err := thirdPartyRows.Err(); err != nil {
		return "", err
	}

	assetRows := make([]docgen.AssetListRow, len(assets))
	for i, a := range assets {
		thirdParties := "-"
		if v, ok := thirdPartiesByAsset[a.id]; ok && len(v) > 0 {
			thirdParties = strings.Join(v, ", ")
		}

		assetRows[i] = docgen.AssetListRow{
			Name:            a.name,
			AssetType:       formatAssetTypeString(a.assetType),
			Amount:          a.amount,
			DataTypesStored: a.dataTypesStored,
			Owner:           a.ownerName,
			ThirdParties:    thirdParties,
		}
	}

	docData := docgen.AssetListData{
		Title:            "Asset List",
		OrganizationName: orgName,
		CreatedAt:        publishedAt,
		TotalAssets:      len(assetRows),
		Rows:             assetRows,
	}

	return probo.BuildAssetListDocument(docData)
}

func formatAssetTypeString(t string) string {
	switch t {
	case "PHYSICAL":
		return "Physical"
	case "VIRTUAL":
		return "Virtual"
	default:
		return t
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
