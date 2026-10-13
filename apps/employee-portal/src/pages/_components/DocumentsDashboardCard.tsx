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

import { ArrowRightIcon, CheckIcon, FilesIcon, TrayIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentsDashboardCard_viewer$key } from "#/__generated__/core/DocumentsDashboardCard_viewer.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

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
    <Card variant="soft" padding="none" size={3} className={slots.frame()}>
      <div className={slots.wash()} />
      <div className={slots.header()}>
        <FilesIcon className={slots.icon()} />
        <div className={slots.copy()}>
          <Heading level={2} size={2} weight="medium" highContrast>
            {t("homePage.dashboard.documents.title")}
          </Heading>
          <Text size={1} color="current" className={slots.description()}>
            {t("homePage.dashboard.documents.description")}
          </Text>
        </div>
        <Link
          to={`/${employeePortalId}/documents`}
          size={2}
          color="neutral"
          underline={false}
          className={slots.view()}
        >
          {t("homePage.dashboard.view")}
        </Link>
      </div>
      <div className={slots.body()}>
        <DocumentsDashboardCardBody
          bodyState={bodyState}
          pendingCount={pendingCount}
          to={`/${employeePortalId}/documents`}
        />
      </div>
    </Card>
  );
}

function DocumentsDashboardCardBody({
  bodyState,
  pendingCount,
  to,
}: {
  bodyState: "empty" | "pending" | "allDone";
  pendingCount: number;
  to: string;
}): ReactNode {
  const { t } = useTranslation();
  const slots = dashboardCard();

  if (bodyState === "empty") {
    return (
      <div className={slots.empty()}>
        <TrayIcon className={slots.emptyIcon()} />
        <Text size={2} weight="medium" color="current" className={slots.emptyLabel()}>
          {t("homePage.dashboard.documents.empty")}
        </Text>
      </div>
    );
  }

  if (bodyState === "allDone") {
    return (
      <div className={slots.empty()}>
        <CheckIcon className={slots.emptyIcon()} />
        <Text size={2} weight="medium" color="current" className={slots.emptyLabel()}>
          {t("homePage.dashboard.allDone")}
        </Text>
      </div>
    );
  }

  return (
    <>
      <Heading level={3} size={5} weight="medium" highContrast>
        {t("homePage.dashboard.documents.pendingCount", { count: pendingCount })}
      </Heading>
      <ButtonLink
        to={to}
        size={2}
        variant="soft"
        color="neutral"
        iconEnd={<ArrowRightIcon />}
      >
        {t("homePage.dashboard.documents.action")}
      </ButtonLink>
    </>
  );
}
