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

// Command migrate-document-versions-markdown rewrites document_versions.content
// from legacy markdown to ProseMirror JSON (see pkg/coredata/document_version.go).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		pgDSN           string
		dryRun          bool
		continueOnError bool
	)

	flag.StringVar(
		&pgDSN,
		"pg-dsn",
		os.Getenv("DATABASE_URL"),
		"PostgreSQL connection URL (default: DATABASE_URL env)",
	)
	flag.BoolVar(&dryRun, "dry-run", false, "list rows that would be migrated without writing")
	flag.BoolVar(
		&continueOnError,
		"continue-on-error",
		false,
		"keep going when a row fails; exit non-zero if any failed",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return fmt.Errorf("cannot connect to database: %w", err)
	}
	defer pool.Close()

	rows, err := pool.Query(
		ctx,
		`
SELECT id::text
FROM document_versions
WHERE content IS NOT NULL
	AND btrim(content) <> ''
ORDER BY id;
`,
	)
	if err != nil {
		return fmt.Errorf("cannot list document versions: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("cannot scan document version id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate document versions: %w", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	var failures int
	for _, idStr := range ids {
		if err := migrateOne(ctx, pgxConn, idStr, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			failures++
			if !continueOnError {
				return fmt.Errorf("stopped after %d failure(s)", failures)
			}
			continue
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d row(s) failed", failures)
	}

	fmt.Println("done")
	return nil
}

func migrateOne(ctx context.Context, conn pg.Conn, idStr string, dryRun bool) error {
	versionID, err := gid.ParseGID(idStr)
	if err != nil {
		return fmt.Errorf("invalid document version id %q: %w", idStr, err)
	}

	dv := &coredata.DocumentVersion{}
	if err := dv.LoadByID(ctx, conn, coredata.NewNoScope(), versionID); err != nil {
		return fmt.Errorf("cannot load document version %q: %w", idStr, err)
	}

	if isProseMirrorDocJSON(dv.Content) {
		fmt.Printf("skip %s (already ProseMirror doc JSON)\n", idStr)
		return nil
	}

	doc, err := prosemirror.ParseMarkdown(dv.Content)
	if err != nil {
		return fmt.Errorf("cannot parse markdown for %q: %w", idStr, err)
	}

	out, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("cannot marshal prosemirror for %q: %w", idStr, err)
	}

	if dryRun {
		fmt.Printf("would migrate %s successfully\n", idStr)
		return nil
	}

	dv.Content = string(out)
	dv.UpdatedAt = time.Now()

	if err := dv.Update(ctx, conn, coredata.NewNoScope()); err != nil {
		return fmt.Errorf("cannot update document version %q: %w", idStr, err)
	}

	fmt.Printf("updated %s\n", idStr)
	return nil
}

func isProseMirrorDocJSON(s string) bool {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(s), &probe); err != nil {
		return false
	}
	return probe.Type == "doc"
}
