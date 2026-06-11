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

package thirdparty

import (
	"encoding/json"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// Provenance sources and per-field statuses recorded in the enrichment
// JSON. The "source" distinguishes values written by this enricher from
// values owned externally (curated seed data or a human edit), which the
// enricher must never overwrite.
const (
	enrichmentSourceEnrichment = "enrichment"
	enrichmentSourceExternal   = "external"

	enrichmentFieldStatusFound         = "found"
	enrichmentFieldStatusNotFound      = "not_found"
	enrichmentFieldStatusLowConfidence = "low_confidence"
	enrichmentFieldStatusExternal      = "exists_external"

	// Run-level status recorded at the top of the enrichment payload.
	enrichmentStatusDone    = "done"
	enrichmentStatusPartial = "partial"
	enrichmentStatusFailed  = "failed"
)

type (
	// EnrichedField is the per-field unit the enrichment agents return:
	// the resolved value plus a self-assessed confidence and the source
	// URL where it was verified. The worker applies a confidence
	// threshold before writing the value to its column.
	EnrichedField struct {
		Value      string  `json:"value" jsonschema:"The resolved value, or an empty string when not confidently found. Never guess."`
		Confidence float64 `json:"confidence" jsonschema:"Confidence from 0.0 to 1.0 that the value is correct. Use 0 when the value was not found."`
		SourceURL  string  `json:"source_url" jsonschema:"The URL where this value was verified, or an empty string."`
	}

	// CertificationsField is the list-valued counterpart of EnrichedField
	// used for the certifications array.
	CertificationsField struct {
		Values     []string `json:"values" jsonschema:"Certification or compliance framework names the vendor publicly claims (e.g. 'SOC 2 Type II', 'ISO 27001', 'HIPAA'). Empty when none are found."`
		Confidence float64  `json:"confidence" jsonschema:"Confidence from 0.0 to 1.0 in the certifications list. Use 0 when none are found."`
		SourceURL  string   `json:"source_url" jsonschema:"The URL where the certifications were found (trust or security page), or an empty string."`
	}

	// EnrichmentFieldMeta is the per-field provenance recorded in the
	// common_third_parties.enrichment JSON column.
	EnrichmentFieldMeta struct {
		Confidence float64   `json:"confidence"`
		SourceURL  string    `json:"source_url,omitempty"`
		Status     string    `json:"status"`
		Source     string    `json:"source"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	// EnrichmentMetadata is the full payload stored in the enrichment
	// JSON column: run-level bookkeeping plus per-field provenance keyed
	// by the column name.
	EnrichmentMetadata struct {
		Model       string                         `json:"model,omitempty"`
		AttemptedAt time.Time                      `json:"attempted_at"`
		Status      string                         `json:"status"`
		Error       string                         `json:"error,omitempty"`
		Fields      map[string]EnrichmentFieldMeta `json:"fields"`
	}
)

// scalarField describes one *string column the enricher can fill,
// pairing the agent's resolved value with accessors into the receiver.
type scalarField struct {
	name   string
	get    func(*coredata.CommonThirdParty) *string
	set    func(*coredata.CommonThirdParty, *string)
	result EnrichedField
}

// parseEnrichmentFields extracts the prior per-field provenance from a
// row's enrichment payload. A missing or malformed payload yields an
// empty map, which the merge treats as "no field is enrichment-owned".
func parseEnrichmentFields(raw json.RawMessage) map[string]EnrichmentFieldMeta {
	if len(raw) == 0 {
		return map[string]EnrichmentFieldMeta{}
	}

	var meta EnrichmentMetadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		return map[string]EnrichmentFieldMeta{}
	}

	if meta.Fields == nil {
		return map[string]EnrichmentFieldMeta{}
	}

	return meta.Fields
}

// applyScalarField merges one resolved scalar field into party, honoring
// the confidence threshold and prior provenance, and records the
// resulting per-field metadata. A value already present that this
// enricher did not write (curated seed data or a human edit) is left
// untouched. The column is written only when the value clears the
// threshold; below it the column keeps its prior value and the metadata
// records why nothing was written.
func applyScalarField(
	party *coredata.CommonThirdParty,
	meta map[string]EnrichmentFieldMeta,
	prior map[string]EnrichmentFieldMeta,
	field scalarField,
	threshold float64,
	now time.Time,
) {
	value := strings.TrimSpace(field.result.Value)
	sourceURL := strings.TrimSpace(field.result.SourceURL)

	existing := field.get(party)
	hasExisting := existing != nil && strings.TrimSpace(*existing) != ""

	priorMeta, hadPrior := prior[field.name]
	enrichmentOwned := hadPrior && priorMeta.Source == enrichmentSourceEnrichment

	if hasExisting && !enrichmentOwned {
		meta[field.name] = EnrichmentFieldMeta{
			Status:    enrichmentFieldStatusExternal,
			Source:    enrichmentSourceExternal,
			UpdatedAt: now,
		}

		return
	}

	if value != "" && field.result.Confidence >= threshold {
		v := value
		field.set(party, &v)
		meta[field.name] = EnrichmentFieldMeta{
			Confidence: field.result.Confidence,
			SourceURL:  sourceURL,
			Status:     enrichmentFieldStatusFound,
			Source:     enrichmentSourceEnrichment,
			UpdatedAt:  now,
		}

		return
	}

	status := enrichmentFieldStatusNotFound
	if value != "" {
		status = enrichmentFieldStatusLowConfidence
	}

	meta[field.name] = EnrichmentFieldMeta{
		Confidence: field.result.Confidence,
		SourceURL:  sourceURL,
		Status:     status,
		Source:     enrichmentSourceEnrichment,
		UpdatedAt:  now,
	}
}

// applyCertifications is the list-valued counterpart of
// applyScalarField for the certifications column.
func applyCertifications(
	party *coredata.CommonThirdParty,
	meta map[string]EnrichmentFieldMeta,
	prior map[string]EnrichmentFieldMeta,
	result CertificationsField,
	threshold float64,
	now time.Time,
) {
	const name = "certifications"

	values := normalizeCertifications(result.Values)
	sourceURL := strings.TrimSpace(result.SourceURL)

	hasExisting := len(party.Certifications) > 0

	priorMeta, hadPrior := prior[name]
	enrichmentOwned := hadPrior && priorMeta.Source == enrichmentSourceEnrichment

	if hasExisting && !enrichmentOwned {
		meta[name] = EnrichmentFieldMeta{
			Status:    enrichmentFieldStatusExternal,
			Source:    enrichmentSourceExternal,
			UpdatedAt: now,
		}

		return
	}

	if len(values) > 0 && result.Confidence >= threshold {
		party.Certifications = values
		meta[name] = EnrichmentFieldMeta{
			Confidence: result.Confidence,
			SourceURL:  sourceURL,
			Status:     enrichmentFieldStatusFound,
			Source:     enrichmentSourceEnrichment,
			UpdatedAt:  now,
		}

		return
	}

	status := enrichmentFieldStatusNotFound
	if len(values) > 0 {
		status = enrichmentFieldStatusLowConfidence
	}

	meta[name] = EnrichmentFieldMeta{
		Confidence: result.Confidence,
		SourceURL:  sourceURL,
		Status:     status,
		Source:     enrichmentSourceEnrichment,
		UpdatedAt:  now,
	}
}

// normalizeCertifications trims, drops blanks, and de-duplicates the
// certification names returned by the agent, preserving order.
func normalizeCertifications(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		out = append(out, v)
	}

	return out
}

// scalarFields returns the descriptor list pairing each *string column
// with the resolved value from the two agents. website_url comes from
// the company-profile agent; the document and page URLs come from the
// compliance-docs agent.
func scalarFields(
	company CompanyProfileResult,
	compliance ComplianceDocsResult,
) []scalarField {
	return []scalarField{
		{
			name:   "legal_name",
			get:    func(p *coredata.CommonThirdParty) *string { return p.LegalName },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.LegalName = v },
			result: company.LegalName,
		},
		{
			name:   "headquarter_address",
			get:    func(p *coredata.CommonThirdParty) *string { return p.HeadquarterAddress },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.HeadquarterAddress = v },
			result: company.HeadquarterAddress,
		},
		{
			name:   "website_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.WebsiteURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.WebsiteURL = v },
			result: company.WebsiteURL,
		},
		{
			name:   "privacy_policy_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.PrivacyPolicyURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.PrivacyPolicyURL = v },
			result: compliance.PrivacyPolicyURL,
		},
		{
			name:   "terms_of_service_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.TermsOfServiceURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.TermsOfServiceURL = v },
			result: compliance.TermsOfServiceURL,
		},
		{
			name:   "service_level_agreement_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.ServiceLevelAgreementURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.ServiceLevelAgreementURL = v },
			result: compliance.ServiceLevelAgreementURL,
		},
		{
			name:   "service_software_agreement_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.ServiceSoftwareAgreementURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.ServiceSoftwareAgreementURL = v },
			result: compliance.ServiceSoftwareAgreementURL,
		},
		{
			name:   "data_processing_agreement_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.DataProcessingAgreementURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.DataProcessingAgreementURL = v },
			result: compliance.DataProcessingAgreementURL,
		},
		{
			name:   "business_associate_agreement_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.BusinessAssociateAgreementURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.BusinessAssociateAgreementURL = v },
			result: compliance.BusinessAssociateAgreementURL,
		},
		{
			name:   "subprocessors_list_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.SubprocessorsListURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.SubprocessorsListURL = v },
			result: compliance.SubprocessorsListURL,
		},
		{
			name:   "status_page_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.StatusPageURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.StatusPageURL = v },
			result: compliance.StatusPageURL,
		},
		{
			name:   "security_page_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.SecurityPageURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.SecurityPageURL = v },
			result: compliance.SecurityPageURL,
		},
		{
			name:   "trust_page_url",
			get:    func(p *coredata.CommonThirdParty) *string { return p.TrustPageURL },
			set:    func(p *coredata.CommonThirdParty, v *string) { p.TrustPageURL = v },
			result: compliance.TrustPageURL,
		},
	}
}
