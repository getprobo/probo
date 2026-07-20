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

package management

import (
	"go.probo.inc/probo/pkg/coredata"
)

const (
	ScopeV1CompliancePortalRead coredata.OAuth2Scope = "v1:compliance-page:read"
	ScopeV1CompliancePortal     coredata.OAuth2Scope = "v1:compliance-page"
)

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
		ActionCompliancePortalCommitmentGroupList,
		ActionCompliancePortalCommitmentList,
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
		ActionCompliancePortalCommitmentGroupList,
		ActionCompliancePortalCommitmentList,
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
		ActionCompliancePortalCommitmentGroupCreate,
		ActionCompliancePortalCommitmentGroupUpdate,
		ActionCompliancePortalCommitmentGroupUpdateRank,
		ActionCompliancePortalCommitmentGroupDelete,
		ActionCompliancePortalCommitmentCreate,
		ActionCompliancePortalCommitmentUpdate,
		ActionCompliancePortalCommitmentUpdateRank,
		ActionCompliancePortalCommitmentDelete,
	},
}
