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

package coredata

// ResourceTypeName returns a human-readable name for an entity type.
func ResourceTypeName(entityType uint16) string {
	switch entityType {
	case OrganizationEntityType:
		return "Organization"
	case FrameworkEntityType:
		return "Framework"
	case MeasureEntityType:
		return "Measure"
	case TaskEntityType:
		return "Task"
	case EvidenceEntityType:
		return "Evidence"
	case ConnectorEntityType:
		return "Connector"
	case ThirdPartyRiskAssessmentEntityType:
		return "ThirdPartyRiskAssessment"
	case ThirdPartyEntityType:
		return "ThirdParty"
	case ThirdPartyComplianceReportEntityType:
		return "ThirdPartyComplianceReport"
	case DocumentEntityType:
		return "Document"
	case IdentityEntityType:
		return "Identity"
	case ControlEntityType:
		return "Control"
	case RiskEntityType:
		return "Risk"
	case DocumentVersionEntityType:
		return "DocumentVersion"
	case DocumentVersionSignatureEntityType:
		return "DocumentVersionSignature"
	case AssetEntityType:
		return "Asset"
	case DatumEntityType:
		return "Datum"
	case AuditEntityType:
		return "Audit"
	case CompliancePortalEntityType:
		return "CompliancePortal"
	case CompliancePortalAccessEntityType:
		return "CompliancePortalAccess"
	case ThirdPartyBusinessAssociateAgreementEntityType:
		return "ThirdPartyBusinessAssociateAgreement"
	case FileEntityType:
		return "File"
	case ThirdPartyContactEntityType:
		return "ThirdPartyContact"
	case ThirdPartyDataPrivacyAgreementEntityType:
		return "ThirdPartyDataPrivacyAgreement"
	case FindingEntityType:
		return "Finding"
	case ObligationEntityType:
		return "Obligation"
	case ThirdPartyServiceEntityType:
		return "ThirdPartyService"
	case ProcessingActivityEntityType:
		return "ProcessingActivity"
	case CompliancePortalReferenceEntityType:
		return "CompliancePortalReference"
	case CompliancePortalDocumentAccessEntityType:
		return "CompliancePortalDocumentAccess"
	case CustomDomainEntityType:
		return "CustomDomain"
	case InvitationEntityType:
		return "Invitation"
	case MembershipEntityType:
		return "Membership"
	case CompliancePortalFileEntityType:
		return "CompliancePortalFile"
	case DataProtectionImpactAssessmentEntityType:
		return "DataProtectionImpactAssessment"
	case TransferImpactAssessmentEntityType:
		return "TransferImpactAssessment"
	case RightsRequestEntityType:
		return "RightsRequest"
	case StatementOfApplicabilityEntityType:
		return "StatementOfApplicability"
	case ApplicabilityStatementEntityType:
		return "ApplicabilityStatement"
	case WebhookSubscriptionEntityType:
		return "WebhookSubscription"
	case ComplianceFrameworkEntityType:
		return "ComplianceFramework"
	case ComplianceCustomLinkEntityType:
		return "ComplianceCustomLink"
	case MailingListEntityType:
		return "MailingList"
	case MailingListSubscriberEntityType:
		return "MailingListSubscriber"
	case MailingListUpdateEntityType:
		return "MailingListUpdate"
	case AuditLogEntryEntityType:
		return "AuditLogEntry"
	case BusinessFunctionEntityType:
		return "BusinessFunction"
	default:
		return "Unknown"
	}
}
