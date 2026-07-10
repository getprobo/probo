// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package complianceportal

const (
	// Custom domain actions.
	ActionCustomDomainGet    = "compliance-portal:custom-domain:get"
	ActionCustomDomainCreate = "compliance-portal:custom-domain:create"
	ActionCustomDomainDelete = "compliance-portal:custom-domain:delete"

	// Compliance portal actions.
	ActionCompliancePortalGet                          = "compliance-portal:portal:get"
	ActionCompliancePortalUpdate                       = "compliance-portal:portal:update"
	ActionCompliancePortalGetNda                       = "compliance-portal:portal:get-nda"
	ActionCompliancePortalNonDisclosureAgreementUpload = "compliance-portal:portal:upload-nda"
	ActionCompliancePortalNonDisclosureAgreementDelete = "compliance-portal:portal:delete-nda"

	// Compliance portal access actions.
	ActionCompliancePortalAccessGet    = "compliance-portal:portal-access:get"
	ActionCompliancePortalAccessList   = "compliance-portal:portal-access:list"
	ActionCompliancePortalAccessCreate = "compliance-portal:portal-access:create"
	ActionCompliancePortalAccessUpdate = "compliance-portal:portal-access:update"
	ActionCompliancePortalAccessDelete = "compliance-portal:portal-access:delete"

	// Compliance portal reference actions.
	ActionCompliancePortalReferenceList       = "compliance-portal:portal-reference:list"
	ActionCompliancePortalReferenceGetLogoUrl = "compliance-portal:portal-reference:get-logo-url"
	ActionCompliancePortalReferenceCreate     = "compliance-portal:portal-reference:create"
	ActionCompliancePortalReferenceUpdate     = "compliance-portal:portal-reference:update"
	ActionCompliancePortalReferenceDelete     = "compliance-portal:portal-reference:delete"

	// Compliance portal file actions.
	ActionCompliancePortalFileGet        = "compliance-portal:portal-file:get"
	ActionCompliancePortalFileList       = "compliance-portal:portal-file:list"
	ActionCompliancePortalFileGetFileUrl = "compliance-portal:portal-file:get-file-url"
	ActionCompliancePortalFileUpdate     = "compliance-portal:portal-file:update"
	ActionCompliancePortalFileDelete     = "compliance-portal:portal-file:delete"
	ActionCompliancePortalFileCreate     = "compliance-portal:portal-file:create"

	// Compliance portal document access actions.
	ActionCompliancePortalDocumentAccessList = "compliance-portal:portal-document-access:list"

	// MailingListUpdate actions.
	ActionMailingListUpdateList   = "compliance-portal:mailing-list-update:list"
	ActionMailingListUpdateCreate = "compliance-portal:mailing-list-update:create"
	ActionMailingListUpdateUpdate = "compliance-portal:mailing-list-update:update"
	ActionMailingListUpdateSend   = "compliance-portal:mailing-list-update:send"
	ActionMailingListUpdateDelete = "compliance-portal:mailing-list-update:delete"

	// MailingList actions.
	ActionMailingListUpdate = "compliance-portal:mailing-list:update"

	// MailingListSubscriber actions.
	ActionMailingListSubscriberList   = "compliance-portal:mailing-list-subscriber:list"
	ActionMailingListSubscriberCreate = "compliance-portal:mailing-list-subscriber:create"
	ActionMailingListSubscriberDelete = "compliance-portal:mailing-list-subscriber:delete"

	// ComplianceFramework actions.
	ActionComplianceFrameworkList       = "compliance-portal:compliance-framework:list"
	ActionComplianceFrameworkCreate     = "compliance-portal:compliance-framework:create"
	ActionComplianceFrameworkDelete     = "compliance-portal:compliance-framework:delete"
	ActionComplianceFrameworkUpdateRank = "compliance-portal:compliance-framework:update-rank"

	// ComplianceCustomLink actions.
	ActionComplianceCustomLinkList   = "compliance-portal:compliance-custom-link:list"
	ActionComplianceCustomLinkCreate = "compliance-portal:compliance-custom-link:create"
	ActionComplianceCustomLinkUpdate = "compliance-portal:compliance-custom-link:update"
	ActionComplianceCustomLinkDelete = "compliance-portal:compliance-custom-link:delete"
)
