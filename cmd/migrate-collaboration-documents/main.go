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

// Command migrate-collaboration-documents seeds every live document into its
// Automerge collaboration row. Archived and soft-deleted documents are
// skipped. An archived document is seeded on the first open after it is
// unarchived.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/pgconn"
)

type seedAction string

const (
	seedActionSeeded  seedAction = "seeded"
	seedActionSkipped seedAction = "skipped"
	seedActionDryRun  seedAction = "dry-run"
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
	flag.BoolVar(&dryRun, "dry-run", false, "list rows that would be seeded without writing")
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
	logger := log.NewLogger(
		log.WithName("migrate-collaboration-documents"),
		log.WithOutput(os.Stdout),
	)

	pgClient, err := pgconn.NewPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	svc := collaboration.NewService(pgClient, logger)

	logger.InfoCtx(ctx, "listing unseeded documents")

	ids, err := listUnseededDocumentIDs(ctx, pgClient)
	if err != nil {
		return err
	}

	logger.InfoCtx(
		ctx,
		"found unseeded documents",
		log.Int("total", len(ids)),
	)

	var (
		failures int
		seeded   int
		skipped  int
		planned  int
	)

	for i, idStr := range ids {
		action, err := seedOne(ctx, logger, svc, idStr, i+1, len(ids), dryRun)
		if err != nil {
			logger.ErrorCtx(
				ctx,
				"cannot seed document",
				log.String("document_id", idStr),
				log.Int("index", i+1),
				log.Int("total", len(ids)),
				log.Error(err),
			)

			failures++
			if !continueOnError {
				logFinished(ctx, logger, len(ids), seeded, skipped, planned, failures)

				return fmt.Errorf("stopped after %d failure(s)", failures)
			}

			continue
		}

		switch action {
		case seedActionSeeded:
			seeded++
		case seedActionSkipped:
			skipped++
		case seedActionDryRun:
			planned++
		}
	}

	logFinished(ctx, logger, len(ids), seeded, skipped, planned, failures)

	if failures > 0 {
		return fmt.Errorf("%d row(s) failed", failures)
	}

	return nil
}

func logFinished(
	ctx context.Context,
	logger *log.Logger,
	total int,
	seeded int,
	skipped int,
	planned int,
	failures int,
) {
	logger.InfoCtx(
		ctx,
		"finished seeding documents",
		log.Int("total", total),
		log.Int("seeded", seeded),
		log.Int("skipped", skipped),
		log.Int("dry_run", planned),
		log.Int("failed", failures),
	)
}

func listUnseededDocumentIDs(ctx context.Context, pgClient *pg.Client) ([]string, error) {
	var ids []string

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			rows, err := conn.Query(
				ctx,
				`
SELECT d.id
FROM documents d
LEFT JOIN collaboration_documents c ON c.document_id = d.id
WHERE d.archived_at IS NULL
	AND d.deleted_at IS NULL
	AND (c.document_id IS NULL OR c.seeded = FALSE)
ORDER BY d.id;
`,
			)
			if err != nil {
				return fmt.Errorf("cannot list documents: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					return fmt.Errorf("cannot scan document id: %w", err)
				}

				ids = append(ids, id)
			}

			return rows.Err()
		},
	)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func seedOne(
	ctx context.Context,
	logger *log.Logger,
	svc *collaboration.Service,
	idStr string,
	index int,
	total int,
	dryRun bool,
) (seedAction, error) {
	documentID, err := gid.ParseGID(idStr)
	if err != nil {
		return "", fmt.Errorf("invalid document id %q: %w", idStr, err)
	}

	scope := coredata.NewScopeFromObjectID(documentID)

	seeded, err := svc.Seeded(ctx, scope, documentID)
	if err != nil {
		return "", fmt.Errorf("cannot load collaboration seed state for %q: %w", idStr, err)
	}

	if seeded {
		logger.InfoCtx(
			ctx,
			"skipping document",
			log.String("document_id", idStr),
			log.String("reason", "already seeded"),
			log.Int("index", index),
			log.Int("total", total),
		)

		return seedActionSkipped, nil
	}

	if dryRun {
		logger.InfoCtx(
			ctx,
			"would seed document",
			log.String("document_id", idStr),
			log.Int("index", index),
			log.Int("total", total),
		)

		return seedActionDryRun, nil
	}

	logger.InfoCtx(
		ctx,
		"opening document",
		log.String("document_id", idStr),
		log.Int("index", index),
		log.Int("total", total),
	)

	opened, err := svc.Open(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, collaboration.ErrSeedInProgress) {
			logger.InfoCtx(
				ctx,
				"skipping document",
				log.String("document_id", idStr),
				log.String("reason", "seed in progress"),
				log.Int("index", index),
				log.Int("total", total),
			)

			return seedActionSkipped, nil
		}

		return "", fmt.Errorf("cannot seed document %q: %w", idStr, err)
	}

	if opened != nil && opened.Document != nil {
		_ = opened.Document.Close()
	}

	logger.InfoCtx(
		ctx,
		"seeded document",
		log.String("document_id", idStr),
		log.Int("index", index),
		log.Int("total", total),
	)

	return seedActionSeeded, nil
}
