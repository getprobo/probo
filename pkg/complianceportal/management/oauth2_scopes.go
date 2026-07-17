// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package management

import (
	"go.probo.inc/probo/pkg/coredata"
)

// The compliance-page scope string values are part of the external OAuth2
// contract and are kept stable even though the feature is named "compliance
// portal" on the Go side.
const (
	ScopeV1CompliancePortalRead coredata.OAuth2Scope = "v1:compliance-page:read"
	ScopeV1CompliancePortal     coredata.OAuth2Scope = "v1:compliance-page"
)

// OAuth2ScopeMappings maps the compliance portal OAuth2 scopes to the actions
// they grant.
var OAuth2ScopeMappings = map[coredata.OAuth2Scope][]string{
	ScopeV1CompliancePortalRead: {
		ActionCompliancePortalGet,
		ActionCompliancePortalGetNda,
		ActionCompliancePortalAccessGet,
		ActionCompliancePortalAccessList,
		ActionCompliancePortalFileGet,
		ActionCompliancePortalFileList,
		ActionCompliancePortalFileGetFileUrl,
		ActionCompliancePortalReferenceList,
		ActionCompliancePortalReferenceGetLogoUrl,
		ActionCompliancePortalDocumentAccessList,
		ActionMailingListUpdateList,
		ActionMailingListSubscriberList,
		ActionComplianceFrameworkList,
		ActionComplianceCustomLinkList,
		ActionCustomDomainGet,
	},
	ScopeV1CompliancePortal: {
		ActionCompliancePortalGet,
		ActionCompliancePortalGetNda,
		ActionCompliancePortalAccessGet,
		ActionCompliancePortalAccessList,
		ActionCompliancePortalFileGet,
		ActionCompliancePortalFileList,
		ActionCompliancePortalFileGetFileUrl,
		ActionCompliancePortalReferenceList,
		ActionCompliancePortalReferenceGetLogoUrl,
		ActionCompliancePortalDocumentAccessList,
		ActionMailingListUpdateList,
		ActionMailingListSubscriberList,
		ActionComplianceFrameworkList,
		ActionComplianceCustomLinkList,
		ActionCustomDomainGet,
		ActionCompliancePortalUpdate,
		ActionCompliancePortalNonDisclosureAgreementUpload,
		ActionCompliancePortalNonDisclosureAgreementDelete,
		ActionCompliancePortalAccessCreate,
		ActionCompliancePortalAccessUpdate,
		ActionCompliancePortalAccessDelete,
		ActionCompliancePortalFileUpdate,
		ActionCompliancePortalFileDelete,
		ActionCompliancePortalFileCreate,
		ActionCompliancePortalReferenceCreate,
		ActionCompliancePortalReferenceUpdate,
		ActionCompliancePortalReferenceDelete,
		ActionMailingListUpdateCreate,
		ActionMailingListUpdateUpdate,
		ActionMailingListUpdateSend,
		ActionMailingListUpdateDelete,
		ActionMailingListUpdate,
		ActionMailingListSubscriberCreate,
		ActionMailingListSubscriberDelete,
		ActionComplianceFrameworkCreate,
		ActionComplianceFrameworkDelete,
		ActionComplianceFrameworkUpdateRank,
		ActionComplianceCustomLinkCreate,
		ActionComplianceCustomLinkUpdate,
		ActionComplianceCustomLinkDelete,
		ActionCustomDomainCreate,
		ActionCustomDomainDelete,
	},
}
