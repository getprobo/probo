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

import { FilesIcon } from "@phosphor-icons/react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentsDashboardCard_viewer$key } from "#/__generated__/core/DocumentsDashboardCard_viewer.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { DashboardCardFrame } from "./DashboardCardFrame";
import { dashboardCard } from "./variants";

const documentsDashboardCardFragment = graphql`
  fragment DocumentsDashboardCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    pendingSignatures: signableDocuments(
      organizationId: $organizationId
      filter: { signed: false }
    ) {
      totalCount
    }
    completedSignatures: signableDocuments(
      organizationId: $organizationId
      filter: { signed: true }
    ) {
      totalCount
    }
    pendingApprovals: approvableDocuments(
      organizationId: $organizationId
      filter: { approvalStates: [PENDING] }
    ) {
      totalCount
    }
    approvedDocuments: approvableDocuments(
      organizationId: $organizationId
      filter: { approvalStates: [APPROVED] }
    ) {
      totalCount
    }
  }
`;

export interface DocumentsDashboardCardProps {
  viewerKey: DocumentsDashboardCard_viewer$key;
  wash?: boolean;
}

export function DocumentsDashboardCard({
  viewerKey,
  wash = false,
}: DocumentsDashboardCardProps) {
  const { t } = useTranslation();
  const { employeePortalId } = useEmployeePortalRoute();
  const slots = dashboardCard({ wash });
  const viewer = useFragment(documentsDashboardCardFragment, viewerKey);

  const pendingCount
    = viewer.pendingSignatures.totalCount + viewer.pendingApprovals.totalCount;
  const completedCount
    = viewer.completedSignatures.totalCount + viewer.approvedDocuments.totalCount;

  let bodyState: "empty" | "pending" | "allDone" = "empty";
  if (pendingCount > 0) {
    bodyState = "pending";
  } else if (completedCount > 0) {
    bodyState = "allDone";
  }

  return (
    <DashboardCardFrame
      wash={wash}
      icon={<FilesIcon className={slots.icon()} />}
      title={t("homePage.dashboard.documents.title")}
      description={t("homePage.dashboard.documents.description")}
      to={`/${employeePortalId}/documents`}
      bodyState={bodyState}
      pendingLabel={t("homePage.dashboard.documents.pendingCount", { count: pendingCount })}
      emptyLabel={t("homePage.dashboard.documents.empty")}
      actionLabel={t("homePage.dashboard.documents.action")}
    />
  );
}
