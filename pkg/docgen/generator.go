// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/prosemirror"
)

var (
	//go:embed mermaid.min.js
	mermaidJSSource string

	//go:embed template.html
	htmlTemplateContent string

	//go:embed signature_page_template.html
	signaturePageTemplateContent string

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
		"derefString": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
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

	signaturePageTemplate = template.Must(template.New("signaturePage").Funcs(templateFuncs).Parse(signaturePageTemplateContent))
)

type (
	Classification string

	DocumentData struct {
		Title                       string
		Content                     json.RawMessage // ProseMirror/Tiptap document JSON; use ProseMirrorJSONToHTML for HTML
		Major                       int
		Minor                       int
		Classification              Classification
		Approvers                   []string
		Description                 string
		PublishedAt                 *time.Time
		Signatures                  []SignatureData
		CompanyHorizontalLogoBase64 string
		MermaidJS                   template.JS
		Landscape                   bool
	}

	SignatureData struct {
		SignedBy    string
		SignedAt    *time.Time
		State       coredata.DocumentVersionSignatureState
		RequestedAt time.Time
	}

	SignaturePageData struct {
		Signatures []SignatureData
		Landscape  bool
	}

	StatementOfApplicabilityData struct {
		Title            string
		OrganizationName string
		CreatedAt        time.Time
		TotalControls    int
		Rows             []SOARow
	}

	SOARow struct {
		FrameworkName        string
		ControlSection       string
		ControlName          string
		Applicability        string
		Justification        string
		MaturityLevel        string
		NotImplJustification string
		Regulatory           string
		Contractual          string
		BestPractice         string
		RiskAssessment       string
	}

	DataListData struct {
		Title            string
		OrganizationName string
		CreatedAt        time.Time
		TotalData        int
		Rows             []DataListRow
	}

	DataListRow struct {
		Name           string
		Classification string
		Owner          string
		Vendors        string
	}

	AssetListData struct {
		Title            string
		OrganizationName string
		CreatedAt        time.Time
		TotalAssets      int
		Rows             []AssetListRow
	}

	AssetListRow struct {
		Name            string
		AssetType       string
		Amount          int
		DataTypesStored string
		Owner           string
		Vendors         string
	}

	FindingListData struct {
		Title            string
		OrganizationName string
		CreatedAt        time.Time
		TotalFindings    int
		Rows             []FindingListRow
	}

	FindingListRow struct {
		ReferenceID        string
		Kind               string
		Description        string
		Source             string
		IdentifiedOn       string
		RootCause          string
		CorrectiveAction   string
		EffectivenessCheck string
		Status             string
		Priority           string
		Owner              string
		DueDate            string
	}

	ObligationListData struct {
		Title            string
		OrganizationName string
		CreatedAt        time.Time
		TotalObligations int
		Rows             []ObligationListRow
	}

	ObligationListRow struct {
		Area                   string
		Source                 string
		Requirement            string
		ActionsToBeImplemented string
		Status                 string
		Type                   string
		Regulator              string
		Owner                  string
		DueDate                string
	}

	ProcessingActivityListData struct {
		Title                     string
		OrganizationName          string
		CreatedAt                 time.Time
		TotalProcessingActivities int
		Rows                      []ProcessingActivityListRow
	}

	ProcessingActivityListRow struct {
		Name                                 string
		Purpose                              string
		Role                                 string
		DataSubjectCategory                  string
		PersonalDataCategory                 string
		SpecialOrCriminalData                string
		LawfulBasis                          string
		ConsentEvidenceLink                  string
		Recipients                           string
		Location                             string
		InternationalTransfers               string
		TransferSafeguards                   string
		RetentionPeriod                      string
		SecurityMeasures                     string
		DataProtectionImpactAssessmentNeeded string
		TransferImpactAssessmentNeeded       string
		LastReviewDate                       string
		NextReviewDate                       string
		DataProtectionOfficer                string
		Vendors                              string
	}

	DataProtectionImpactAssessmentListData struct {
		Title                                string
		OrganizationName                     string
		CreatedAt                            time.Time
		TotalDataProtectionImpactAssessments int
		Rows                                 []DataProtectionImpactAssessmentListRow
	}

	DataProtectionImpactAssessmentListRow struct {
		ProcessingActivityName      string
		Description                 string
		NecessityAndProportionality string
		PotentialRisk               string
		Mitigations                 string
		ResidualRisk                string
	}

	TransferImpactAssessmentListData struct {
		Title                          string
		OrganizationName               string
		CreatedAt                      time.Time
		TotalTransferImpactAssessments int
		Rows                           []TransferImpactAssessmentListRow
	}

	TransferImpactAssessmentListRow struct {
		ProcessingActivityName string
		DataSubjects           string
		Transfer               string
		LegalMechanism         string
		LocalLawRisk           string
		SupplementaryMeasures  string
	}
)

func BoolLabel(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

func MaturityLabel(l coredata.ControlMaturityLevel) string {
	switch l {
	case coredata.ControlMaturityLevelNone:
		return "0 - None"
	case coredata.ControlMaturityLevelInitial:
		return "1 - Initial"
	case coredata.ControlMaturityLevelManaged:
		return "2 - Managed"
	case coredata.ControlMaturityLevelDefined:
		return "3 - Defined"
	case coredata.ControlMaturityLevelQuantitativelyManaged:
		return "4 - Quantitatively Managed"
	case coredata.ControlMaturityLevelOptimizing:
		return "5 - Optimizing"
	}
	return "Not set"
}

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
)

// ProseMirrorJSONToHTML converts ProseMirror/Tiptap document JSON to an HTML fragment.
// On parse or render failure it returns a single escaped paragraph with the raw input.
func ProseMirrorJSONToHTML(content json.RawMessage) template.HTML {
	s := strings.TrimSpace(string(content))
	if s == "" {
		return template.HTML("")
	}
	node, err := prosemirror.Parse(s)
	if err != nil {
		return template.HTML(fmt.Sprintf("<p>%s</p>", html.EscapeString(s)))
	}
	htmlStr, err := prosemirror.RenderHTML(node)
	if err != nil {
		return template.HTML(fmt.Sprintf("<p>%s</p>", html.EscapeString(s)))
	}
	return template.HTML(htmlStr)
}

func RenderHTML(data DocumentData) ([]byte, error) {
	data.MermaidJS = template.JS(mermaidJSSource)

	page := struct {
		DocumentData
		BodyHTML template.HTML
	}{
		DocumentData: data,
		BodyHTML:     ProseMirrorJSONToHTML(data.Content),
	}

	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, page); err != nil {
		return nil, fmt.Errorf("cannot execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderSignaturePageHTML(data SignaturePageData) ([]byte, error) {
	var buf bytes.Buffer
	if err := signaturePageTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute signature page template: %w", err)
	}

	return buf.Bytes(), nil
}
