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

import {
  ArrowRightIcon,
  CheckIcon,
  SignatureIcon,
  StampIcon,
  TrayIcon,
} from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { dashboardCard } from "./variants";

export interface DashboardCardProps {
  kind: "signatures" | "approvals";
  organizationId: string;
  pendingCount: number;
  completedCount: number;
  firstPendingId: string | null;
}

export function DashboardCard({
  kind,
  organizationId,
  pendingCount,
  completedCount,
  firstPendingId,
}: DashboardCardProps) {
  const { t } = useTranslation();
  const slots = dashboardCard();

  const listPath = kind === "signatures"
    ? `/${organizationId}/signatures`
    : `/${organizationId}/approvals`;

  let bodyState: "empty" | "pending" | "allDone" = "empty";
  if (pendingCount > 0) {
    bodyState = "pending";
  } else if (completedCount > 0) {
    bodyState = "allDone";
  }

  const Icon = kind === "signatures" ? SignatureIcon : StampIcon;

  return (
    <Card variant="soft" padding="none" size={3} className={slots.frame()}>
      <div className={slots.header()}>
        <Icon className={slots.icon()} />
        <div className={slots.copy()}>
          <Heading level={2} size={2} weight="medium" highContrast>
            {t(`homePage.dashboard.${kind}.title`)}
          </Heading>
          <Text size={1} color="current" className={slots.description()}>
            {t(`homePage.dashboard.${kind}.description`)}
          </Text>
        </div>
        {bodyState !== "empty" && (
          <Link
            to={listPath}
            size={2}
            color="neutral"
            underline={false}
            className={slots.view()}
          >
            {t("homePage.dashboard.view")}
          </Link>
        )}
      </div>
      <div className={slots.body()}>
        <DashboardCardBody
          kind={kind}
          bodyState={bodyState}
          pendingCount={pendingCount}
          firstPendingId={firstPendingId}
          organizationId={organizationId}
        />
      </div>
    </Card>
  );
}

function DashboardCardBody({
  kind,
  bodyState,
  pendingCount,
  firstPendingId,
  organizationId,
}: {
  kind: "signatures" | "approvals";
  bodyState: "empty" | "pending" | "allDone";
  pendingCount: number;
  firstPendingId: string | null;
  organizationId: string;
}): ReactNode {
  const { t } = useTranslation();
  const slots = dashboardCard();

  if (bodyState === "empty") {
    return (
      <div className={slots.empty()}>
        <TrayIcon className={slots.emptyIcon()} />
        <Text size={2} weight="medium" color="current" className={slots.description()}>
          {t(`homePage.dashboard.${kind}.empty`)}
        </Text>
      </div>
    );
  }

  if (bodyState === "allDone") {
    return (
      <div className={slots.empty()}>
        <CheckIcon className={slots.emptyIcon()} />
        <Text size={2} weight="medium" color="current" className={slots.description()}>
          {t("homePage.dashboard.allDone")}
        </Text>
      </div>
    );
  }

  const actionPath = kind === "signatures"
    ? `/${organizationId}/signatures/${firstPendingId}`
    : `/${organizationId}/approvals/${firstPendingId}`;

  return (
    <>
      <Heading level={3} size={5} weight="medium" highContrast>
        {t(`homePage.dashboard.${kind}.pendingCount`, { count: pendingCount })}
      </Heading>
      {firstPendingId != null && (
        <ButtonLink
          to={actionPath}
          size={2}
          variant="soft"
          color="neutral"
          iconEnd={<ArrowRightIcon />}
        >
          {t(`homePage.dashboard.${kind}.action`)}
        </ButtonLink>
      )}
    </>
  );
}
