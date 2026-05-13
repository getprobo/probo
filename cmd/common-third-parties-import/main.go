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
// packages/thirdParties/data.json. It is idempotent: re-running upserts on conflict
// (lower(name)) so existing rows keep their id and created_at.
//
// When -fetch-logos is set, the tool inspects each third party's website to
// find the best available logo (SVG icon, apple-touch-icon, etc.) and stores
// it in S3 as a public file, linking it to each common third party via
// logo_file_id.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/version"
	"go.probo.inc/probo/pkg/webinspect"
)

type thirdPartyData struct {
	Name                          string   `json:"name"`
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
		pgDSN          string
		dataPath       string
		fetchLogos     bool
		s3Bucket       string
		s3Endpoint     string
		s3Region       string
		s3AccessKey    string
		s3SecretKey    string
		s3UsePathStyle bool
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
	flag.BoolVar(
		&fetchLogos,
		"fetch-logos",
		false,
		"Fetch favicons from Google and store them in S3",
	)
	flag.StringVar(
		&s3Bucket,
		"s3-bucket",
		os.Getenv("AWS_S3_BUCKET"),
		"S3 bucket name (default: AWS_S3_BUCKET env)",
	)
	flag.StringVar(
		&s3Endpoint,
		"s3-endpoint",
		os.Getenv("AWS_ENDPOINT_URL"),
		"S3 endpoint URL (default: AWS_ENDPOINT_URL env)",
	)
	flag.StringVar(
		&s3Region,
		"s3-region",
		os.Getenv("AWS_REGION"),
		"S3 region (default: AWS_REGION env)",
	)
	flag.StringVar(
		&s3AccessKey,
		"s3-access-key",
		os.Getenv("AWS_ACCESS_KEY_ID"),
		"S3 access key ID (default: AWS_ACCESS_KEY_ID env)",
	)
	flag.StringVar(
		&s3SecretKey,
		"s3-secret-key",
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"S3 secret access key (default: AWS_SECRET_ACCESS_KEY env)",
	)
	flag.BoolVar(
		&s3UsePathStyle,
		"s3-path-style",
		false,
		"Use S3 path-style addressing",
	)
	flag.Parse()

	if pgDSN == "" {
		return fmt.Errorf("set -pg-dsn or DATABASE_URL")
	}

	if fetchLogos && s3Bucket == "" {
		return fmt.Errorf("set -s3-bucket or AWS_S3_BUCKET when using -fetch-logos")
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

	if fetchLogos {
		if err := fetchAndStoreLogos(ctx, pgClient, thirdParties, s3Bucket, s3Endpoint, s3Region, s3AccessKey, s3SecretKey, s3UsePathStyle); err != nil {
			return fmt.Errorf("cannot fetch logos: %w", err)
		}
	}

	return nil
}

func fetchAndStoreLogos(
	ctx context.Context,
	pgClient *pg.Client,
	thirdParties []thirdPartyData,
	bucket, endpoint, region, accessKey, secretKey string,
	usePathStyle bool,
) error {
	s3Client := newS3Client(endpoint, region, accessKey, secretKey, usePathStyle)
	fileMgr := filemanager.NewService(s3Client)
	httpClient := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	httpClient.Transport = &userAgentTransport{
		next: httpClient.Transport,
		ua:   version.UserAgent("common-third-parties-import"),
	}
	scope := coredata.NewScope(gid.NilTenant)

	var fetched, skipped, failed int

	for _, tp := range thirdParties {
		if tp.WebsiteURL == nil || *tp.WebsiteURL == "" {
			skipped++
			continue
		}

		var party coredata.CommonThirdParty
		if err := pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return party.LoadByName(ctx, conn, tp.Name)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot load %q, skipping logo: %v\n", tp.Name, err)
			failed++
			continue
		}

		if party.LogoFileID != nil {
			skipped++
			continue
		}

		var logoURL string
		pageInfo, err := webinspect.Parse(ctx, httpClient, *tp.WebsiteURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot inspect page for %q, trying default apple-touch-icon: %v\n", tp.Name, err)
		} else {
			logoURL, err = webinspect.FindLogoURL(pageInfo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: cannot find logo for %q, trying default apple-touch-icon: %v\n", tp.Name, err)
			}
		}

		parsed, err := url.Parse(*tp.WebsiteURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot parse URL for %q, skipping logo: %v\n", tp.Name, err)
			failed++
			continue
		}

		var candidateURLs []string
		if logoURL != "" {
			candidateURLs = append(candidateURLs, logoURL)
		}
		base := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
		candidateURLs = append(candidateURLs,
			base+"/apple-touch-icon.png",
			base+"/apple-touch-icon-precomposed.png",
			"https://logo.debounce.com/"+parsed.Host,
		)

		var (
			body        []byte
			contentType string
		)
		for _, candidate := range candidateURLs {
			resp, err := httpClient.Get(candidate)
			if err != nil {
				continue
			}

			b, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if err != nil || resp.StatusCode != http.StatusOK || len(b) == 0 {
				continue
			}

			body = b
			contentType = resp.Header.Get("Content-Type")
			break
		}

		if len(body) == 0 {
			fmt.Fprintf(os.Stderr, "warning: cannot fetch logo for %q from any candidate URL\n", tp.Name)
			failed++
			continue
		}

		if contentType == "" {
			contentType = "image/png"
		}

		objectKey, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("cannot generate object key: %w", err)
		}

		now := time.Now()
		fileID := gid.New(gid.NilTenant, coredata.FileEntityType)

		fileRecord := &coredata.File{
			ID:             fileID,
			OrganizationID: gid.Nil,
			BucketName:     bucket,
			MimeType:       contentType,
			FileName:       tp.Name + "-logo" + webinspect.ExtensionForMIME(contentType),
			FileKey:        objectKey.String(),
			FileSize:       int64(len(body)),
			Visibility:     coredata.FileVisibilityPublic,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if _, err := fileMgr.PutFile(ctx, fileRecord, bytes.NewReader(body), map[string]string{
			"type":                  "common-third-party-logo",
			"common-third-party-id": party.ID.String(),
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot upload logo for %q to S3: %v\n", tp.Name, err)
			failed++
			continue
		}

		if err := pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			if err := fileRecord.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file record: %w", err)
			}

			party.LogoFileID = &fileID
			party.UpdatedAt = now
			if err := party.UpdateLogoFileID(ctx, tx); err != nil {
				return fmt.Errorf("cannot update logo_file_id: %w", err)
			}

			return nil
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot store logo for %q: %v\n", tp.Name, err)
			failed++
			continue
		}

		fetched++
		fmt.Printf("  fetched logo for %q\n", tp.Name)
	}

	fmt.Printf("logos: %d fetched, %d skipped, %d failed\n", fetched, skipped, failed)
	return nil
}

func newS3Client(endpoint, region, accessKey, secretKey string, usePathStyle bool) *s3.Client {
	if region == "" {
		region = "us-east-2"
	}

	cfg := aws.Config{
		Region: region,
	}

	if accessKey != "" && secretKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	}

	if endpoint != "" {
		cfg.BaseEndpoint = &endpoint
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})
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

func parseCategory(tp thirdPartyData) coredata.ThirdPartyCategory {
	if tp.Category == nil || *tp.Category == "" {
		return coredata.ThirdPartyCategoryOther
	}

	var c coredata.ThirdPartyCategory
	if err := c.Scan(*tp.Category); err != nil {
		fmt.Fprintf(os.Stderr, "warning: third party %q has unknown category %q, falling back to OTHER\n", tp.Name, *tp.Category)
		return coredata.ThirdPartyCategoryOther
	}

	return c
}

func newPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN (check URL format)")
	}

	var opts []pg.Option

	switch u.Query().Get("sslmode") {
	case "", "disable":
		// plain connection, no TLS
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

type userAgentTransport struct {
	next http.RoundTripper
	ua   string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	return t.next.RoundTrip(req)
}
