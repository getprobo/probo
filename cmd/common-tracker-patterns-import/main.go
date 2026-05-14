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
// table from the Open Cookie Database (https://github.com/jkwakman/Open-Cookie-Database).
// It clones the repository into a temporary directory, reads
// open-cookie-database.json, and upserts each entry. Re-running is safe:
// existing rows are updated on the unique constraint so ids and created_at
// are preserved.
//
// Entries are linked to the matching common_third_parties row (by
// case-insensitive name lookup on the platform key). Entries whose platform
// cannot be resolved are inserted with a NULL common_third_party_id.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	ocdRepoURL  = "https://github.com/jkwakman/Open-Cookie-Database.git"
	ocdJSONFile = "open-cookie-database.json"
)

type (
	ocdEntry struct {
		ID              string `json:"id"`
		Category        string `json:"category"`
		Cookie          string `json:"cookie"`
		Domain          string `json:"domain"`
		Description     string `json:"description"`
		RetentionPeriod string `json:"retentionPeriod"`
		DataController  string `json:"dataController"`
		PrivacyLink     string `json:"privacyLink"`
		WildcardMatch   string `json:"wildcardMatch"`
	}

	trackerPatternData struct {
		Pattern        string
		TrackerType    string
		MatchType      string
		ThirdPartyName *string
		Description    string
		MaxAgeSeconds  *int
		Confidence     float32
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var pgDSN string

	flag.StringVar(
		&pgDSN,
		"pg-dsn",
		os.Getenv("DATABASE_URL"),
		"PostgreSQL connection URL (default: DATABASE_URL env)",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	ctx := context.Background()

	fmt.Printf("cloning %s\n", ocdRepoURL)

	tmpDir, cleanup, err := cloneRepo()
	if err != nil {
		return fmt.Errorf("cannot clone repository: %w", err)
	}
	defer cleanup()

	patterns, err := loadPatternsFromOCD(tmpDir)
	if err != nil {
		return fmt.Errorf("cannot load tracker pattern data: %w", err)
	}

	pgClient, err := newPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	fmt.Printf("importing %d common tracker patterns from Open Cookie Database\n", len(patterns))

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

				_, wasInserted, err := pattern.Upsert(ctx, tx)
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

func cloneRepo() (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "ocd-*")
	if err != nil {
		return "", nil, fmt.Errorf("cannot create temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	_, err = git.PlainClone(
		tmpDir,
		false,
		&git.CloneOptions{
			URL:   ocdRepoURL,
			Depth: 1,
		},
	)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cannot clone %s: %w", ocdRepoURL, err)
	}

	return tmpDir, cleanup, nil
}

func loadPatternsFromOCD(dir string) ([]trackerPatternData, error) {
	f, err := os.Open(filepath.Join(dir, ocdJSONFile))
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", ocdJSONFile, err)
	}
	defer func() { _ = f.Close() }()

	var db map[string][]ocdEntry
	if err := json.NewDecoder(f).Decode(&db); err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", ocdJSONFile, err)
	}

	platforms := make([]string, 0, len(db))
	for k := range db {
		platforms = append(platforms, k)
	}
	sort.Strings(platforms)

	var patterns []trackerPatternData
	for _, platform := range platforms {
		for _, e := range db[platform] {
			if e.Cookie == "" {
				continue
			}

			matchType := "EXACT"
			if e.WildcardMatch == "1" {
				matchType = "GLOB"
			}

			cookiePattern := e.Cookie
			if matchType == "GLOB" && !strings.ContainsAny(cookiePattern, "*?") {
				cookiePattern += "*"
			}

			patterns = append(
				patterns,
				trackerPatternData{
					Pattern:        cookiePattern,
					TrackerType:    "COOKIE",
					MatchType:      matchType,
					ThirdPartyName: new(platform),
					Description:    e.Description,
					MaxAgeSeconds:  parseRetentionPeriod(e.RetentionPeriod),
					Confidence:     1.0,
				},
			)
		}
	}

	return patterns, nil
}

var retentionRe = regexp.MustCompile(`(?i)^(\d+)\s+(second|seconds|sec|secs|minute|minutes|mins|min|hour|hours|day|days|week|weeks|month|months|year|years)`)

func parseRetentionPeriod(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	lower := strings.ToLower(s)
	switch {
	case lower == "session" || lower == "sessions" || lower == "seesion" ||
		lower == "session cookie" || strings.HasPrefix(lower, "end of session"):
		return nil
	case lower == "varies" || lower == "various" || lower == "unknown" ||
		lower == "undefined" || lower == "persistent" || lower == "permanent" ||
		lower == "forever" || lower == "unlimited" || lower == "no expiration" ||
		lower == "local storage":
		return nil
	}

	m := retentionRe.FindStringSubmatch(s)
	if m == nil {
		return nil
	}

	n, err := strconv.Atoi(m[1])
	if err != nil {
		return nil
	}

	var multiplier int
	switch strings.ToLower(m[2]) {
	case "second", "seconds", "sec", "secs":
		multiplier = 1
	case "minute", "minutes", "mins", "min":
		multiplier = 60
	case "hour", "hours":
		multiplier = 3600
	case "day", "days":
		multiplier = 86400
	case "week", "weeks":
		multiplier = 604800
	case "month", "months":
		multiplier = 2592000
	case "year", "years":
		multiplier = 31536000
	default:
		return nil
	}

	result := n * multiplier
	return &result
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
