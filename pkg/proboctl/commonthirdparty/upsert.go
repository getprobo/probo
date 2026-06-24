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

package commonthirdparty

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	"go.probo.inc/probo/pkg/slug"
)

func newCmdUpsert(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName                          string
		flagSlug                          string
		flagCategory                      string
		flagWebsiteURL                    string
		flagLegalName                     string
		flagHeadquarterAddress            string
		flagPrivacyPolicyURL              string
		flagServiceLevelAgreementURL      string
		flagServiceSoftwareAgreementURL   string
		flagDataProcessingAgreementURL    string
		flagBusinessAssociateAgreementURL string
		flagSubprocessorsListURL          string
		flagStatusPageURL                 string
		flagTermsOfServiceURL             string
		flagSecurityPageURL               string
		flagTrustPageURL                  string
		flagCertifications                []string
		flagDryRun                        bool
		flagEnrich                        bool
	)

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a common third party in the global catalog",
		Long: "Insert a new common third party or update an existing one keyed by " +
			"slug. Only --name and --category are required; every other field is " +
			"updated only when its flag is passed, so an existing row's other " +
			"columns are preserved.\n\n" +
			"Pass --enrich to queue the row for the async enrichment worker after " +
			"writing, so a minimal name/category row gets its profile, compliance " +
			"documents, owned domains, and logo filled in. Enrichment is expensive " +
			"(LLM + browser per row).",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Display name (required)")
	cmd.Flags().StringVar(&flagSlug, "slug", "", "Slug key (default: derived from --name)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Third party category (required)")
	cmd.Flags().StringVar(&flagWebsiteURL, "website-url", "", "Website URL")
	cmd.Flags().StringVar(&flagLegalName, "legal-name", "", "Legal name")
	cmd.Flags().StringVar(&flagHeadquarterAddress, "headquarter-address", "", "Headquarter address")
	cmd.Flags().StringVar(&flagPrivacyPolicyURL, "privacy-policy-url", "", "Privacy policy URL")
	cmd.Flags().StringVar(&flagServiceLevelAgreementURL, "service-level-agreement-url", "", "Service level agreement URL")
	cmd.Flags().StringVar(&flagServiceSoftwareAgreementURL, "service-software-agreement-url", "", "Service software agreement URL")
	cmd.Flags().StringVar(&flagDataProcessingAgreementURL, "data-processing-agreement-url", "", "Data processing agreement URL")
	cmd.Flags().StringVar(&flagBusinessAssociateAgreementURL, "business-associate-agreement-url", "", "Business associate agreement URL")
	cmd.Flags().StringVar(&flagSubprocessorsListURL, "subprocessors-list-url", "", "Subprocessors list URL")
	cmd.Flags().StringVar(&flagStatusPageURL, "status-page-url", "", "Status page URL")
	cmd.Flags().StringVar(&flagTermsOfServiceURL, "terms-of-service-url", "", "Terms of service URL")
	cmd.Flags().StringVar(&flagSecurityPageURL, "security-page-url", "", "Security page URL")
	cmd.Flags().StringVar(&flagTrustPageURL, "trust-page-url", "", "Trust page URL")
	cmd.Flags().StringSliceVar(&flagCertifications, "certifications", nil, "Certifications (repeatable)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the resulting row without writing")
	cmd.Flags().BoolVar(&flagEnrich, "enrich", false, "Queue the row for the async enrichment worker after writing")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("category")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		category := coredata.ThirdPartyCategory(flagCategory)
		if !category.IsValid() {
			return fmt.Errorf("invalid --category value %q", flagCategory)
		}

		partySlug := flagSlug
		if partySlug == "" {
			partySlug = slug.Make(flagName)
		}

		if partySlug == "" {
			return fmt.Errorf("cannot derive a slug from --name %q; pass --slug explicitly", flagName)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		var (
			party    coredata.CommonThirdParty
			inserted bool
		)

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				now := time.Now()

				existing := coredata.CommonThirdParty{}

				err := existing.LoadBySlug(ctx, tx, partySlug)
				switch {
				case err == nil:
					party = existing
				case errors.Is(err, coredata.ErrResourceNotFound):
					party = coredata.CommonThirdParty{
						ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
						Slug:           partySlug,
						Certifications: []string{},
						CreatedAt:      now,
					}
				default:
					return fmt.Errorf("cannot load common third party by slug: %w", err)
				}

				party.Name = flagName
				party.Category = category
				party.Slug = partySlug
				party.UpdatedAt = now

				applyFlag(cmd, "website-url", &party.WebsiteURL, flagWebsiteURL)
				applyFlag(cmd, "legal-name", &party.LegalName, flagLegalName)
				applyFlag(cmd, "headquarter-address", &party.HeadquarterAddress, flagHeadquarterAddress)
				applyFlag(cmd, "privacy-policy-url", &party.PrivacyPolicyURL, flagPrivacyPolicyURL)
				applyFlag(cmd, "service-level-agreement-url", &party.ServiceLevelAgreementURL, flagServiceLevelAgreementURL)
				applyFlag(cmd, "service-software-agreement-url", &party.ServiceSoftwareAgreementURL, flagServiceSoftwareAgreementURL)
				applyFlag(cmd, "data-processing-agreement-url", &party.DataProcessingAgreementURL, flagDataProcessingAgreementURL)
				applyFlag(cmd, "business-associate-agreement-url", &party.BusinessAssociateAgreementURL, flagBusinessAssociateAgreementURL)
				applyFlag(cmd, "subprocessors-list-url", &party.SubprocessorsListURL, flagSubprocessorsListURL)
				applyFlag(cmd, "status-page-url", &party.StatusPageURL, flagStatusPageURL)
				applyFlag(cmd, "terms-of-service-url", &party.TermsOfServiceURL, flagTermsOfServiceURL)
				applyFlag(cmd, "security-page-url", &party.SecurityPageURL, flagSecurityPageURL)
				applyFlag(cmd, "trust-page-url", &party.TrustPageURL, flagTrustPageURL)

				if cmd.Flags().Changed("certifications") {
					party.Certifications = flagCertifications
				}

				if flagDryRun {
					return nil
				}

				inserted, err = party.Upsert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot upsert common third party: %w", err)
				}

				// Arm enrichment explicitly rather than via the receiver:
				// Upsert sets enrichment_requested_at from the receiver only
				// on insert and leaves it untouched on conflict, so this is
				// the only path that re-arms an existing row uniformly. It
				// also resets the attempt budget for a fresh run.
				if flagEnrich {
					var parties coredata.CommonThirdParties

					if _, err := parties.RequestEnrichmentByIDs(ctx, tx, []gid.GID{party.ID}); err != nil {
						return fmt.Errorf("cannot queue enrichment: %w", err)
					}
				}

				return nil
			},
		); err != nil {
			return err
		}

		enrichSuffix := ""
		if flagEnrich {
			enrichSuffix = " (queued for enrichment)"
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would upsert common third party %q (slug %q, category %s)%s.\n", party.Name, party.Slug, party.Category, enrichSuffix)
			return nil
		}

		action := "Updated"
		if inserted {
			action = "Created"
		}

		_, _ = fmt.Fprintf(out, "%s common third party %s (%q, slug %q)%s.\n", action, party.ID.String(), party.Name, party.Slug, enrichSuffix)

		return nil
	}

	return cmd
}

// applyFlag overrides target with value when the named flag was passed.
// An empty value clears the column; an unset flag leaves it untouched, so
// an upsert can update one field without blanking the rest of the row.
func applyFlag(cmd *cobra.Command, name string, target **string, value string) {
	if !cmd.Flags().Changed(name) {
		return
	}

	if value == "" {
		*target = nil
		return
	}

	*target = &value
}
