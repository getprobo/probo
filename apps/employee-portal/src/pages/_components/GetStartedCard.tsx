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

import type { GetStartedCard_viewer$key } from "#/__generated__/core/GetStartedCard_viewer.graphql";
import type { DocumentQueueKind } from "#/pages/_lib/documentQueue";
import { useDocumentQueue } from "#/pages/_lib/DocumentQueueContext";

import { GetStartedStep } from "./GetStartedStep";
import { getStartedCard } from "./variants";

const getStartedCardFragment = graphql`
  fragment GetStartedCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    pendingSignatures: signableDocuments(
      organizationId: $organizationId
      filter: { signed: false }
    ) {
      totalCount
    }
    pendingApprovals: approvableDocuments(
      organizationId: $organizationId
      filter: { approvalStates: [PENDING] }
    ) {
      totalCount
    }
  }
`;

export interface GetStartedCardProps {
  viewerKey: GetStartedCard_viewer$key;
}

export function GetStartedCard({ viewerKey }: GetStartedCardProps) {
  const { t } = useTranslation();
  const { advancing, startQueue } = useDocumentQueue();
  const slots = getStartedCard();
  const viewer = useFragment(getStartedCardFragment, viewerKey);

  const pendingSignatureCount = viewer.pendingSignatures.totalCount;
  const pendingApprovalCount = viewer.pendingApprovals.totalCount;

  const steps: {
    key: DocumentQueueKind;
    title: string;
    description: string;
    actionLabel: string;
  }[] = [];

  if (pendingSignatureCount > 0) {
    steps.push({
      key: "signatures",
      title: t("homePage.getStarted.signatures.count", { count: pendingSignatureCount }),
      description: t("homePage.getStarted.signatures.description"),
      actionLabel: t("homePage.getStarted.signatures.action"),
    });
  }

  if (pendingApprovalCount > 0) {
    steps.push({
      key: "approvals",
      title: t("homePage.getStarted.approvals.count", { count: pendingApprovalCount }),
      description: t("homePage.getStarted.approvals.description"),
      actionLabel: t("homePage.getStarted.approvals.action"),
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
              actionBusy={advancing}
              tone={index === 0 ? "current" : "upcoming"}
              onAction={() => {
                startQueue(step.key);
              }}
            />
          ))}
        </div>
      )}
    </Card>
  );
}
