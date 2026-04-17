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

package vetting

// Output types for all vetting sub-agents. Each struct defines the JSON
// schema enforced via agent.WithOutputType on the corresponding sub-agent.

type (
	// --- Crawler ---

	DiscoveredURL struct {
		Category string `json:"category" jsonschema:"URL category: privacy_policy, terms_of_service, dpa, security, trust, compliance, status, subprocessors, sla, about, team, ai_policy, blog, careers, pricing, other"`
		URL      string `json:"url" jsonschema:"The discovered URL"`
	}

	CrawlerOutput struct {
		VendorName     string          `json:"vendor_name" jsonschema:"The vendor's display name as found on the website"`
		VendorDomain   string          `json:"vendor_domain" jsonschema:"The vendor's primary domain"`
		DiscoveredURLs []DiscoveredURL `json:"discovered_urls" jsonschema:"All categorized URLs discovered during crawling"`
		Notes          string          `json:"notes" jsonschema:"Observations about the site structure or crawl limitations"`
	}

	// --- Security ---

	SecurityCheckResult struct {
		Status  string `json:"status" jsonschema:"Check result: pass, warning, fail, or error"`
		Details string `json:"details" jsonschema:"Detailed findings for this check"`
	}

	WHOISResult struct {
		Registrar    string `json:"registrar" jsonschema:"Domain registrar name"`
		CreationDate string `json:"creation_date" jsonschema:"Domain creation date"`
		Organization string `json:"organization" jsonschema:"Registrant organization"`
		NameServers  string `json:"name_servers" jsonschema:"Comma-separated name servers"`
	}

	SecurityOutput struct {
		SSL      SecurityCheckResult `json:"ssl" jsonschema:"SSL/TLS certificate and protocol check"`
		Headers  SecurityCheckResult `json:"headers" jsonschema:"HTTP security headers check (HSTS, X-Frame-Options, etc.)"`
		DMARC    SecurityCheckResult `json:"dmarc" jsonschema:"DMARC email authentication policy check"`
		SPF      SecurityCheckResult `json:"spf" jsonschema:"SPF email authentication record check"`
		Breaches SecurityCheckResult `json:"breaches" jsonschema:"Known data breaches check via HIBP"`
		DNSSEC   SecurityCheckResult `json:"dnssec" jsonschema:"DNSSEC validation check"`
		CSP      SecurityCheckResult `json:"csp" jsonschema:"Content Security Policy analysis"`
		CORS     SecurityCheckResult `json:"cors" jsonschema:"CORS configuration check"`
		DNS      SecurityCheckResult `json:"dns" jsonschema:"DNS records analysis (A, MX, TXT, NS)"`
		WHOIS    WHOISResult         `json:"whois" jsonschema:"Domain WHOIS registration details"`
		Summary  string              `json:"summary" jsonschema:"Overall security posture summary"`
	}

	// --- Document Analyzer ---

	DocumentAnalysisOutput struct {
		DocumentType       string   `json:"document_type" jsonschema:"Type of document: privacy_policy, terms_of_service, dpa, sla, security_policy, acceptable_use, engagement_letter, other"`
		DocumentTitle      string   `json:"document_title" jsonschema:"Title of the document as shown on the page"`
		LastUpdated        string   `json:"last_updated" jsonschema:"Last updated date if found, empty string otherwise"`
		DataRetention      string   `json:"data_retention" jsonschema:"Data retention policy details"`
		DataLocations      []string `json:"data_locations" jsonschema:"Countries or regions where data is processed or stored"`
		GDPRIndicators     string   `json:"gdpr_indicators" jsonschema:"GDPR compliance indicators found"`
		CCPAIndicators     string   `json:"ccpa_indicators" jsonschema:"CCPA/CPRA compliance indicators found"`
		SecurityMeasures   string   `json:"security_measures" jsonschema:"Security measures described in the document"`
		BreachNotification string   `json:"breach_notification" jsonschema:"Breach notification commitments and timelines"`
		DataDeletion       string   `json:"data_deletion" jsonschema:"Data deletion procedures and timelines"`
		LiabilityCaps      string   `json:"liability_caps" jsonschema:"Liability limitations and caps"`
		Indemnification    string   `json:"indemnification" jsonschema:"Indemnification obligations"`
		Termination        string   `json:"termination" jsonschema:"Termination provisions and data return"`
		GoverningLaw       string   `json:"governing_law" jsonschema:"Governing law and jurisdiction"`
		PrivacyClauses     []string `json:"privacy_clauses" jsonschema:"Notable privacy contractual clauses found"`
		AIClauses          []string `json:"ai_clauses" jsonschema:"Notable AI-related contractual clauses found"`
		SubprocessorTerms  string   `json:"subprocessor_terms" jsonschema:"Sub-processor management terms (approval mechanism, notification)"`
		Summary            string   `json:"summary" jsonschema:"Key findings summary"`
		SourceURL          string   `json:"source_url" jsonschema:"URL of the analyzed document"`
	}

	// --- Compliance ---

	CertificationEntry struct {
		Name    string `json:"name" jsonschema:"Certification name (e.g. SOC 2 Type II, ISO 27001)"`
		Status  string `json:"status" jsonschema:"Certification status: current, in_progress, claimed_unverified, not_specified"`
		Details string `json:"details" jsonschema:"Additional details: audit date, certificate number, accreditation body"`
	}

	ComplianceOutput struct {
		Certifications      []CertificationEntry `json:"certifications" jsonschema:"All certifications and compliance frameworks found"`
		PenetrationTesting  string               `json:"penetration_testing" jsonschema:"Penetration testing practices (frequency, third-party firm)"`
		BugBounty           string               `json:"bug_bounty" jsonschema:"Bug bounty or responsible disclosure program details"`
		EncryptionStandards string               `json:"encryption_standards" jsonschema:"Encryption standards mentioned (AES-256, TLS 1.3, etc.)"`
		AuditReports        string               `json:"audit_reports" jsonschema:"Audit report availability (downloadable, on request, not available)"`
		OtherFrameworks     []string             `json:"other_frameworks" jsonschema:"Other frameworks or standards mentioned"`
		Summary             string               `json:"summary" jsonschema:"Overall compliance posture summary"`
		Sources             []string             `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Market Presence ---

	MarketOutput struct {
		NotableCustomers   []string `json:"notable_customers" jsonschema:"Notable customer names or logos identified"`
		CaseStudies        []string `json:"case_studies" jsonschema:"Case study summaries with customer names"`
		Partnerships       []string `json:"partnerships" jsonschema:"Strategic partnerships or integrations"`
		CompanySizeSignals string   `json:"company_size_signals" jsonschema:"Employee count, office locations, funding indicators"`
		FundingInfo        string   `json:"funding_info" jsonschema:"Known funding rounds, investors, or valuation signals"`
		MarketPosition     string   `json:"market_position" jsonschema:"Market positioning and competitive stance"`
		Summary            string   `json:"summary" jsonschema:"Overall market presence assessment"`
		Sources            []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Data Processing ---

	DataProcessingOutput struct {
		EncryptionAtRest     string   `json:"encryption_at_rest" jsonschema:"Encryption at rest details (algorithm, key size)"`
		EncryptionInTransit  string   `json:"encryption_in_transit" jsonschema:"Encryption in transit details (TLS version, cipher suites)"`
		KeyManagement        string   `json:"key_management" jsonschema:"Key management practices (HSM, rotation, customer-managed keys)"`
		RetentionPeriod      string   `json:"retention_period" jsonschema:"Data retention period and policy"`
		DeletionProcess      string   `json:"deletion_process" jsonschema:"Data deletion process and timeline"`
		CustomerControls     string   `json:"customer_controls" jsonschema:"Customer-facing data management controls"`
		DataLocations        []string `json:"data_locations" jsonschema:"Countries or regions where data is processed or stored"`
		TransferMechanisms   []string `json:"transfer_mechanisms" jsonschema:"Cross-border transfer mechanisms (SCCs, BCRs, adequacy decisions)"`
		DataResidency        string   `json:"data_residency" jsonschema:"Data residency options and restrictions"`
		BackupRecovery       string   `json:"backup_recovery" jsonschema:"Backup and disaster recovery for data"`
		Anonymization        string   `json:"anonymization" jsonschema:"Anonymization or pseudonymization practices"`
		DPAStatus            string   `json:"dpa_status" jsonschema:"DPA availability: available, available_on_request, not_found, behind_login"`
		ControllerProcessor  string   `json:"controller_processor" jsonschema:"Data processing role: controller, processor, subprocessor"`
		AuditRights          string   `json:"audit_rights" jsonschema:"Customer audit rights described"`
		SubprocessorApproval string   `json:"subprocessor_approval" jsonschema:"Sub-processor change approval mechanism"`
		BreachNotification   string   `json:"breach_notification" jsonschema:"Breach notification timeline and obligations"`
		DataReturn           string   `json:"data_return" jsonschema:"Data return and deletion on contract termination"`
		DSARHandling         string   `json:"dsar_handling" jsonschema:"DSAR handling capability and timeline"`
		DataMinimization     string   `json:"data_minimization" jsonschema:"Data minimization practices"`
		PurposeLimitation    string   `json:"purpose_limitation" jsonschema:"Purpose limitation commitments"`
		Rating               string   `json:"rating" jsonschema:"Overall data processing rating: Strong, Adequate, or Weak"`
		Summary              string   `json:"summary" jsonschema:"Key findings summary"`
		Sources              []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Subprocessor ---

	SubprocessorOutput struct {
		Subprocessors []Subprocessor `json:"subprocessors" jsonschema:"List of sub-processors discovered"`
		TotalCount    int            `json:"total_count" jsonschema:"Total number of sub-processors found"`
		Source        string         `json:"source" jsonschema:"URL where the sub-processor list was found"`
		IsComplete    bool           `json:"is_complete" jsonschema:"Whether the full list was extracted (false if pagination was incomplete)"`
		Notes         string         `json:"notes" jsonschema:"Observations about the sub-processor list"`
	}

	// --- Incident Response ---

	IncidentResponseOutput struct {
		IRPlan                 string   `json:"ir_plan" jsonschema:"Incident response plan documentation status"`
		NotificationTimeline   string   `json:"notification_timeline" jsonschema:"Breach notification timeline (e.g. 72 hours)"`
		NotificationMethod     string   `json:"notification_method" jsonschema:"How customers are notified of incidents"`
		ContractualObligations string   `json:"contractual_obligations" jsonschema:"Contractual IR obligations found"`
		StatusPageURL          string   `json:"status_page_url" jsonschema:"Status page URL if found"`
		StatusPageActive       bool     `json:"status_page_active" jsonschema:"Whether the status page is actively maintained"`
		UpdateFrequency        string   `json:"update_frequency" jsonschema:"How frequently status updates are provided during incidents"`
		PostMortems            string   `json:"post_mortems" jsonschema:"Post-mortem publication practices"`
		RemediationApproach    string   `json:"remediation_approach" jsonschema:"Approach to incident remediation"`
		RecentIncidents        []string `json:"recent_incidents" jsonschema:"Recent incidents found with dates and descriptions"`
		SecurityContact        string   `json:"security_contact" jsonschema:"Security contact email or reporting mechanism"`
		BugBounty              string   `json:"bug_bounty" jsonschema:"Bug bounty or vulnerability disclosure program"`
		Rating                 string   `json:"rating" jsonschema:"Overall incident response rating: Strong, Adequate, or Weak"`
		Summary                string   `json:"summary" jsonschema:"Key findings summary"`
		Sources                []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Business Continuity ---

	BusinessContinuityOutput struct {
		DRPlan             string   `json:"dr_plan" jsonschema:"Disaster recovery plan documentation status"`
		RTO                string   `json:"rto" jsonschema:"Recovery Time Objective"`
		RPO                string   `json:"rpo" jsonschema:"Recovery Point Objective"`
		TestingFrequency   string   `json:"testing_frequency" jsonschema:"DR testing frequency and last test date"`
		CloudProviders     []string `json:"cloud_providers" jsonschema:"Cloud infrastructure providers used"`
		MultiRegion        string   `json:"multi_region" jsonschema:"Multi-region deployment details"`
		Failover           string   `json:"failover" jsonschema:"Failover mechanisms and automation"`
		UptimeSLA          string   `json:"uptime_sla" jsonschema:"Uptime SLA commitment (e.g. 99.99%)"`
		SLACredits         string   `json:"sla_credits" jsonschema:"SLA credit or penalty structure"`
		HistoricalUptime   string   `json:"historical_uptime" jsonschema:"Historical uptime performance"`
		MaintenanceWindows string   `json:"maintenance_windows" jsonschema:"Scheduled maintenance window policy"`
		Regions            []string `json:"regions" jsonschema:"Geographic regions with infrastructure"`
		CDN                string   `json:"cdn" jsonschema:"CDN usage and provider"`
		BackupStrategy     string   `json:"backup_strategy" jsonschema:"Backup frequency, retention, and encryption"`
		BCPDocumented      string   `json:"bcp_documented" jsonschema:"Business continuity plan documentation status"`
		ISO22301           string   `json:"iso_22301" jsonschema:"ISO 22301 certification status"`
		Rating             string   `json:"rating" jsonschema:"Overall business continuity rating: Strong, Adequate, or Weak"`
		Summary            string   `json:"summary" jsonschema:"Key findings summary"`
		Sources            []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Professional Standing ---

	ProfessionalStandingOutput struct {
		VendorType      string   `json:"vendor_type" jsonschema:"Type of professional services firm: law_firm, accounting, consulting, audit, staffing, other"`
		Licensing       string   `json:"licensing" jsonschema:"Professional licensing details (bar admissions, CPA licenses)"`
		Memberships     []string `json:"memberships" jsonschema:"Industry body memberships (ABA, AICPA, Big Four network, etc.)"`
		Insurance       string   `json:"insurance" jsonschema:"Professional liability / E&O insurance coverage details"`
		TeamCredentials string   `json:"team_credentials" jsonschema:"Key team member qualifications and credentials"`
		COIPolicy       string   `json:"coi_policy" jsonschema:"Conflict of interest policy details"`
		ClientBase      string   `json:"client_base" jsonschema:"Client base signals (notable clients, industry focus)"`
		Rating          string   `json:"rating" jsonschema:"Overall professional standing rating: Strong, Adequate, Weak, or N/A"`
		KeyObservations string   `json:"key_observations" jsonschema:"Key observations about professional standing"`
		Sources         []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- AI Risk ---

	AIRiskOutput struct {
		AIInvolvement        string   `json:"ai_involvement" jsonschema:"AI involvement status: yes, no, or unclear"`
		UseCases             []string `json:"use_cases" jsonschema:"AI/ML use cases in the product or service"`
		AIPolicyURL          string   `json:"ai_policy_url" jsonschema:"URL to AI governance or responsible AI documentation"`
		ModelTransparency    string   `json:"model_transparency" jsonschema:"Model transparency and explainability findings"`
		BiasControls         string   `json:"bias_controls" jsonschema:"Bias detection and fairness measures"`
		CustomerDataTraining string   `json:"customer_data_training" jsonschema:"Whether customer data is used for model training"`
		OptOutAvailable      string   `json:"opt_out_available" jsonschema:"Whether training data opt-out is available"`
		TrainingDataDetails  string   `json:"training_data_details" jsonschema:"Training data governance details"`
		HumanOversight       string   `json:"human_oversight" jsonschema:"Human oversight mechanisms for AI decisions"`
		AIIncidentHandling   string   `json:"ai_incident_handling" jsonschema:"AI-specific incident handling procedures"`
		AutomatedDecisions   string   `json:"automated_decisions" jsonschema:"GDPR Art. 22 automated decision-making compliance"`
		EUAIAct              string   `json:"eu_ai_act" jsonschema:"EU AI Act awareness and compliance indicators"`
		Rating               string   `json:"rating" jsonschema:"Overall AI risk rating: Strong, Adequate, Weak, or N/A"`
		Summary              string   `json:"summary" jsonschema:"Key findings summary"`
		Sources              []string `json:"sources" jsonschema:"URLs visited during assessment"`
	}

	// --- Regulatory Compliance ---

	RegulatoryArticle struct {
		Article string `json:"article" jsonschema:"Article or section identifier (e.g. article_28, hipaa_security_rule)"`
		Status  string `json:"status" jsonschema:"Compliance status: compliant, partially_compliant, non_compliant, not_assessed, not_applicable"`
		Notes   string `json:"notes" jsonschema:"Evidence or reasoning for the status determination"`
	}

	RegulatoryFramework struct {
		Applicable    bool                `json:"applicable" jsonschema:"Whether this framework applies to the vendor"`
		OverallStatus string              `json:"overall_status" jsonschema:"Overall compliance status for this framework"`
		Articles      []RegulatoryArticle `json:"articles" jsonschema:"Per-article compliance assessment"`
		Notes         string              `json:"notes" jsonschema:"General notes about framework applicability"`
	}

	CrossBorderTransferInfo struct {
		Mechanisms    []string `json:"mechanisms" jsonschema:"Transfer mechanisms used (SCCs, BCRs, adequacy decisions)"`
		DataLocations []string `json:"data_locations" jsonschema:"Countries where data is stored or processed"`
		TIAEvidence   bool     `json:"tia_evidence" jsonschema:"Whether Transfer Impact Assessment evidence was found"`
	}

	RegulatoryComplianceOutput struct {
		GDPR                 RegulatoryFramework     `json:"gdpr" jsonschema:"GDPR compliance assessment"`
		HIPAA                RegulatoryFramework     `json:"hipaa" jsonschema:"HIPAA compliance assessment"`
		PCIDSS               RegulatoryFramework     `json:"pci_dss" jsonschema:"PCI DSS compliance assessment"`
		SOX                  RegulatoryFramework     `json:"sox" jsonschema:"SOX compliance assessment"`
		IndustrySpecific     []string                `json:"industry_specific" jsonschema:"Other industry-specific regulations found"`
		CrossBorderTransfers CrossBorderTransferInfo `json:"cross_border_transfers" jsonschema:"Cross-border data transfer assessment"`
		Gaps                 []string                `json:"gaps" jsonschema:"Identified compliance gaps"`
		Recommendations      []string                `json:"recommendations" jsonschema:"Recommended actions to address gaps"`
	}

	// --- Web Search ---

	WebSearchOutput struct {
		SecurityIncidents    string   `json:"security_incidents" jsonschema:"Known security incidents or breaches found"`
		RegulatoryActions    string   `json:"regulatory_actions" jsonschema:"Regulatory actions, fines, or investigations"`
		CustomerSentiment    string   `json:"customer_sentiment" jsonschema:"Customer reviews and sentiment summary"`
		RecentNews           string   `json:"recent_news" jsonschema:"Recent news coverage and press"`
		IndustryRecognition  string   `json:"industry_recognition" jsonschema:"Industry awards, analyst recognition, rankings"`
		ProfessionalStanding string   `json:"professional_standing" jsonschema:"Professional disciplinary actions or regulatory findings (for services firms)"`
		RedFlags             []string `json:"red_flags" jsonschema:"Red flags or concerning findings"`
		PositiveSignals      []string `json:"positive_signals" jsonschema:"Positive external signals"`
		Summary              string   `json:"summary" jsonschema:"Overall external research summary"`
		Sources              []string `json:"sources" jsonschema:"URLs visited during research"`
	}

	// --- Financial Stability ---

	FinancialStabilityOutput struct {
		CompanyAge        string   `json:"company_age" jsonschema:"Year founded and company age"`
		Funding           string   `json:"funding" jsonschema:"Funding history (rounds, amounts, investors)"`
		EmployeeCount     string   `json:"employee_count" jsonschema:"Estimated employee count and source"`
		RevenueSignals    string   `json:"revenue_signals" jsonschema:"Revenue indicators (ARR mentions, growth signals)"`
		CustomerBase      string   `json:"customer_base" jsonschema:"Customer base signals (count, notable names)"`
		LegalStanding     string   `json:"legal_standing" jsonschema:"Active lawsuits, regulatory issues, bankruptcy filings"`
		Ownership         string   `json:"ownership" jsonschema:"Ownership structure (public, PE-backed, founder-led, acquired)"`
		RiskSignals       []string `json:"risk_signals" jsonschema:"Financial risk signals identified"`
		OverallAssessment string   `json:"overall_assessment" jsonschema:"Overall financial stability: Strong, Adequate, Weak, or Concerning"`
		Confidence        string   `json:"confidence" jsonschema:"Assessment confidence level: High, Medium, or Low"`
		Notes             string   `json:"notes" jsonschema:"Additional observations"`
		Sources           []string `json:"sources" jsonschema:"URLs visited during research"`
	}

	// --- Code Security ---

	SecurityAdvisorySummary struct {
		Total        int    `json:"total" jsonschema:"Total number of security advisories"`
		Critical     int    `json:"critical" jsonschema:"Critical severity advisories"`
		High         int    `json:"high" jsonschema:"High severity advisories"`
		Medium       int    `json:"medium" jsonschema:"Medium severity advisories"`
		Low          int    `json:"low" jsonschema:"Low severity advisories"`
		AvgTimeToFix string `json:"avg_time_to_fix" jsonschema:"Average time to fix advisories"`
		Notes        string `json:"notes" jsonschema:"Additional context about advisories"`
	}

	CodeSecurityOutput struct {
		HasPublicRepos       bool                    `json:"has_public_repos" jsonschema:"Whether the vendor has public repositories"`
		GithubOrg            string                  `json:"github_org" jsonschema:"GitHub organization or user name"`
		MainRepos            []string                `json:"main_repos" jsonschema:"Main public repositories identified"`
		SecurityAdvisories   SecurityAdvisorySummary `json:"security_advisories" jsonschema:"Security advisory summary"`
		DependencyManagement string                  `json:"dependency_management" jsonschema:"Dependency management practices (Dependabot, Renovate, etc.)"`
		ReleaseCadence       string                  `json:"release_cadence" jsonschema:"Release frequency and last release date"`
		SecurityPolicy       string                  `json:"security_policy" jsonschema:"SECURITY.md or vulnerability disclosure policy"`
		CISecurity           string                  `json:"ci_security" jsonschema:"CI/CD security practices (SAST, DAST, container scanning)"`
		CodeSigning          string                  `json:"code_signing" jsonschema:"Code or release signing practices"`
		OpenSecurityIssues   string                  `json:"open_security_issues" jsonschema:"Open security-related issues or PRs"`
		License              string                  `json:"license" jsonschema:"Open source license type"`
		OverallAssessment    string                  `json:"overall_assessment" jsonschema:"Overall code security: Strong, Adequate, Weak, or Not_Applicable"`
		RiskSignals          []string                `json:"risk_signals" jsonschema:"Code security risk signals identified"`
		Notes                string                  `json:"notes" jsonschema:"Additional observations"`
		Sources              []string                `json:"sources" jsonschema:"URLs visited during research"`
	}

	// --- Vendor Comparison ---

	AlternativeVendor struct {
		Name           string   `json:"name" jsonschema:"Alternative vendor name"`
		Website        string   `json:"website" jsonschema:"Alternative vendor website URL"`
		Certifications []string `json:"certifications" jsonschema:"Visible certifications"`
		TrustCenter    bool     `json:"trust_center" jsonschema:"Whether a trust center page was found"`
		PrivacyPolicy  bool     `json:"privacy_policy" jsonschema:"Whether a privacy policy was found"`
		CompanySize    string   `json:"company_size" jsonschema:"Estimated company size"`
		SecurityScore  string   `json:"security_score" jsonschema:"Quick security impression: Strong, Adequate, or Weak"`
	}

	ComparisonSummary struct {
		SecurityMaturity  string `json:"security_maturity" jsonschema:"Relative security maturity vs alternatives"`
		CompliancePosture string `json:"compliance_posture" jsonschema:"Relative compliance posture vs alternatives"`
		MarketPosition    string `json:"market_position" jsonschema:"Relative market position vs alternatives"`
		Transparency      string `json:"transparency" jsonschema:"Relative transparency vs alternatives"`
	}

	VendorComparisonOutput struct {
		VendorCategory    string              `json:"vendor_category" jsonschema:"The vendor's product category"`
		AssessedVendor    string              `json:"assessed_vendor" jsonschema:"The vendor being assessed"`
		Alternatives      []AlternativeVendor `json:"alternatives" jsonschema:"Alternative vendors identified and evaluated"`
		ComparisonSummary ComparisonSummary   `json:"comparison_summary" jsonschema:"Summary comparison across dimensions"`
		VendorStrengths   []string            `json:"vendor_strengths" jsonschema:"Assessed vendor's strengths vs alternatives"`
		VendorWeaknesses  []string            `json:"vendor_weaknesses" jsonschema:"Assessed vendor's weaknesses vs alternatives"`
		OverallPosition   string              `json:"overall_position" jsonschema:"Vendor position: Above_Average, Average, or Below_Average"`
		Notes             string              `json:"notes" jsonschema:"Additional comparison notes"`
	}
)
