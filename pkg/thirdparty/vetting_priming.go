// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/vetting"
)

// resolveKnownFacts gathers the vendor facts used to prime a vetting run:
// a catalog-linked third party reuses the catalog row (no LLM), otherwise
// the profiler runs the general-info agents on demand. Best-effort: any
// failure is logged and vetting proceeds without facts. Never writes the
// resolved facts back to the common catalog.
func (h *vettingHandler) resolveKnownFacts(
	ctx context.Context,
	thirdParty *coredata.ThirdParty,
) *vetting.KnownFacts {
	if thirdParty.CommonThirdPartyID != nil {
		facts, err := h.knownFactsFromCatalog(ctx, *thirdParty.CommonThirdPartyID)
		if err != nil {
			h.logger.WarnCtx(
				ctx,
				"cannot load common third party facts for vetting priming",
				log.Error(err),
				log.String("third_party_id", thirdParty.ID.String()),
			)
		} else if !facts.IsEmpty() {
			return facts
		}
	}

	if h.profiler == nil || thirdParty.VettingWebsiteURL == nil {
		return nil
	}

	profile, err := h.profiler.Profile(
		ctx,
		GeneralInfoInput{
			Name:       thirdParty.Name,
			WebsiteURL: ref.UnrefOrZero(thirdParty.VettingWebsiteURL),
			LegalName:  ref.UnrefOrZero(thirdParty.LegalName),
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			ctx,
			"general-info profiling failed for vetting priming",
			log.Error(err),
			log.String("third_party_id", thirdParty.ID.String()),
		)

		return nil
	}

	facts := knownFactsFromProfile(thirdParty.Name, profile)
	if facts.IsEmpty() {
		return nil
	}

	return &facts
}

// knownFactsFromCatalog loads the catalog row and its owned domains and
// maps them to known facts.
func (h *vettingHandler) knownFactsFromCatalog(
	ctx context.Context,
	commonThirdPartyID gid.GID,
) (*vetting.KnownFacts, error) {
	party := &coredata.CommonThirdParty{}

	var domains coredata.CommonThirdPartyDomains

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := party.LoadByID(ctx, conn, commonThirdPartyID); err != nil {
				return fmt.Errorf("cannot load common third party: %w", err)
			}

			if err := domains.LoadByCommonThirdPartyID(ctx, conn, commonThirdPartyID); err != nil {
				return fmt.Errorf("cannot load common third party domains: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	domainStrings := make([]string, 0, len(domains))
	for _, d := range domains {
		domainStrings = append(domainStrings, d.Domain)
	}

	facts := knownFactsFromCommonThirdParty(party, domainStrings)

	return &facts, nil
}

// knownFactsFromCommonThirdParty maps a catalog row to vetting known facts.
func knownFactsFromCommonThirdParty(
	party *coredata.CommonThirdParty,
	domains []string,
) vetting.KnownFacts {
	return vetting.KnownFacts{
		Name:                          party.Name,
		LegalName:                     ref.UnrefOrZero(party.LegalName),
		HeadquarterAddress:            ref.UnrefOrZero(party.HeadquarterAddress),
		WebsiteURL:                    ref.UnrefOrZero(party.WebsiteURL),
		PrivacyPolicyURL:              ref.UnrefOrZero(party.PrivacyPolicyURL),
		TermsOfServiceURL:             ref.UnrefOrZero(party.TermsOfServiceURL),
		ServiceLevelAgreementURL:      ref.UnrefOrZero(party.ServiceLevelAgreementURL),
		DataProcessingAgreementURL:    ref.UnrefOrZero(party.DataProcessingAgreementURL),
		BusinessAssociateAgreementURL: ref.UnrefOrZero(party.BusinessAssociateAgreementURL),
		SubprocessorsListURL:          ref.UnrefOrZero(party.SubprocessorsListURL),
		SecurityPageURL:               ref.UnrefOrZero(party.SecurityPageURL),
		TrustPageURL:                  ref.UnrefOrZero(party.TrustPageURL),
		StatusPageURL:                 ref.UnrefOrZero(party.StatusPageURL),
		Certifications:                party.Certifications,
		Domains:                       domains,
	}
}

// knownFactsFromProfile maps an on-demand general-info profile to known
// facts.
func knownFactsFromProfile(name string, profile GeneralInfoProfile) vetting.KnownFacts {
	company := profile.Company
	compliance := profile.Compliance

	return vetting.KnownFacts{
		Name:                          name,
		LegalName:                     strings.TrimSpace(company.LegalName.Value),
		HeadquarterAddress:            strings.TrimSpace(company.HeadquarterAddress.Value),
		WebsiteURL:                    profile.WebsiteURL,
		PrivacyPolicyURL:              strings.TrimSpace(compliance.PrivacyPolicyURL.Value),
		TermsOfServiceURL:             strings.TrimSpace(compliance.TermsOfServiceURL.Value),
		ServiceLevelAgreementURL:      strings.TrimSpace(compliance.ServiceLevelAgreementURL.Value),
		DataProcessingAgreementURL:    strings.TrimSpace(compliance.DataProcessingAgreementURL.Value),
		BusinessAssociateAgreementURL: strings.TrimSpace(compliance.BusinessAssociateAgreementURL.Value),
		SubprocessorsListURL:          strings.TrimSpace(compliance.SubprocessorsListURL.Value),
		SecurityPageURL:               strings.TrimSpace(compliance.SecurityPageURL.Value),
		TrustPageURL:                  strings.TrimSpace(compliance.TrustPageURL.Value),
		StatusPageURL:                 strings.TrimSpace(compliance.StatusPageURL.Value),
		Certifications:                compliance.Certifications.Values,
		Domains:                       ownedDomainStrings(profile.Domains),
	}
}

// ownedDomainStrings collects the resolved domain names, dropping blanks.
func ownedDomainStrings(result DomainsResult) []string {
	out := make([]string, 0, len(result.Domains))
	for _, d := range result.Domains {
		domain := strings.TrimSpace(d.Domain)
		if domain == "" {
			continue
		}

		out = append(out, domain)
	}

	return out
}
