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

package commonthirdparties

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	"go.probo.inc/probo/pkg/slug"
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
	Domains                       []string `json:"domains,omitempty"`
}

func NewCmdCommonThirdParties(f *cmdutil.Factory) *cobra.Command {
	var (
		flagData           string
		flagFetchLogos     bool
		flagS3Bucket       string
		flagS3Endpoint     string
		flagS3Region       string
		flagS3AccessKey    string
		flagS3SecretKey    string
		flagS3UsePathStyle bool
	)

	cmd := &cobra.Command{
		Use:   "common-third-parties",
		Short: "Seed common third parties from a data.json file",
		Long: "Seed the common_third_parties table from a JSON file. " +
			"Re-running is safe: existing rows are upserted on conflict (slug) " +
			"so ids and created_at are preserved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := f.IOStreams.Out
			errOut := f.IOStreams.ErrOut

			if flagFetchLogos && flagS3Bucket == "" {
				return fmt.Errorf("set --s3-bucket or AWS_S3_BUCKET when using --fetch-logos")
			}

			ctx := cmd.Context()

			thirdParties, err := loadThirdParties(flagData)
			if err != nil {
				return fmt.Errorf("cannot load third-party data: %w", err)
			}

			pgClient, err := f.PgClient()
			if err != nil {
				return fmt.Errorf("cannot create pg client: %w", err)
			}

			_, _ = fmt.Fprintf(out, "seeding %d common third parties from %s\n", len(thirdParties), flagData)

			var inserted, updated, domainsInserted, domainsUpdated int

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					now := time.Now()

					for _, tp := range thirdParties {
						party := coredata.CommonThirdParty{
							ID:                            gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
							Name:                          tp.Name,
							Slug:                          slug.Make(tp.Name),
							Category:                      parseCategory(errOut, tp),
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
							if err := party.LoadByName(ctx, tx, tp.Name); err != nil {
								return fmt.Errorf("cannot reload common third party %q: %w", tp.Name, err)
							}
						}

						for _, domain := range tp.Domains {
							d := coredata.CommonThirdPartyDomain{
								ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
								CommonThirdPartyID: party.ID,
								Domain:             domain,
								CreatedAt:          now,
								UpdatedAt:          now,
							}

							domainInserted, err := d.Upsert(ctx, tx)
							if err != nil {
								return fmt.Errorf("cannot upsert domain %q for %q: %w", domain, tp.Name, err)
							}

							if domainInserted {
								domainsInserted++
							} else {
								domainsUpdated++
							}
						}
					}

					return nil
				},
			); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(out, "seeded %d third parties (%d inserted, %d updated)\n", len(thirdParties), inserted, updated)
			_, _ = fmt.Fprintf(out, "seeded %d domains (%d inserted, %d updated)\n", domainsInserted+domainsUpdated, domainsInserted, domainsUpdated)

			if flagFetchLogos {
				if err := fetchAndStoreLogos(
					ctx, out, errOut, pgClient, thirdParties,
					flagS3Bucket, flagS3Endpoint, flagS3Region, flagS3AccessKey, flagS3SecretKey, flagS3UsePathStyle,
				); err != nil {
					return fmt.Errorf("cannot fetch logos: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagData, "data", "", "Path to the third-party data.json file")
	_ = cmd.MarkFlagRequired("data")
	cmd.Flags().BoolVar(&flagFetchLogos, "fetch-logos", false, "Fetch favicons and store them in S3")
	cmd.Flags().StringVar(&flagS3Bucket, "s3-bucket", os.Getenv("AWS_S3_BUCKET"), "S3 bucket name (default: AWS_S3_BUCKET env)")
	cmd.Flags().StringVar(&flagS3Endpoint, "s3-endpoint", os.Getenv("AWS_ENDPOINT_URL"), "S3 endpoint URL (default: AWS_ENDPOINT_URL env)")
	cmd.Flags().StringVar(&flagS3Region, "s3-region", os.Getenv("AWS_REGION"), "S3 region (default: AWS_REGION env)")
	cmd.Flags().StringVar(&flagS3AccessKey, "s3-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "S3 access key ID (default: AWS_ACCESS_KEY_ID env)")
	cmd.Flags().StringVar(&flagS3SecretKey, "s3-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "S3 secret access key (default: AWS_SECRET_ACCESS_KEY env)")
	cmd.Flags().BoolVar(&flagS3UsePathStyle, "s3-path-style", false, "Use S3 path-style addressing")

	return cmd
}

func fetchAndStoreLogos(
	ctx context.Context,
	out, errOut io.Writer,
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
		ua:   version.UserAgent("proboctl"),
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
			_, _ = fmt.Fprintf(errOut, "warning: cannot load %q, skipping logo: %v\n", tp.Name, err)
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
			_, _ = fmt.Fprintf(errOut, "warning: cannot inspect page for %q, trying default apple-touch-icon: %v\n", tp.Name, err)
		} else {
			logoURL, err = webinspect.FindLogoURL(pageInfo)
			if err != nil {
				_, _ = fmt.Fprintf(errOut, "warning: cannot find logo for %q, trying default apple-touch-icon: %v\n", tp.Name, err)
			}
		}

		parsed, err := url.Parse(*tp.WebsiteURL)
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "warning: cannot parse URL for %q, skipping logo: %v\n", tp.Name, err)
			failed++
			continue
		}

		var candidateURLs []string
		if logoURL != "" {
			candidateURLs = append(candidateURLs, logoURL)
		}
		base := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
		candidateURLs = append(
			candidateURLs,
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
			_, _ = fmt.Fprintf(errOut, "warning: cannot fetch logo for %q from any candidate URL\n", tp.Name)
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
			_, _ = fmt.Fprintf(errOut, "warning: cannot upload logo for %q to S3: %v\n", tp.Name, err)
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
			_, _ = fmt.Fprintf(errOut, "warning: cannot store logo for %q: %v\n", tp.Name, err)
			failed++
			continue
		}

		fetched++
		_, _ = fmt.Fprintf(out, "  fetched logo for %q\n", tp.Name)
	}

	_, _ = fmt.Fprintf(out, "logos: %d fetched, %d skipped, %d failed\n", fetched, skipped, failed)
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

func parseCategory(errOut io.Writer, tp thirdPartyData) coredata.ThirdPartyCategory {
	if tp.Category == nil || *tp.Category == "" {
		return coredata.ThirdPartyCategoryOther
	}

	var c coredata.ThirdPartyCategory
	if err := c.Scan(*tp.Category); err != nil {
		_, _ = fmt.Fprintf(errOut, "warning: third party %q has unknown category %q, falling back to OTHER\n", tp.Name, *tp.Category)
		return coredata.ThirdPartyCategoryOther
	}

	return c
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
