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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	"go.probo.inc/probo/pkg/slug"
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
	var flagData string

	cmd := &cobra.Command{
		Use:   "common-third-parties",
		Short: "Seed common third parties from a data.json file",
		Long: "Seed the common_third_parties table from a JSON file. " +
			"Re-running is safe: existing rows are upserted on conflict (slug) " +
			"so ids and created_at are preserved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := f.IOStreams.Out
			errOut := f.IOStreams.ErrOut
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

			return nil
		},
	}

	cmd.Flags().StringVar(&flagData, "data", "", "Path to the third-party data.json file")
	_ = cmd.MarkFlagRequired("data")

	return cmd
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
