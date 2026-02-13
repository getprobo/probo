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
	_                                          uint16 = 8 // PeopleEntityType - removed
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
	ApplicabilityStatementEntityType           uint16 = 50
	MembershipProfileEntityType                uint16 = 51
	SCIMConfigurationEntityType                uint16 = 52
	SCIMEventEntityType                        uint16 = 53
	TokenEntityType                            uint16 = 54
	SCIMBridgeEntityType                       uint16 = 55
	WebhookSubscriptionEntityType              uint16 = 56
	WebhookDataEntityType                      uint16 = 57
	WebhookEventEntityType                     uint16 = 58
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
	case RightsRequestEntityType:
		return &RightsRequest{ID: id}, true
	case StateOfApplicabilityEntityType:
		return &StateOfApplicability{ID: id}, true
	case ApplicabilityStatementEntityType:
		return &ApplicabilityStatement{ID: id}, true
	case MembershipProfileEntityType:
		return &MembershipProfile{ID: id}, true
	case SCIMConfigurationEntityType:
		return &SCIMConfiguration{ID: id}, true
	case SCIMEventEntityType:
		return &SCIMEvent{ID: id}, true
	case TokenEntityType:
		return &Token{ID: id}, true
	case SCIMBridgeEntityType:
		return &SCIMBridge{ID: id}, true
	case WebhookSubscriptionEntityType:
		return &WebhookSubscription{ID: id}, true
	case WebhookDataEntityType:
		return &WebhookData{ID: id}, true
	case WebhookEventEntityType:
		return &WebhookEvent{ID: id}, true
	default:
		return nil, false
	}
}
