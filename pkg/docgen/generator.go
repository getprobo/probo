// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package docgen

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"go.probo.inc/probo/pkg/coredata"
)

var (
	//go:embed template.html
	htmlTemplateContent string

	//go:embed processing_activities_template.html
	processingActivitiesTemplateContent string

	//go:embed data_protection_impact_assessments_template.html
	dataProtectionImpactAssessmentsTemplateContent string

	//go:embed transfer_impact_assessments_template.html
	transferImpactAssessmentsTemplateContent string

	//go:embed soa_template.html
	soaTemplateContent string

	templateFuncs = template.FuncMap{
		"now":                  func() time.Time { return time.Now() },
		"eq":                   func(a, b any) bool { return a == b },
		"add":                  func(a, b int) int { return a + b },
		"string":               func(v fmt.Stringer) string { return v.String() },
		"lower":                func(s string) string { return strings.ToLower(s) },
		"classificationString": func(c Classification) string { return string(c) },
		"boolToYesNo": func(b *bool) string {
			if b == nil {
				return ""
			}
			if *b {
				return "yes"
			}
			return "no"
		},
		"boolToYesNoDash": func(b *bool) string {
			if b == nil {
				return "-"
			}
			if *b {
				return "Yes"
			}
			return "No"
		},
		"formatContent": func(content string) template.HTML {
			md := goldmark.New(
				goldmark.WithExtensions(extension.Table),
				goldmark.WithRendererOptions(
					gmhtml.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(content), &buf); err != nil {
				return template.HTML(fmt.Sprintf("<p>%s</p>", html.EscapeString(content)))
			}
			return template.HTML(buf.String())
		},
		"imgTag": func(src, alt, class string) template.HTML {
			return template.HTML(fmt.Sprintf(`<img src="%s" alt="%s" class="%s">`, html.EscapeString(src), html.EscapeString(alt), html.EscapeString(class)))
		},
		"formatLawfulBasis": func(basis coredata.ProcessingActivityLawfulBasis) string {
			switch basis {
			case coredata.ProcessingActivityLawfulBasisConsent:
				return "Consent"
			case coredata.ProcessingActivityLawfulBasisContractualNecessity:
				return "Contractual Necessity"
			case coredata.ProcessingActivityLawfulBasisLegalObligation:
				return "Legal Obligation"
			case coredata.ProcessingActivityLawfulBasisLegitimateInterest:
				return "Legitimate Interest"
			case coredata.ProcessingActivityLawfulBasisPublicTask:
				return "Public Task"
			case coredata.ProcessingActivityLawfulBasisVitalInterests:
				return "Vital Interests"
			default:
				return basis.String()
			}
		},
		"formatSpecialOrCriminalData": func(data coredata.ProcessingActivitySpecialOrCriminalDatum) string {
			switch data {
			case coredata.ProcessingActivitySpecialOrCriminalDatumYes:
				return "Yes"
			case coredata.ProcessingActivitySpecialOrCriminalDatumNo:
				return "No"
			case coredata.ProcessingActivitySpecialOrCriminalDatumPossible:
				return "Possible"
			default:
				return data.String()
			}
		},
		"formatTransferSafeguard": func(safeguard *coredata.ProcessingActivityTransferSafeguard) string {
			if safeguard == nil {
				return ""
			}
			switch *safeguard {
			case coredata.ProcessingActivityTransferSafeguardStandardContractualClauses:
				return "Standard Contractual Clauses"
			case coredata.ProcessingActivityTransferSafeguardBindingCorporateRules:
				return "Binding Corporate Rules"
			case coredata.ProcessingActivityTransferSafeguardAdequacyDecision:
				return "Adequacy Decision"
			case coredata.ProcessingActivityTransferSafeguardDerogations:
				return "Derogations"
			case coredata.ProcessingActivityTransferSafeguardCodesOfConduct:
				return "Codes of Conduct"
			case coredata.ProcessingActivityTransferSafeguardCertificationMechanisms:
				return "Certification Mechanisms"
			default:
				return safeguard.String()
			}
		},
		"formatDPIANeeded": func(needed coredata.ProcessingActivityDataProtectionImpactAssessment) string {
			switch needed {
			case coredata.ProcessingActivityDataProtectionImpactAssessmentNeeded:
				return "Yes"
			case coredata.ProcessingActivityDataProtectionImpactAssessmentNotNeeded:
				return "No"
			default:
				return needed.String()
			}
		},
		"formatTIANeeded": func(needed coredata.ProcessingActivityTransferImpactAssessment) string {
			switch needed {
			case coredata.ProcessingActivityTransferImpactAssessmentNeeded:
				return "Yes"
			case coredata.ProcessingActivityTransferImpactAssessmentNotNeeded:
				return "No"
			default:
				return needed.String()
			}
		},
		"formatRole": func(role coredata.ProcessingActivityRole) string {
			switch role {
			case coredata.ProcessingActivityRoleController:
				return "Controller"
			case coredata.ProcessingActivityRoleProcessor:
				return "Processor"
			default:
				return role.String()
			}
		},
		"formatResidualRisk": func(risk *coredata.DataProtectionImpactAssessmentResidualRisk) string {
			if risk == nil {
				return ""
			}
			switch *risk {
			case coredata.DataProtectionImpactAssessmentResidualRiskLow:
				return "Low"
			case coredata.DataProtectionImpactAssessmentResidualRiskMedium:
				return "Medium"
			case coredata.DataProtectionImpactAssessmentResidualRiskHigh:
				return "High"
			default:
				return risk.String()
			}
		},
	}

	documentTemplate = template.Must(template.New("document").Funcs(templateFuncs).Parse(htmlTemplateContent))

	processingActivitiesTemplate = template.Must(template.New("processingActivities").Funcs(templateFuncs).Parse(processingActivitiesTemplateContent))

	dataProtectionImpactAssessmentsTemplate = template.Must(template.New("dataProtectionImpactAssessments").Funcs(templateFuncs).Parse(dataProtectionImpactAssessmentsTemplateContent))

	transferImpactAssessmentsTemplate = template.Must(template.New("transferImpactAssessments").Funcs(templateFuncs).Parse(transferImpactAssessmentsTemplateContent))

	stateOfApplicabilityTemplate = template.Must(template.New("state-of-applicability").Funcs(templateFuncs).Parse(soaTemplateContent))
)

type (
	Classification string

	DocumentData struct {
		Title                       string
		Content                     string
		Version                     int
		Classification              Classification
		Approvers                   []string
		Description                 string
		PublishedAt                 *time.Time
		Signatures                  []SignatureData
		CompanyHorizontalLogoBase64 string
	}

	SignatureData struct {
		SignedBy    string
		SignedAt    *time.Time
		State       coredata.DocumentVersionSignatureState
		RequestedAt time.Time
	}

	ProcessingActivityTableData struct {
		CompanyName                 string
		CompanyHorizontalLogoBase64 string
		Version                     int
		PublishedAt                 time.Time
		Activities                  []ProcessingActivityRowData
	}

	ProcessingActivityRowData struct {
		Name                                 string
		Purpose                              *string
		DataSubjectCategory                  *string
		PersonalDataCategory                 *string
		SpecialOrCriminalData                coredata.ProcessingActivitySpecialOrCriminalDatum
		ConsentEvidenceLink                  *string
		LawfulBasis                          coredata.ProcessingActivityLawfulBasis
		Recipients                           *string
		Location                             *string
		InternationalTransfers               bool
		TransferSafeguards                   *coredata.ProcessingActivityTransferSafeguard
		RetentionPeriod                      *string
		SecurityMeasures                     *string
		DataProtectionImpactAssessmentNeeded coredata.ProcessingActivityDataProtectionImpactAssessment
		TransferImpactAssessmentNeeded       coredata.ProcessingActivityTransferImpactAssessment
		LastReviewDate                       *time.Time
		NextReviewDate                       *time.Time
		Role                                 coredata.ProcessingActivityRole
		DataProtectionOfficerFullName        *string
		Vendors                              string
	}

	DataProtectionImpactAssessmentTableData struct {
		CompanyName                 string
		CompanyHorizontalLogoBase64 string
		Version                     int
		PublishedAt                 time.Time
		Assessments                 []DataProtectionImpactAssessmentRowData
	}

	DataProtectionImpactAssessmentRowData struct {
		ProcessingActivityName      string
		Description                 *string
		NecessityAndProportionality *string
		PotentialRisk               *string
		Mitigations                 *string
		ResidualRisk                *coredata.DataProtectionImpactAssessmentResidualRisk
	}

	TransferImpactAssessmentTableData struct {
		CompanyName                 string
		CompanyHorizontalLogoBase64 string
		Version                     int
		PublishedAt                 time.Time
		Assessments                 []TransferImpactAssessmentRowData
	}

	TransferImpactAssessmentRowData struct {
		ProcessingActivityName string
		DataSubjects           *string
		LegalMechanism         *string
		Transfer               *string
		LocalLawRisk           *string
		SupplementaryMeasures  *string
	}

	StateOfApplicabilityData struct {
		Title                       string
		OrganizationName            string
		CreatedAt                   time.Time
		TotalControls               int
		FrameworkGroups             []FrameworkControlGroup
		CompanyHorizontalLogoBase64 string
		Version                     int
		PublishedAt                 time.Time
		Approver                    string
	}

	FrameworkControlGroup struct {
		FrameworkName string
		Controls      []ControlData
	}

	ControlData struct {
		FrameworkName  string
		SectionTitle   string
		Name           string
		Applicability  *bool
		Justification  *string
		BestPractice   bool
		Regulatory     *bool
		Contractual    *bool
		RiskAssessment *bool
	}
)

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
)

func RenderHTML(data DocumentData) ([]byte, error) {
	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderProcessingActivitiesTableHTML(data ProcessingActivityTableData) ([]byte, error) {
	var buf bytes.Buffer
	if err := processingActivitiesTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute processing activities template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderDataProtectionImpactAssessmentsTableHTML(data DataProtectionImpactAssessmentTableData) ([]byte, error) {
	var buf bytes.Buffer
	if err := dataProtectionImpactAssessmentsTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute data protection impact assessments template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderTransferImpactAssessmentsTableHTML(data TransferImpactAssessmentTableData) ([]byte, error) {
	var buf bytes.Buffer
	if err := transferImpactAssessmentsTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute transfer impact assessments template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderStateOfApplicabilityHTML(data StateOfApplicabilityData) ([]byte, error) {
	var buf bytes.Buffer
	if err := stateOfApplicabilityTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute SOA template: %w", err)
	}

	return buf.Bytes(), nil
}
