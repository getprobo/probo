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

package vetting

import (
	"fmt"
	"strings"
)

// KnownFacts is the factual vendor profile fed to the assessment
// orchestrator as verified starting points, sourced from the catalog or a
// general-info profiling pass so sub-agents skip rediscovering basics.
type KnownFacts struct {
	Name                          string
	LegalName                     string
	HeadquarterAddress            string
	WebsiteURL                    string
	PrivacyPolicyURL              string
	TermsOfServiceURL             string
	ServiceLevelAgreementURL      string
	DataProcessingAgreementURL    string
	BusinessAssociateAgreementURL string
	SubprocessorsListURL          string
	SecurityPageURL               string
	TrustPageURL                  string
	StatusPageURL                 string
	Certifications                []string
	Domains                       []string
}

// IsEmpty reports whether there is nothing worth priming with. The name
// is disregarded: it is always available and no head start on its own.
func (k KnownFacts) IsEmpty() bool {
	return strings.TrimSpace(k.LegalName) == "" &&
		strings.TrimSpace(k.HeadquarterAddress) == "" &&
		strings.TrimSpace(k.WebsiteURL) == "" &&
		strings.TrimSpace(k.PrivacyPolicyURL) == "" &&
		strings.TrimSpace(k.TermsOfServiceURL) == "" &&
		strings.TrimSpace(k.ServiceLevelAgreementURL) == "" &&
		strings.TrimSpace(k.DataProcessingAgreementURL) == "" &&
		strings.TrimSpace(k.BusinessAssociateAgreementURL) == "" &&
		strings.TrimSpace(k.SubprocessorsListURL) == "" &&
		strings.TrimSpace(k.SecurityPageURL) == "" &&
		strings.TrimSpace(k.TrustPageURL) == "" &&
		strings.TrimSpace(k.StatusPageURL) == "" &&
		len(k.Certifications) == 0 &&
		len(k.Domains) == 0
}

// render formats the non-empty facts into a <known_facts> block prepended
// to the orchestrator's first message.
func (k KnownFacts) render() string {
	var b strings.Builder

	b.WriteString("<known_facts>\n")
	b.WriteString("Vendor facts already known from the catalog or a prior profiling pass. Treat them as verified starting points to confirm, not as final evidence.\n")

	writeFact(&b, "name", k.Name)
	writeFact(&b, "legal_name", k.LegalName)
	writeFact(&b, "headquarter_address", k.HeadquarterAddress)
	writeFact(&b, "website_url", k.WebsiteURL)
	writeFact(&b, "privacy_policy_url", k.PrivacyPolicyURL)
	writeFact(&b, "terms_of_service_url", k.TermsOfServiceURL)
	writeFact(&b, "service_level_agreement_url", k.ServiceLevelAgreementURL)
	writeFact(&b, "data_processing_agreement_url", k.DataProcessingAgreementURL)
	writeFact(&b, "business_associate_agreement_url", k.BusinessAssociateAgreementURL)
	writeFact(&b, "subprocessors_list_url", k.SubprocessorsListURL)
	writeFact(&b, "security_page_url", k.SecurityPageURL)
	writeFact(&b, "trust_page_url", k.TrustPageURL)
	writeFact(&b, "status_page_url", k.StatusPageURL)

	if len(k.Certifications) > 0 {
		writeFact(&b, "certifications", strings.Join(k.Certifications, ", "))
	}

	if len(k.Domains) > 0 {
		writeFact(&b, "owned_domains", strings.Join(k.Domains, ", "))
	}

	b.WriteString("</known_facts>")

	return b.String()
}

func writeFact(b *strings.Builder, label, value string) {
	v := strings.TrimSpace(value)
	if v == "" {
		return
	}

	fmt.Fprintf(b, "- %s: %s\n", label, v)
}
