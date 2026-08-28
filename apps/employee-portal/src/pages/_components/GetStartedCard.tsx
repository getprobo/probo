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

import { HandWavingIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useParams } from "react-router";

import type { GetStartedCard_viewer$key } from "#/__generated__/core/GetStartedCard_viewer.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { GetStartedStep } from "./GetStartedStep";
import { getStartedCard } from "./variants";

const getStartedCardFragment = graphql`
  fragment GetStartedCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    pendingSignatures: signableDocuments(
      organizationId: $organizationId
      first: 1
      filter: { signed: false }
    ) {
      totalCount
      edges {
        node {
          id
        }
      }
    }
    pendingApprovals: approvableDocuments(
      organizationId: $organizationId
      first: 1
      filter: { approvalStates: [PENDING] }
    ) {
      totalCount
      edges {
        node {
          id
        }
      }
    }
  }
`;

export interface GetStartedCardProps {
  viewerKey: GetStartedCard_viewer$key;
}

export function GetStartedCard({ viewerKey }: GetStartedCardProps) {
  const { t } = useTranslation();
  const { organizationId } = useParams();
  const slots = getStartedCard();
  const viewer = useFragment(getStartedCardFragment, viewerKey);

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  const pendingSignatureCount = viewer.pendingSignatures.totalCount;
  const pendingApprovalCount = viewer.pendingApprovals.totalCount;
  const firstPendingSignatureId = viewer.pendingSignatures.edges[0]?.node.id ?? null;
  const firstPendingApprovalId = viewer.pendingApprovals.edges[0]?.node.id ?? null;

  const steps = [];

  if (pendingSignatureCount > 0 && firstPendingSignatureId != null) {
    steps.push({
      key: "signatures",
      title: t("homePage.getStarted.signatures.count", { count: pendingSignatureCount }),
      description: t("homePage.getStarted.signatures.description"),
      actionLabel: t("homePage.getStarted.signatures.action"),
      to: `/${organizationId}/signatures/${firstPendingSignatureId}`,
    });
  }

  if (pendingApprovalCount > 0 && firstPendingApprovalId != null) {
    steps.push({
      key: "approvals",
      title: t("homePage.getStarted.approvals.count", { count: pendingApprovalCount }),
      description: t("homePage.getStarted.approvals.description"),
      actionLabel: t("homePage.getStarted.approvals.action"),
      to: `/${organizationId}/approvals/${firstPendingApprovalId}`,
    });
  }

  return (
    <Card variant="soft" padding="none" size={3} className={slots.frame()}>
      <div className={slots.wash()} />
      <div className={slots.header()}>
        <HandWavingIcon className={slots.icon()} />
        <div className={slots.copy()}>
          <Heading level={2} size={6} weight="medium" highContrast align="center">
            {t("homePage.getStarted.title")}
          </Heading>
          <Text size={2} color="neutral" align="center" className={slots.description()}>
            {t("homePage.getStarted.description")}
          </Text>
        </div>
      </div>
      {steps.length > 0 && (
        <div className={slots.steps()}>
          {steps.map((step, index) => (
            <GetStartedStep
              key={step.key}
              index={index + 1}
              title={step.title}
              description={step.description}
              actionLabel={step.actionLabel}
              to={step.to}
              tone={index === 0 ? "current" : "upcoming"}
            />
          ))}
        </div>
      )}
    </Card>
  );
}
