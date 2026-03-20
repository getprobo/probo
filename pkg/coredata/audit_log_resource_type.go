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
	case VendorRiskAssessmentEntityType:
		return "VendorRiskAssessment"
	case VendorEntityType:
		return "Vendor"
	case VendorComplianceReportEntityType:
		return "VendorComplianceReport"
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
	case ReportEntityType:
		return "Report"
	case TrustCenterEntityType:
		return "TrustCenter"
	case TrustCenterAccessEntityType:
		return "TrustCenterAccess"
	case VendorBusinessAssociateAgreementEntityType:
		return "VendorBusinessAssociateAgreement"
	case FileEntityType:
		return "File"
	case VendorContactEntityType:
		return "VendorContact"
	case VendorDataPrivacyAgreementEntityType:
		return "VendorDataPrivacyAgreement"
	case FindingEntityType:
		return "Finding"
	case ObligationEntityType:
		return "Obligation"
	case VendorServiceEntityType:
		return "VendorService"
	case SnapshotEntityType:
		return "Snapshot"
	case ProcessingActivityEntityType:
		return "ProcessingActivity"
	case TrustCenterReferenceEntityType:
		return "TrustCenterReference"
	case TrustCenterDocumentAccessEntityType:
		return "TrustCenterDocumentAccess"
	case CustomDomainEntityType:
		return "CustomDomain"
	case InvitationEntityType:
		return "Invitation"
	case MembershipEntityType:
		return "Membership"
	case TrustCenterFileEntityType:
		return "TrustCenterFile"
	case MeetingEntityType:
		return "Meeting"
	case DataProtectionImpactAssessmentEntityType:
		return "DataProtectionImpactAssessment"
	case TransferImpactAssessmentEntityType:
		return "TransferImpactAssessment"
	case RightsRequestEntityType:
		return "RightsRequest"
	case StateOfApplicabilityEntityType:
		return "StateOfApplicability"
	case ApplicabilityStatementEntityType:
		return "ApplicabilityStatement"
	case WebhookSubscriptionEntityType:
		return "WebhookSubscription"
	case ComplianceFrameworkEntityType:
		return "ComplianceFramework"
	case ComplianceExternalURLEntityType:
		return "ComplianceExternalURL"
	case MailingListEntityType:
		return "MailingList"
	case MailingListSubscriberEntityType:
		return "MailingListSubscriber"
	case MailingListUpdateEntityType:
		return "MailingListUpdate"
	case AuditLogEntryEntityType:
		return "AuditLogEntry"
	default:
		return "Unknown"
	}
}
