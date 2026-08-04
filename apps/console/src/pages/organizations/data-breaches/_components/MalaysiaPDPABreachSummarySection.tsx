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

import { dateTimeFormat } from "@probo/i18n";
import { Badge, Card } from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { MalaysiaPDPABreachSummarySection_incident$key } from "#/__generated__/core/MalaysiaPDPABreachSummarySection_incident.graphql";

import { getBreachDecisionBadgeVariant } from "../_lib/breachDisplay";

const incidentFragment = graphql`
  fragment MalaysiaPDPABreachSummarySection_incident on MalaysiaPDPABreachIncident {
    significantHarm
    significantScale
    notificationRecommendation
    notificationReasons
    notificationDecision
    decisionRationale
    decisionEvidence
    assessedAt
    ruleVersion
    ruleSource
    commissionerNotificationDueAt
    commissionerNotificationOverdue
    commissionerNotifiedAt
    commissionerNotificationReference
    commissionerConfirmationReceivedAt
    commissionerConfirmationReference
    phasedInformationDueAt
    dataSubjectsNotificationDueAt
    dataSubjectsNotificationOverdue
    dataSubjectsNotifiedAt
    dataSubjectsNotificationEvidence
  }
`;

interface MalaysiaPDPABreachSummarySectionProps {
  incidentKey: MalaysiaPDPABreachSummarySection_incident$key;
}

export function MalaysiaPDPABreachSummarySection({
  incidentKey,
}: MalaysiaPDPABreachSummarySectionProps) {
  const { i18n, t } = useTranslation("organizations/data-breaches");
  const incident = useFragment(incidentFragment, incidentKey);

  return (
    <Card padded className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-1">
          <h2 className="text-base font-medium">{t("summary.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("summary.description")}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {incident.significantHarm && (
            <Badge variant="danger">{t("summary.significantHarm")}</Badge>
          )}
          {incident.significantScale && (
            <Badge variant="warning">{t("summary.significantScale")}</Badge>
          )}
          {!incident.significantHarm && !incident.significantScale && (
            <Badge variant="success">{t("summary.noTrigger")}</Badge>
          )}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <DecisionBlock
          label={t("summary.recommendation")}
          decision={incident.notificationRecommendation}
        />
        <DecisionBlock
          label={t("summary.recordedDecision")}
          decision={incident.notificationDecision}
        />
      </div>

      {incident.notificationReasons.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-medium">{t("summary.reasons")}</h3>
          <ul className="list-disc space-y-1 pl-5 text-sm text-txt-secondary">
            {incident.notificationReasons.map(reason => (
              <li key={reason}>{t(`reasons.${reason}`)}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="grid gap-3 md:grid-cols-3">
        <DeadlineBlock
          label={t("deadlines.commissioner")}
          dueAt={incident.commissionerNotificationDueAt}
          completedAt={incident.commissionerNotifiedAt}
          overdue={incident.commissionerNotificationOverdue}
        />
        <DeadlineBlock
          label={t("deadlines.phasedInformation")}
          dueAt={incident.phasedInformationDueAt}
          completedAt={incident.commissionerConfirmationReceivedAt}
          overdue={false}
        />
        <DeadlineBlock
          label={t("deadlines.dataSubjects")}
          dueAt={incident.dataSubjectsNotificationDueAt}
          completedAt={incident.dataSubjectsNotifiedAt}
          overdue={incident.dataSubjectsNotificationOverdue}
        />
      </div>

      {(incident.decisionRationale || incident.decisionEvidence) && (
        <div className="grid gap-4 border-t border-border-low pt-4 md:grid-cols-2">
          <TextBlock
            label={t("summary.decisionRationale")}
            value={incident.decisionRationale}
          />
          <TextBlock
            label={t("summary.decisionEvidence")}
            value={incident.decisionEvidence}
          />
        </div>
      )}

      {(incident.commissionerNotificationReference
        || incident.commissionerConfirmationReference
        || incident.dataSubjectsNotificationEvidence) && (
        <div className="grid gap-4 border-t border-border-low pt-4 md:grid-cols-3">
          <TextBlock
            label={t("summary.commissionerReference")}
            value={incident.commissionerNotificationReference}
          />
          <TextBlock
            label={t("summary.confirmationReference")}
            value={incident.commissionerConfirmationReference}
          />
          <TextBlock
            label={t("summary.dataSubjectsEvidence")}
            value={incident.dataSubjectsNotificationEvidence}
          />
        </div>
      )}

      <p className="text-xs text-txt-tertiary">
        {t("summary.assessmentMetadata", {
          date: dateTimeFormat(i18n.language, incident.assessedAt),
          source: incident.ruleSource,
          version: incident.ruleVersion,
        })}
      </p>
    </Card>
  );
}

interface DecisionBlockProps {
  decision: string;
  label: string;
}

function DecisionBlock({ decision, label }: DecisionBlockProps) {
  const { t } = useTranslation("organizations/data-breaches");

  return (
    <div className="space-y-2 rounded-xl bg-subtle p-4">
      <p className="text-xs font-medium text-txt-tertiary">{label}</p>
      <Badge variant={getBreachDecisionBadgeVariant(decision)}>
        {t(`decisions.${decision}`)}
      </Badge>
    </div>
  );
}

interface DeadlineBlockProps {
  completedAt: string | null | undefined;
  dueAt: string | null | undefined;
  label: string;
  overdue: boolean;
}

function DeadlineBlock({
  completedAt,
  dueAt,
  label,
  overdue,
}: DeadlineBlockProps) {
  const { i18n, t } = useTranslation("organizations/data-breaches");
  let content: ReactNode = t("common.notApplicable");
  let badge: ReactNode;

  if (completedAt) {
    content = dateTimeFormat(i18n.language, completedAt);
    badge = <Badge variant="success">{t("deadlines.completed")}</Badge>;
  } else if (dueAt) {
    content = dateTimeFormat(i18n.language, dueAt);
    badge = overdue
      ? <Badge variant="danger">{t("deadlines.overdue")}</Badge>
      : <Badge variant="info">{t("deadlines.pending")}</Badge>;
  }

  return (
    <div className="space-y-2 rounded-xl border border-border-low p-4">
      <p className="text-xs font-medium text-txt-tertiary">{label}</p>
      <p className="text-sm font-medium">{content}</p>
      {badge}
    </div>
  );
}

interface TextBlockProps {
  label: string;
  value: string | null | undefined;
}

function TextBlock({ label, value }: TextBlockProps) {
  return (
    <div className="space-y-1">
      <p className="text-xs font-medium text-txt-tertiary">{label}</p>
      <p className="whitespace-pre-wrap text-sm text-txt-secondary">
        {value || "—"}
      </p>
    </div>
  );
}
