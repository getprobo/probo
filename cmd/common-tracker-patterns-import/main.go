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

// Command common-tracker-patterns-import seeds the common_tracker_patterns
// table from packages/common-tracker-patterns/data.json. It is idempotent:
// re-running upserts on the unique constraint so existing rows keep their id
// and created_at.
//
// Entries with a thirdPartyName are linked to the matching common_third_parties
// row (by case-insensitive name lookup). Entries without one are inserted with
// a NULL common_third_party_id.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type trackerPatternData struct {
	Pattern        string  `json:"pattern"`
	TrackerType    string  `json:"trackerType"`
	MatchType      string  `json:"matchType"`
	ThirdPartyName *string `json:"thirdPartyName,omitempty"`
	Description    string  `json:"description"`
	MaxAgeSeconds  *int    `json:"maxAgeSeconds,omitempty"`
	Confidence     float32 `json:"confidence"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		pgDSN    string
		dataPath string
	)

	flag.StringVar(
		&pgDSN,
		"pg-dsn",
		os.Getenv("DATABASE_URL"),
		"PostgreSQL connection URL (default: DATABASE_URL env)",
	)
	flag.StringVar(
		&dataPath,
		"data",
		"",
		"Path to the tracker-patterns data.json file",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	if dataPath == "" {
		return fmt.Errorf("set -data to the path of data.json")
	}

	ctx := context.Background()

	patterns, err := loadPatterns(dataPath)
	if err != nil {
		return fmt.Errorf("cannot load tracker pattern data: %w", err)
	}

	pgClient, err := newPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	fmt.Printf("importing %d common tracker patterns from %s\n", len(patterns), dataPath)

	var inserted, updated, skipped int

	if err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()
			thirdPartyCache := make(map[string]*gid.GID)

			for _, p := range patterns {
				var thirdPartyID *gid.GID

				if p.ThirdPartyName != nil {
					if cached, ok := thirdPartyCache[*p.ThirdPartyName]; ok {
						thirdPartyID = cached
					} else {
						var party coredata.CommonThirdParty
						if err := party.LoadByName(ctx, tx, *p.ThirdPartyName); err != nil {
							fmt.Fprintf(
								os.Stderr,
								"warning: cannot find third party %q for pattern %q, skipping link\n",
								*p.ThirdPartyName,
								p.Pattern,
							)
							thirdPartyCache[*p.ThirdPartyName] = nil
						} else {
							thirdPartyCache[*p.ThirdPartyName] = &party.ID
							thirdPartyID = &party.ID
						}
					}
				}

				trackerType, err := parseTrackerType(p.TrackerType)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: %v, skipping pattern %q\n", err, p.Pattern)
					skipped++
					continue
				}

				matchType, err := parseMatchType(p.MatchType)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: %v, skipping pattern %q\n", err, p.Pattern)
					skipped++
					continue
				}

				pattern := coredata.CommonTrackerPattern{
					ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
					CommonThirdPartyID: thirdPartyID,
					TrackerType:        trackerType,
					Pattern:            p.Pattern,
					MatchType:          matchType,
					Description:        p.Description,
					MaxAgeSeconds:      p.MaxAgeSeconds,
					Confidence:         p.Confidence,
					CreatedAt:          now,
					UpdatedAt:          now,
				}

				wasInserted, err := pattern.Upsert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot upsert common tracker pattern %q: %w", p.Pattern, err)
				}

				if wasInserted {
					inserted++
				} else {
					updated++
				}
			}

			return nil
		},
	); err != nil {
		return err
	}

	fmt.Printf(
		"imported %d patterns (%d inserted, %d updated, %d skipped)\n",
		len(patterns)-skipped,
		inserted,
		updated,
		skipped,
	)

	return nil
}

func loadPatterns(path string) ([]trackerPatternData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var patterns []trackerPatternData
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&patterns); err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", path, err)
	}

	return patterns, nil
}

func parseTrackerType(s string) (coredata.TrackerType, error) {
	switch s {
	case "COOKIE":
		return coredata.TrackerTypeCookie, nil
	case "LOCAL_STORAGE":
		return coredata.TrackerTypeLocalStorage, nil
	case "SESSION_STORAGE":
		return coredata.TrackerTypeSessionStorage, nil
	case "INDEXED_DB":
		return coredata.TrackerTypeIndexedDB, nil
	default:
		return "", fmt.Errorf("unknown tracker type %q", s)
	}
}

func parseMatchType(s string) (coredata.TrackerPatternMatchType, error) {
	switch s {
	case "EXACT":
		return coredata.TrackerPatternMatchTypeExact, nil
	case "GLOB":
		return coredata.TrackerPatternMatchTypeGlob, nil
	case "PREFIX":
		return coredata.TrackerPatternMatchTypePrefix, nil
	default:
		return "", fmt.Errorf("unknown match type %q", s)
	}
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN (check URL format)")
	}

	var opts []pg.Option

	switch u.Query().Get("sslmode") {
	case "", "disable":
	case "require":
		opts = append(opts, pg.WithUnsecureTLS())
	case "prefer":
		return nil, fmt.Errorf("unsupported sslmode %q (prefer fallback semantics are not supported)", u.Query().Get("sslmode"))
	default:
		return nil, fmt.Errorf("unsupported sslmode %q", u.Query().Get("sslmode"))
	}

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
