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

// Command common-third-parties-import seeds the common_third_parties table from
// packages/vendors/data.json. It is idempotent: re-running upserts on conflict
// (lower(name)) so existing rows keep their id and created_at.
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

type thirdPartyData struct {
	Name                          string   `json:"name"`
	Description                   *string  `json:"description,omitempty"`
	Category                      *string  `json:"category,omitempty"`
	HeadquarterAddress            *string  `json:"headquarterAddress,omitempty"`
	LegalName                     *string  `json:"legalName,omitempty"`
	WebsiteURL                    *string  `json:"websiteUrl,omitempty"`
	PrivacyPolicyURL              *string  `json:"privacyPolicyUrl,omitempty"`
	ServiceLevelAgreementURL      *string  `json:"serviceLevelAgreementUrl,omitempty"`
	ServiceSoftwareAgreementURL   *string  `json:"serviceSoftwareAgreementUrl,omitempty"`
	DataProcessingAgreementURL    *string  `json:"dataProcessingAgreementUrl,omitempty"`
	BusinessAssociateAgreementURL *string  `json:"businessAssociateAgreementUrl,omitempty"`
	SubprocessorsListURL          *string  `json:"subprocessorsListUrl,omitempty"`
	Certifications                []string `json:"certifications,omitempty"`
	StatusPageURL                 *string  `json:"statusPageUrl,omitempty"`
	TermsOfServiceURL             *string  `json:"termsOfServiceUrl,omitempty"`
	SecurityPageURL               *string  `json:"securityPageUrl,omitempty"`
	TrustPageURL                  *string  `json:"trustPageUrl,omitempty"`
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
		"Path to the third-party data.json file",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	ctx := context.Background()

	thirdParties, err := loadThirdParties(dataPath)
	if err != nil {
		return fmt.Errorf("cannot load third-party data: %w", err)
	}

	pgClient, err := newPgClientFromDSN(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	fmt.Printf("importing %d common third parties from %s\n", len(thirdParties), dataPath)

	var inserted, updated int

	if err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			for _, tp := range thirdParties {
				party := coredata.CommonThirdParty{
					ID:                            gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
					Name:                          tp.Name,
					Description:                   tp.Description,
					Category:                      parseCategory(tp),
					HeadquarterAddress:            tp.HeadquarterAddress,
					LegalName:                     tp.LegalName,
					WebsiteURL:                    tp.WebsiteURL,
					PrivacyPolicyURL:              tp.PrivacyPolicyURL,
					ServiceLevelAgreementURL:      tp.ServiceLevelAgreementURL,
					ServiceSoftwareAgreementURL:   tp.ServiceSoftwareAgreementURL,
					DataProcessingAgreementURL:    tp.DataProcessingAgreementURL,
					BusinessAssociateAgreementURL: tp.BusinessAssociateAgreementURL,
					SubprocessorsListURL:          tp.SubprocessorsListURL,
					Certifications:                tp.Certifications,
					StatusPageURL:                 tp.StatusPageURL,
					TermsOfServiceURL:             tp.TermsOfServiceURL,
					SecurityPageURL:               tp.SecurityPageURL,
					TrustPageURL:                  tp.TrustPageURL,
					CreatedAt:                     now,
					UpdatedAt:                     now,
				}

				wasInserted, err := party.Upsert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot upsert common third party %q: %w", tp.Name, err)
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

	fmt.Printf("imported %d rows (%d inserted, %d updated)\n", len(thirdParties), inserted, updated)

	return nil
}

func loadThirdParties(path string) ([]thirdPartyData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var thirdParties []thirdPartyData
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&thirdParties); err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", path, err)
	}

	return thirdParties, nil
}

func parseCategory(tp thirdPartyData) coredata.VendorCategory {
	if tp.Category == nil || *tp.Category == "" {
		return coredata.VendorCategoryOther
	}

	var c coredata.VendorCategory
	if err := c.Scan(*tp.Category); err != nil {
		fmt.Fprintf(os.Stderr, "warning: third party %q has unknown category %q, falling back to OTHER\n", tp.Name, *tp.Category)
		return coredata.VendorCategoryOther
	}

	return c
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
	}

	opts := []pg.Option{pg.WithUnsecureTLS()}

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
