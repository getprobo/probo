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

// Command link-document-versions ties the document versions that predate the
// link to the document history to the revision they were taken at.
//
// A version is a tag on a revision of the document's CRDT history. Versions
// published before the link existed carry no heads, so each one is matched to
// the state of the history that reproduces its content. A version that matches
// no state stays unlinked and stands alone in the history.
package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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
	flag.BoolVar(&dryRun, "dry-run", false, "report the matches without writing")
	flag.BoolVar(
		&continueOnError,
		"continue-on-error",
		false,
		"keep going when a document fails; exit non-zero if any failed",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	ctx := context.Background()

	pgClient, err := newPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	documentIDs, err := listDocumentIDs(ctx, pgClient)
	if err != nil {
		return err
	}

	service := collaboration.NewService(pgClient, log.NewLogger())

	var (
		matched   int
		synthetic int
		failures  int
	)

	for _, documentID := range documentIDs {
		scope := coredata.NewScopeFromObjectID(documentID)

		links, err := service.LinkVersionHeads(ctx, scope, documentID, dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "document %s: %v\n", documentID, err)

			failures++
			if !continueOnError {
				return fmt.Errorf("stopped after %d failure(s)", failures)
			}

			continue
		}

		for _, link := range links {
			if link.Heads == nil {
				synthetic++

				fmt.Printf("synthetic %s %d.%d (document %s)\n", link.VersionID, link.Major, link.Minor, documentID)

				continue
			}

			matched++

			verb := "linked"
			if dryRun {
				verb = "would link"
			}

			fmt.Printf("%s %s %d.%d to %s\n", verb, link.VersionID, link.Major, link.Minor, hex.EncodeToString(link.Heads))
		}
	}

	fmt.Printf("done: %d matched, %d synthetic, %d failed\n", matched, synthetic, failures)

	if failures > 0 {
		return fmt.Errorf("%d document(s) failed", failures)
	}

	return nil
}

func listDocumentIDs(ctx context.Context, pgClient *pg.Client) ([]gid.GID, error) {
	var ids []gid.GID

	err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		rows, err := conn.Query(
			ctx,
			`
SELECT DISTINCT document_id::text
FROM document_versions
WHERE collaboration_heads IS NULL
ORDER BY document_id;
`,
		)
		if err != nil {
			return fmt.Errorf("cannot list documents with unlinked versions: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				return fmt.Errorf("cannot scan document id: %w", err)
			}

			id, err := gid.ParseGID(raw)
			if err != nil {
				return fmt.Errorf("invalid document id %q: %w", raw, err)
			}

			ids = append(ids, id)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
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
