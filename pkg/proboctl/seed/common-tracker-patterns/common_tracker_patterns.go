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

package commontrackerpatterns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	"go.probo.inc/probo/pkg/slug"
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
		Domain         string
		Category       string
		Description    string
		MaxAgeSeconds  *int
		Confidence     float32
	}
)

func NewCmdCommonTrackerPatterns(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "common-tracker-patterns",
		Short: "Seed common tracker patterns from the Open Cookie Database",
		Long: "Seed the common_tracker_patterns table from the Open Cookie Database " +
			"(https://github.com/jkwakman/Open-Cookie-Database). " +
			"The repository is cloned into a temporary directory. " +
			"Re-running is safe: existing rows are upserted so ids and created_at are preserved. " +
			"Entries are linked to matching common_third_parties rows via slug lookup, " +
			"domain lookup, then auto-create.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := f.IOStreams.Out
			errOut := f.IOStreams.ErrOut
			ctx := cmd.Context()

			_, _ = fmt.Fprintf(out, "cloning %s\n", ocdRepoURL)

			tmpDir, cleanup, err := cloneRepo()
			if err != nil {
				return fmt.Errorf("cannot clone repository: %w", err)
			}
			defer cleanup()

			patterns, err := loadPatternsFromOCD(tmpDir)
			if err != nil {
				return fmt.Errorf("cannot load tracker pattern data: %w", err)
			}

			pgClient, err := f.PgClient()
			if err != nil {
				return fmt.Errorf("cannot create pg client: %w", err)
			}

			_, _ = fmt.Fprintf(out, "seeding %d common tracker patterns from Open Cookie Database\n", len(patterns))

			var inserted, updated, skipped int
			var partiesCreated int

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					now := time.Now()
					thirdPartyCache := make(map[string]*gid.GID)

					for _, p := range patterns {
						thirdPartyID, err := resolveThirdParty(ctx, tx, p, thirdPartyCache, now, &partiesCreated)
						if err != nil {
							return err
						}

						trackerType, err := parseTrackerType(p.TrackerType)
						if err != nil {
							_, _ = fmt.Fprintf(errOut, "warning: %v, skipping pattern %q\n", err, p.Pattern)
							skipped++
							continue
						}

						matchType, err := parseMatchType(p.MatchType)
						if err != nil {
							_, _ = fmt.Fprintf(errOut, "warning: %v, skipping pattern %q\n", err, p.Pattern)
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

			_, _ = fmt.Fprintf(
				out,
				"seeded %d patterns (%d inserted, %d updated, %d skipped, %d third parties auto-created)\n",
				len(patterns)-skipped,
				inserted,
				updated,
				skipped,
				partiesCreated,
			)

			return nil
		},
	}
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
					Domain:         e.Domain,
					Category:       e.Category,
					Description:    e.Description,
					MaxAgeSeconds:  parseRetentionPeriod(e.RetentionPeriod),
					Confidence:     1.0,
				},
			)
		}
	}

	return patterns, nil
}

func resolveThirdParty(
	ctx context.Context,
	tx pg.Tx,
	p trackerPatternData,
	cache map[string]*gid.GID,
	now time.Time,
	created *int,
) (*gid.GID, error) {
	if p.ThirdPartyName == nil || *p.ThirdPartyName == "" {
		return nil, nil
	}

	platformSlug := slug.Make(*p.ThirdPartyName)
	if platformSlug == "" {
		return nil, nil
	}
	if cached, ok := cache[platformSlug]; ok {
		return cached, nil
	}

	var party coredata.CommonThirdParty
	if err := party.LoadBySlug(ctx, tx, platformSlug); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, fmt.Errorf("cannot look up common third party by slug %q: %w", platformSlug, err)
		}
	} else {
		cache[platformSlug] = &party.ID
		return &party.ID, nil
	}

	domain := normalizeDomain(p.Domain)
	if domain != "" {
		var domainRow coredata.CommonThirdPartyDomain
		if err := domainRow.LoadByDomain(ctx, tx, domain); err != nil {
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, fmt.Errorf("cannot look up common third party by domain %q: %w", domain, err)
			}
		} else {
			if err := party.LoadByID(ctx, tx, domainRow.CommonThirdPartyID); err != nil {
				return nil, fmt.Errorf("cannot load common third party by ID %s: %w", domainRow.CommonThirdPartyID, err)
			}
			cache[platformSlug] = &party.ID
			return &party.ID, nil
		}
	}

	party = coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           *p.ThirdPartyName,
		Slug:           platformSlug,
		Category:       mapOCDCategory(p.Category),
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := party.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot auto-create common third party %q: %w", *p.ThirdPartyName, err)
	}

	if err := party.LoadBySlug(ctx, tx, platformSlug); err != nil {
		return nil, fmt.Errorf("cannot reload auto-created common third party %q: %w", *p.ThirdPartyName, err)
	}

	if domain != "" {
		d := coredata.CommonThirdPartyDomain{
			ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
			CommonThirdPartyID: party.ID,
			Domain:             domain,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := d.Upsert(ctx, tx); err != nil {
			return nil, fmt.Errorf("cannot upsert domain %q for %q: %w", domain, *p.ThirdPartyName, err)
		}
	}

	*created++
	cache[platformSlug] = &party.ID
	return &party.ID, nil
}

var domainValidRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*[a-zA-Z]$`)

func normalizeDomain(s string) string {
	s = strings.TrimSpace(s)

	s = strings.Map(
		func(r rune) rune {
			if r == '\u200b' || r == '\ufeff' {
				return -1
			}
			return r
		},
		s,
	)

	if idx := strings.Index(s, " or "); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimPrefix(s, "or ")

	if idx := strings.IndexByte(s, '('); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	}

	if strings.ContainsAny(s, "[]") {
		return ""
	}

	s = strings.TrimPrefix(s, ".")
	s = strings.TrimSuffix(s, ".")
	s = strings.TrimSpace(s)

	if s == "" || !domainValidRe.MatchString(s) {
		return ""
	}
	return s
}

func mapOCDCategory(s string) coredata.ThirdPartyCategory {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "analytics":
		return coredata.ThirdPartyCategoryAnalytics
	case "marketing":
		return coredata.ThirdPartyCategoryMarketing
	default:
		return coredata.ThirdPartyCategoryOther
	}
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

	result := min(n*multiplier, math.MaxInt32)
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
