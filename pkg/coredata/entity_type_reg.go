// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package coredata

import "go.probo.inc/probo/pkg/gid"

type ctxKey struct{ name string }

var (
	ContextKeyIPAddress = &ctxKey{name: "ip_address"}
)

const (
	OrganizationEntityType                           uint16 = 0
	FrameworkEntityType                              uint16 = 1
	MeasureEntityType                                uint16 = 2
	TaskEntityType                                   uint16 = 3
	EvidenceEntityType                               uint16 = 4
	ConnectorEntityType                              uint16 = 5
	ThirdPartyRiskAssessmentEntityType               uint16 = 6
	ThirdPartyEntityType                             uint16 = 7
	_                                                uint16 = 8 // PeopleEntityType - removed
	ThirdPartyComplianceReportEntityType             uint16 = 9
	DocumentEntityType                               uint16 = 10
	IdentityEntityType                               uint16 = 11
	SessionEntityType                                uint16 = 12
	EmailEntityType                                  uint16 = 13
	ControlEntityType                                uint16 = 14
	RiskEntityType                                   uint16 = 15
	DocumentVersionEntityType                        uint16 = 16
	DocumentVersionSignatureEntityType               uint16 = 17
	AssetEntityType                                  uint16 = 18
	DatumEntityType                                  uint16 = 19
	AuditEntityType                                  uint16 = 20
	_                                                uint16 = 21 // ReportEntityType - removed
	CompliancePortalEntityType                       uint16 = 22
	CompliancePortalAccessEntityType                 uint16 = 23
	ThirdPartyBusinessAssociateAgreementEntityType   uint16 = 24
	FileEntityType                                   uint16 = 25
	ThirdPartyContactEntityType                      uint16 = 26
	ThirdPartyDataPrivacyAgreementEntityType         uint16 = 27
	_                                                uint16 = 28 // NonconformityEntityType - removed
	ObligationEntityType                             uint16 = 29
	ThirdPartyServiceEntityType                      uint16 = 30
	_                                                uint16 = 31 // SnapshotEntityType - removed
	_                                                uint16 = 32 // ContinualImprovementEntityType - removed
	ProcessingActivityEntityType                     uint16 = 33
	ExportJobEntityType                              uint16 = 34
	CompliancePortalReferenceEntityType              uint16 = 35
	CompliancePortalDocumentAccessEntityType         uint16 = 36
	CustomDomainEntityType                           uint16 = 37
	InvitationEntityType                             uint16 = 38
	MembershipEntityType                             uint16 = 39
	SlackMessageEntityType                           uint16 = 40
	CompliancePortalFileEntityType                   uint16 = 41
	SAMLConfigurationEntityType                      uint16 = 42
	PersonalAPIKeyEntityType                         uint16 = 43
	_                                                uint16 = 44 // PersonalAPIKeyMembershipEntityType - removed
	_                                                uint16 = 45 // MeetingEntityType - removed
	DataProtectionImpactAssessmentEntityType         uint16 = 46
	TransferImpactAssessmentEntityType               uint16 = 47
	RightsRequestEntityType                          uint16 = 48
	StatementOfApplicabilityEntityType               uint16 = 49
	ApplicabilityStatementEntityType                 uint16 = 50
	MembershipProfileEntityType                      uint16 = 51
	SCIMConfigurationEntityType                      uint16 = 52
	SCIMEventEntityType                              uint16 = 53
	TokenEntityType                                  uint16 = 54
	SCIMBridgeEntityType                             uint16 = 55
	WebhookSubscriptionEntityType                    uint16 = 56
	WebhookDataEntityType                            uint16 = 57
	WebhookEventEntityType                           uint16 = 58
	ElectronicSignatureEntityType                    uint16 = 59
	ElectronicSignatureEventEntityType               uint16 = 60
	EmailAttachmentEntityType                        uint16 = 61
	ComplianceFrameworkEntityType                    uint16 = 62
	ComplianceCustomLinkEntityType                   uint16 = 63
	MailingListEntityType                            uint16 = 64
	MailingListSubscriberEntityType                  uint16 = 65
	MailingListUpdateEntityType                      uint16 = 66
	FindingEntityType                                uint16 = 67
	AuditLogEntryEntityType                          uint16 = 68
	DocumentVersionApprovalQuorumEntityType          uint16 = 69
	DocumentVersionApprovalDecisionEntityType        uint16 = 70
	AccessReviewSourceEntityType                     uint16 = 71
	AccessReviewCampaignEntityType                   uint16 = 72
	AccessReviewEntryEntityType                      uint16 = 73
	AccessReviewEntryDecisionHistoryEntityType       uint16 = 74
	CookieBannerEntityType                           uint16 = 75
	CookieCategoryEntityType                         uint16 = 76
	CookieConsentRecordEntityType                    uint16 = 77
	CookieBannerVersionEntityType                    uint16 = 78
	OAuth2ClientEntityType                           uint16 = 79
	OAuth2ConsentEntityType                          uint16 = 80
	OAuth2AccessTokenEntityType                      uint16 = 81
	OAuth2RefreshTokenEntityType                     uint16 = 82
	OAuth2AuthorizationCodeEntityType                uint16 = 83
	OAuth2DeviceCodeEntityType                       uint16 = 84
	_                                                uint16 = 85 // CookieEntityType - removed
	CookieBannerTranslationEntityType                uint16 = 86
	AgentExecutionEntityType                         uint16 = 87
	_                                                uint16 = 88 // CookiePatternEntityType - removed
	TrackerPatternEntityType                         uint16 = 89
	DetectedTrackerEntityType                        uint16 = 90
	TrackerResourceEntityType                        uint16 = 91
	CommonThirdPartyEntityType                       uint16 = 92
	CommonThirdPartyDomainEntityType                 uint16 = 93
	CommonTrackerPatternEntityType                   uint16 = 94
	RiskAnalysisEntityType                           uint16 = 95
	RiskAnalysisNodeEntityType                       uint16 = 96
	RiskAnalysisProcessEntityType                    uint16 = 97
	RiskAnalysisThreatEntityType                     uint16 = 98
	RiskAnalysisDiagramEntityType                    uint16 = 99
	RiskAnalysisScenarioEntityType                   uint16 = 100
	RiskAnalysisBoundaryEntityType                   uint16 = 101
	AccessReviewCampaignSourceEntityType             uint16 = 102
	AccessReviewCampaignSourceFetchAttemptEntityType uint16 = 103
	CompliancePortalCommitmentGroupEntityType        uint16 = 104
	CompliancePortalCommitmentEntityType             uint16 = 105
	CertificateEntityType                            uint16 = 106
	DeviceEntityType                                 uint16 = 107
	DevicePostureEntityType                          uint16 = 108
	DeviceEnrollmentTokenEntityType                  uint16 = 109
	DevicePostureReportEntityType                    uint16 = 110
	CompliancePortalAuditEntityType                  uint16 = 111
	CompliancePortalThirdPartyEntityType             uint16 = 112
	CompliancePortalDocumentEntityType               uint16 = 113
	BusinessFunctionEntityType                       uint16 = 114
	AiSystemEntityType                               uint16 = 115
	ProbotIdentityBindingEntityType                  uint16 = 116
	SlackbotInstallationEntityType                   uint16 = 117
	BotDeliveryDestinationEntityType                 uint16 = 118
	SlackbotMessageEntityType                        uint16 = 119
	SlackbotEventEntityType                          uint16 = 120
	_                                                uint16 = 121 // BotConversationEntityType - removed
	SlackbotInteractiveCommandEntityType             uint16 = 122
	OperationReceiptEntityType                       uint16 = 123
	SlackDeliveryOperationEntityType                 uint16 = 124
	AgentInputEntityType                             uint16 = 125
	AgentExecutionAnchorEntityType                   uint16 = 126
	BotMessageEntityType                             uint16 = 127
	BotThreadSubjectEntityType                       uint16 = 128
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
	case ThirdPartyRiskAssessmentEntityType:
		return &ThirdPartyRiskAssessment{ID: id}, true
	case ThirdPartyEntityType:
		return &ThirdParty{ID: id}, true
	case ThirdPartyComplianceReportEntityType:
		return &ThirdPartyComplianceReport{ID: id}, true
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
	case CompliancePortalEntityType:
		return &CompliancePortal{ID: id}, true
	case CompliancePortalAccessEntityType:
		return &CompliancePortalAccess{ID: id}, true
	case ThirdPartyBusinessAssociateAgreementEntityType:
		return &ThirdPartyBusinessAssociateAgreement{ID: id}, true
	case FileEntityType:
		return &File{ID: id}, true
	case ThirdPartyContactEntityType:
		return &ThirdPartyContact{ID: id}, true
	case ThirdPartyDataPrivacyAgreementEntityType:
		return &ThirdPartyDataPrivacyAgreement{ID: id}, true
	case FindingEntityType:
		return &Finding{ID: id}, true
	case ObligationEntityType:
		return &Obligation{ID: id}, true
	case ThirdPartyServiceEntityType:
		return &ThirdPartyService{ID: id}, true
	case ProcessingActivityEntityType:
		return &ProcessingActivity{ID: id}, true
	case ExportJobEntityType:
		return &ExportJob{ID: id}, true
	case CompliancePortalReferenceEntityType:
		return &CompliancePortalReference{ID: id}, true
	case CompliancePortalDocumentAccessEntityType:
		return &CompliancePortalDocumentAccess{ID: id}, true
	case CustomDomainEntityType:
		return &CustomDomain{ID: id}, true
	case InvitationEntityType:
		return &Invitation{ID: id}, true
	case MembershipEntityType:
		return &Membership{ID: id}, true
	case SlackMessageEntityType:
		return &SlackMessage{ID: id}, true
	case CompliancePortalFileEntityType:
		return &CompliancePortalFile{ID: id}, true
	case SAMLConfigurationEntityType:
		return &SAMLConfiguration{ID: id}, true
	case PersonalAPIKeyEntityType:
		return &PersonalAPIKey{ID: id}, true
	case DataProtectionImpactAssessmentEntityType:
		return &DataProtectionImpactAssessment{ID: id}, true
	case TransferImpactAssessmentEntityType:
		return &TransferImpactAssessment{ID: id}, true
	case RightsRequestEntityType:
		return &RightsRequest{ID: id}, true
	case StatementOfApplicabilityEntityType:
		return &StatementOfApplicability{ID: id}, true
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
	case ElectronicSignatureEntityType:
		return &ElectronicSignature{ID: id}, true
	case ElectronicSignatureEventEntityType:
		return &ElectronicSignatureEvent{ID: id}, true
	case EmailAttachmentEntityType:
		return &EmailAttachment{ID: id}, true
	case ComplianceFrameworkEntityType:
		return &ComplianceFramework{ID: id}, true
	case ComplianceCustomLinkEntityType:
		return &ComplianceCustomLink{ID: id}, true
	case MailingListEntityType:
		return &MailingList{ID: id}, true
	case MailingListSubscriberEntityType:
		return &MailingListSubscriber{ID: id}, true
	case MailingListUpdateEntityType:
		return &MailingListUpdate{ID: id}, true
	case AuditLogEntryEntityType:
		return &AuditLogEntry{ID: id}, true
	case DocumentVersionApprovalDecisionEntityType:
		return &DocumentVersionApprovalDecision{ID: id}, true
	case DocumentVersionApprovalQuorumEntityType:
		return &DocumentVersionApprovalQuorum{ID: id}, true
	case AccessReviewSourceEntityType:
		return &AccessReviewSource{ID: id}, true
	case AccessReviewCampaignEntityType:
		return &AccessReviewCampaign{ID: id}, true
	case AccessReviewEntryEntityType:
		return &AccessReviewEntry{ID: id}, true
	case AccessReviewEntryDecisionHistoryEntityType:
		return &AccessReviewEntryDecisionHistory{ID: id}, true
	case CookieBannerEntityType:
		return &CookieBanner{ID: id}, true
	case CookieCategoryEntityType:
		return &CookieCategory{ID: id}, true
	case CookieConsentRecordEntityType:
		return &CookieConsentRecord{ID: id}, true
	case CookieBannerVersionEntityType:
		return &CookieBannerVersion{ID: id}, true
	case OAuth2ClientEntityType:
		return &OAuth2Client{ID: id}, true
	case OAuth2ConsentEntityType:
		return &OAuth2Consent{ID: id}, true
	case OAuth2AccessTokenEntityType:
		return &OAuth2AccessToken{ID: id}, true
	case OAuth2RefreshTokenEntityType:
		return &OAuth2RefreshToken{ID: id}, true
	case OAuth2AuthorizationCodeEntityType:
		return &OAuth2AuthorizationCode{ID: id}, true
	case OAuth2DeviceCodeEntityType:
		return &OAuth2DeviceCode{ID: id}, true
	case CookieBannerTranslationEntityType:
		return &CookieBannerTranslation{ID: id}, true
	case AgentExecutionEntityType:
		return &AgentExecution{ID: id}, true
	case TrackerPatternEntityType:
		return &TrackerPattern{ID: id}, true
	case DetectedTrackerEntityType:
		return &DetectedTracker{ID: id}, true
	case TrackerResourceEntityType:
		return &TrackerResource{ID: id}, true
	case CommonThirdPartyEntityType:
		return &CommonThirdParty{ID: id}, true
	case CommonThirdPartyDomainEntityType:
		return &CommonThirdPartyDomain{ID: id}, true
	case CommonTrackerPatternEntityType:
		return &CommonTrackerPattern{ID: id}, true
	case RiskAnalysisEntityType:
		return &RiskAnalysis{ID: id}, true
	case RiskAnalysisNodeEntityType:
		return &RiskAnalysisNode{ID: id}, true
	case RiskAnalysisProcessEntityType:
		return &RiskAnalysisProcess{ID: id}, true
	case RiskAnalysisThreatEntityType:
		return &RiskAnalysisThreat{ID: id}, true
	case RiskAnalysisDiagramEntityType:
		return &RiskAnalysisDiagram{ID: id}, true
	case RiskAnalysisScenarioEntityType:
		return &RiskAnalysisScenario{ID: id}, true
	case RiskAnalysisBoundaryEntityType:
		return &RiskAnalysisBoundary{ID: id}, true
	case AccessReviewCampaignSourceEntityType:
		return &AccessReviewCampaignSource{ID: id}, true
	case AccessReviewCampaignSourceFetchAttemptEntityType:
		return &AccessReviewCampaignSourceFetchAttempt{ID: id}, true
	case CompliancePortalCommitmentGroupEntityType:
		return &CompliancePortalCommitmentGroup{ID: id}, true
	case CompliancePortalCommitmentEntityType:
		return &CompliancePortalCommitment{ID: id}, true
	case CompliancePortalDocumentEntityType:
		return &CompliancePortalDocument{ID: id}, true
	case CompliancePortalAuditEntityType:
		return &CompliancePortalAudit{ID: id}, true
	case CompliancePortalThirdPartyEntityType:
		return &CompliancePortalThirdParty{ID: id}, true
	case CertificateEntityType:
		return &Certificate{ID: id}, true
	case DeviceEntityType:
		return &Device{ID: id}, true
	case DevicePostureEntityType:
		return &DevicePosture{ID: id}, true
	case DeviceEnrollmentTokenEntityType:
		return &DeviceEnrollmentToken{ID: id}, true
	case DevicePostureReportEntityType:
		return &DevicePostureReport{ID: id}, true
	case BusinessFunctionEntityType:
		return &BusinessFunction{ID: id}, true
	case AiSystemEntityType:
		return &AiSystem{ID: id}, true
	case ProbotIdentityBindingEntityType:
		return &ProbotIdentityBinding{ID: id}, true
	case SlackbotInstallationEntityType:
		return &SlackbotInstallation{ID: id}, true
	case BotDeliveryDestinationEntityType:
		return &BotDeliveryDestination{ID: id}, true
	case SlackbotMessageEntityType:
		return &SlackbotMessage{ID: id}, true
	case SlackbotEventEntityType:
		return &SlackbotEvent{ID: id}, true
	case SlackbotInteractiveCommandEntityType:
		return &SlackbotInteractiveCommand{ID: id}, true
	case OperationReceiptEntityType:
		return &OperationReceipt{ID: id}, true
	case SlackDeliveryOperationEntityType:
		return &SlackDeliveryOperation{ID: id}, true
	case AgentInputEntityType:
		return &AgentInput{ID: id}, true
	case AgentExecutionAnchorEntityType:
		return &AgentExecutionAnchor{ID: id}, true
	case BotMessageEntityType:
		return &BotMessage{ID: id}, true
	case BotThreadSubjectEntityType:
		return &BotThreadSubject{ID: id}, true
	default:
		return nil, false
	}
}
