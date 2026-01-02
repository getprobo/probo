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

package coredata

import "go.probo.inc/probo/pkg/gid"

type ctxKey struct{ name string }

var (
	ContextKeyIPAddress = &ctxKey{name: "ip_address"}
)

const (
	OrganizationEntityType                     uint16 = 0
	FrameworkEntityType                        uint16 = 1
	MeasureEntityType                          uint16 = 2
	TaskEntityType                             uint16 = 3
	EvidenceEntityType                         uint16 = 4
	ConnectorEntityType                        uint16 = 5
	VendorRiskAssessmentEntityType             uint16 = 6
	VendorEntityType                           uint16 = 7
	PeopleEntityType                           uint16 = 8
	VendorComplianceReportEntityType           uint16 = 9
	DocumentEntityType                         uint16 = 10
	IdentityEntityType                         uint16 = 11
	SessionEntityType                          uint16 = 12
	EmailEntityType                            uint16 = 13
	ControlEntityType                          uint16 = 14
	RiskEntityType                             uint16 = 15
	DocumentVersionEntityType                  uint16 = 16
	DocumentVersionSignatureEntityType         uint16 = 17
	AssetEntityType                            uint16 = 18
	DatumEntityType                            uint16 = 19
	AuditEntityType                            uint16 = 20
	ReportEntityType                           uint16 = 21
	TrustCenterEntityType                      uint16 = 22
	TrustCenterAccessEntityType                uint16 = 23
	VendorBusinessAssociateAgreementEntityType uint16 = 24
	FileEntityType                             uint16 = 25
	VendorContactEntityType                    uint16 = 26
	VendorDataPrivacyAgreementEntityType       uint16 = 27
	NonconformityEntityType                    uint16 = 28
	ObligationEntityType                       uint16 = 29
	VendorServiceEntityType                    uint16 = 30
	SnapshotEntityType                         uint16 = 31
	ContinualImprovementEntityType             uint16 = 32
	ProcessingActivityEntityType               uint16 = 33
	ExportJobEntityType                        uint16 = 34
	TrustCenterReferenceEntityType             uint16 = 35
	TrustCenterDocumentAccessEntityType        uint16 = 36
	CustomDomainEntityType                     uint16 = 37
	InvitationEntityType                       uint16 = 38
	MembershipEntityType                       uint16 = 39
	SlackMessageEntityType                     uint16 = 40
	TrustCenterFileEntityType                  uint16 = 41
	SAMLConfigurationEntityType                uint16 = 42
	PersonalAPIKeyEntityType                   uint16 = 43
	_                                          uint16 = 44 // PersonalAPIKeyMembershipEntityType - removed
	MeetingEntityType                          uint16 = 45
	DataProtectionImpactAssessmentEntityType   uint16 = 46
	TransferImpactAssessmentEntityType         uint16 = 47
	RightsRequestEntityType                    uint16 = 48
	StateOfApplicabilityEntityType             uint16 = 49
	StateOfApplicabilityControlEntityType      uint16 = 50
	MembershipProfileEntityType                uint16 = 51
)

func NewEntityFromID(id gid.GID) (any, bool) {
	switch id.EntityType() {
	case OrganizationEntityType:
		return &Organization{ID: id}, true
	case FrameworkEntityType:
		return &Framework{ID: id}, true
	case MeasureEntityType:
		return &Measure{ID: id}, true
	case TaskEntityType:
		return &Task{ID: id}, true
	case EvidenceEntityType:
		return &Evidence{ID: id}, true
	case ConnectorEntityType:
		return &Connector{ID: id}, true
	case VendorRiskAssessmentEntityType:
		return &VendorRiskAssessment{ID: id}, true
	case VendorEntityType:
		return &Vendor{ID: id}, true
	case PeopleEntityType:
		return &People{ID: id}, true
	case VendorComplianceReportEntityType:
		return &VendorComplianceReport{ID: id}, true
	case DocumentEntityType:
		return &Document{ID: id}, true
	case IdentityEntityType:
		return &Identity{ID: id}, true
	case SessionEntityType:
		return &Session{ID: id}, true
	case EmailEntityType:
		return &Email{ID: id}, true
	case ControlEntityType:
		return &Control{ID: id}, true
	case RiskEntityType:
		return &Risk{ID: id}, true
	case DocumentVersionEntityType:
		return &DocumentVersion{ID: id}, true
	case DocumentVersionSignatureEntityType:
		return &DocumentVersionSignature{ID: id}, true
	case AssetEntityType:
		return &Asset{ID: id}, true
	case DatumEntityType:
		return &Datum{ID: id}, true
	case AuditEntityType:
		return &Audit{ID: id}, true
	case ReportEntityType:
		return &Report{ID: id}, true
	case TrustCenterEntityType:
		return &TrustCenter{ID: id}, true
	case TrustCenterAccessEntityType:
		return &TrustCenterAccess{ID: id}, true
	case VendorBusinessAssociateAgreementEntityType:
		return &VendorBusinessAssociateAgreement{ID: id}, true
	case FileEntityType:
		return &File{ID: id}, true
	case VendorContactEntityType:
		return &VendorContact{ID: id}, true
	case VendorDataPrivacyAgreementEntityType:
		return &VendorDataPrivacyAgreement{ID: id}, true
	case NonconformityEntityType:
		return &Nonconformity{ID: id}, true
	case ObligationEntityType:
		return &Obligation{ID: id}, true
	case VendorServiceEntityType:
		return &VendorService{ID: id}, true
	case SnapshotEntityType:
		return &Snapshot{ID: id}, true
	case ContinualImprovementEntityType:
		return &ContinualImprovement{ID: id}, true
	case ProcessingActivityEntityType:
		return &ProcessingActivity{ID: id}, true
	case ExportJobEntityType:
		return &ExportJob{ID: id}, true
	case TrustCenterReferenceEntityType:
		return &TrustCenterReference{ID: id}, true
	case TrustCenterDocumentAccessEntityType:
		return &TrustCenterDocumentAccess{ID: id}, true
	case CustomDomainEntityType:
		return &CustomDomain{ID: id}, true
	case InvitationEntityType:
		return &Invitation{ID: id}, true
	case MembershipEntityType:
		return &Membership{ID: id}, true
	case SlackMessageEntityType:
		return &SlackMessage{ID: id}, true
	case TrustCenterFileEntityType:
		return &TrustCenterFile{ID: id}, true
	case SAMLConfigurationEntityType:
		return &SAMLConfiguration{ID: id}, true
	case PersonalAPIKeyEntityType:
		return &PersonalAPIKey{ID: id}, true
	case MeetingEntityType:
		return &Meeting{ID: id}, true
	case DataProtectionImpactAssessmentEntityType:
		return &DataProtectionImpactAssessment{ID: id}, true
	case TransferImpactAssessmentEntityType:
		return &TransferImpactAssessment{ID: id}, true
	case MembershipProfileEntityType:
		return &MembershipProfile{ID: id}, true
	default:
		return nil, false
	}
}

type EntityInfo struct {
	Model string
	Table string
}

var entityRegistry = map[uint16]EntityInfo{
	OrganizationEntityType: {
		Model: "Organization",
		Table: "organizations",
	},
	FrameworkEntityType: {
		Model: "Framework",
		Table: "frameworks",
	},
	MeasureEntityType: {
		Model: "Measure",
		Table: "measures",
	},
	TaskEntityType: {
		Model: "Task",
		Table: "tasks",
	},
	EvidenceEntityType: {
		Model: "Evidence",
		Table: "evidences",
	},
	ConnectorEntityType: {
		Model: "Connector",
		Table: "connectors",
	},
	VendorRiskAssessmentEntityType: {
		Model: "VendorRiskAssessment",
		Table: "vendor_risk_assessments",
	},
	VendorEntityType: {
		Model: "Vendor",
		Table: "vendors",
	},
	PeopleEntityType: {
		Model: "People",
		Table: "peoples",
	},
	VendorComplianceReportEntityType: {
		Model: "VendorComplianceReport",
		Table: "vendor_compliance_reports",
	},
	DocumentEntityType: {
		Model: "Document",
		Table: "documents",
	},
	IdentityEntityType: {
		Model: "Identity",
		Table: "identities",
	},
	SessionEntityType: {
		Model: "Session",
		Table: "iam_sessions",
	},
	EmailEntityType: {
		Model: "Email",
		Table: "emails",
	},
	ControlEntityType: {
		Model: "Control",
		Table: "controls",
	},
	RiskEntityType: {
		Model: "Risk",
		Table: "risks",
	},
	DocumentVersionEntityType: {
		Model: "DocumentVersion",
		Table: "document_versions",
	},
	DocumentVersionSignatureEntityType: {
		Model: "DocumentVersionSignature",
		Table: "document_version_signatures",
	},
	AssetEntityType: {
		Model: "Asset",
		Table: "assets",
	},
	DatumEntityType: {
		Model: "Datum",
		Table: "data",
	},
	AuditEntityType: {
		Model: "Audit",
		Table: "audits",
	},
	ReportEntityType: {
		Model: "Report",
		Table: "reports",
	},
	TrustCenterEntityType: {
		Model: "TrustCenter",
		Table: "trust_centers",
	},
	TrustCenterAccessEntityType: {
		Model: "TrustCenterAccess",
		Table: "trust_center_accesses",
	},
	VendorBusinessAssociateAgreementEntityType: {
		Model: "VendorBusinessAssociateAgreement",
		Table: "vendor_business_associate_agreements",
	},
	FileEntityType: {
		Model: "File",
		Table: "files",
	},
	VendorContactEntityType: {
		Model: "VendorContact",
		Table: "vendor_contacts",
	},
	VendorDataPrivacyAgreementEntityType: {
		Model: "VendorDataPrivacyAgreement",
		Table: "vendor_data_privacy_agreements",
	},
	NonconformityEntityType: {
		Model: "Nonconformity",
		Table: "nonconformities",
	},
	ObligationEntityType: {
		Model: "Obligation",
		Table: "obligations",
	},
	VendorServiceEntityType: {
		Model: "VendorService",
		Table: "vendor_services",
	},
	SnapshotEntityType: {
		Model: "Snapshot",
		Table: "snapshots",
	},
	ContinualImprovementEntityType: {
		Model: "ContinualImprovement",
		Table: "continual_improvements",
	},
	ProcessingActivityEntityType: {
		Model: "ProcessingActivity",
		Table: "processing_activities",
	},
	ExportJobEntityType: {
		Model: "ExportJob",
		Table: "export_jobs",
	},
	TrustCenterReferenceEntityType: {
		Model: "TrustCenterReference",
		Table: "trust_center_references",
	},
	TrustCenterDocumentAccessEntityType: {
		Model: "TrustCenterDocumentAccess",
		Table: "trust_center_document_accesses",
	},
	CustomDomainEntityType: {
		Model: "CustomDomain",
		Table: "custom_domains",
	},
	InvitationEntityType: {
		Model: "Invitation",
		Table: "iam_invitations",
	},
	MembershipEntityType: {
		Model: "Membership",
		Table: "iam_memberships",
	},
	SlackMessageEntityType: {
		Model: "SlackMessage",
		Table: "slack_messages",
	},
	TrustCenterFileEntityType: {
		Model: "TrustCenterFile",
		Table: "trust_center_files",
	},
	SAMLConfigurationEntityType: {
		Model: "SAMLConfiguration",
		Table: "iam_saml_configurations",
	},
	PersonalAPIKeyEntityType: {
		Model: "PersonalAPIKey",
		Table: "iam_personal_api_keys",
	},
	MeetingEntityType: {
		Model: "Meeting",
		Table: "meetings",
	},
	DataProtectionImpactAssessmentEntityType: {
		Model: "DataProtectionImpactAssessment",
		Table: "processing_activity_data_protection_impact_assessments",
	},
	TransferImpactAssessmentEntityType: {
		Model: "TransferImpactAssessment",
		Table: "processing_activity_transfer_impact_assessments",
	},
	RightsRequestEntityType: {
		Model: "RightsRequest",
		Table: "rights_requests",
	},
	StateOfApplicabilityEntityType: {
		Model: "StateOfApplicability",
		Table: "states_of_applicability",
	},
	StateOfApplicabilityControlEntityType: {
		Model: "StateOfApplicabilityControl",
		Table: "states_of_applicability_controls",
	},
	MembershipProfileEntityType: {
		Model: "MembershipProfile",
		Table: "iam_membership_profiles",
	},
}

func EntityTable(entityType uint16) (string, bool) {
	info, ok := entityRegistry[entityType]
	if !ok {
		return "", false
	}
	return info.Table, true
}

func EntityModel(entityType uint16) (string, bool) {
	info, ok := entityRegistry[entityType]
	if !ok {
		return "", false
	}
	return info.Model, true
}
