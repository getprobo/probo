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

import { graphql } from "react-relay";

// Read through the `permission` key on each nav table entry rather than by
// name here, which the rule cannot follow. See visibleNavGroups.
/* eslint-disable relay/unused-fields */
export const navPermissionsFragment = graphql`
  fragment navPermissions_organization on Organization {
    canGetContext: permission(action: "core:organization-context:get")
    canListTasks: permission(action: "core:task:list")
    canListMeasures: permission(action: "core:measure:list")
    canListRisks: permission(action: "core:risk:list")
    canListRiskAnalyses: permission(action: "core:risk-analysis:list")
    canListFrameworks: permission(action: "core:framework:list")
    canListMembers: permission(action: "iam:membership:list")
    canListThirdParties: permission(action: "core:thirdParty:list")
    canListDocuments: permission(action: "core:document:list")
    canListAssets: permission(action: "core:asset:list")
    canListDevices: permission(action: "itam:device:list")
    canListData: permission(action: "core:datum:list")
    canListAudits: permission(action: "core:audit:list")
    canListFindings: permission(action: "core:finding:list")
    canListBusinessFunctions: permission(action: "core:business-function:list")
    canListAiSystems: permission(action: "core:ai-system:list")
    canListObligations: permission(action: "core:obligation:list")
    canListProcessingActivities: permission(action: "core:processing-activity:list")
    canListDataProtectionImpactAssessments: permission(action: "core:data-protection-impact-assessment:list")
    canListTransferImpactAssessments: permission(action: "core:transfer-impact-assessment:list")
    canListStatementsOfApplicability: permission(action: "core:statement-of-applicability:list")
    canListRightsRequests: permission(action: "core:rights-request:list")
    canListAccessReviewCampaigns: permission(action: "access-review:campaign:list")
    canListAccessReviewSources: permission(action: "access-review:source:list")
    canGetCompliancePortal: permission(action: "compliance-portal:portal:get")
    canListCookieBanners: permission(action: "core:cookie-banner:list")
    canListAuditLogEntries: permission(action: "iam:audit-log-entry:list")
    canListWebhookSubscriptions: permission(action: "core:webhook-subscription:list")
    canConnectSlack: permission(action: "core:connector:initiate")
    canUninstallSlack: permission(action: "core:connector:delete")
    canUpdateOrganization: permission(action: "iam:organization:update")
  }
`;
