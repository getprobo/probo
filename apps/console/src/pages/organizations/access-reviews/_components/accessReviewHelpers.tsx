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

import { parseDate } from "@probo/helpers";
import {
  IconCircleCheck,
  IconCircleQuestionmark,
  IconCircleX,
} from "@probo/ui";

const MS_PER_SECOND = 1000;
const MS_PER_MINUTE = MS_PER_SECOND * 60;
const MS_PER_HOUR = MS_PER_MINUTE * 60;
const MS_PER_DAY = MS_PER_HOUR * 24;

// Compact elapsed age for dense list cells (e.g. "now", "5s", "20m", "4h", "1d", "1mo", "2y").
function shortAgeFormat(date: string, now: Date = new Date()): string {
  const elapsedMs = Math.max(0, now.getTime() - parseDate(date).getTime());
  const seconds = Math.floor(elapsedMs / MS_PER_SECOND);

  if (seconds < 60) {
    return seconds === 0 ? "now" : `${seconds}s`;
  }

  const minutes = Math.floor(elapsedMs / MS_PER_MINUTE);
  if (minutes < 60) {
    return `${minutes}m`;
  }

  const hours = Math.floor(elapsedMs / MS_PER_HOUR);
  if (hours < 24) {
    return `${hours}h`;
  }

  const days = Math.floor(elapsedMs / MS_PER_DAY);
  if (days < 30) {
    return `${days}d`;
  }
  if (days < 365) {
    return `${Math.floor(days / 30)}mo`;
  }
  return `${Math.floor(days / 365)}y`;
}

type BadgeVariant = "neutral" | "info" | "warning" | "success" | "danger";

export function statusBadgeVariant(status: string): BadgeVariant {
  switch (status) {
    case "DRAFT":
      return "neutral";
    case "IN_PROGRESS":
      return "info";
    case "PENDING_ACTIONS":
      return "warning";
    case "COMPLETED":
      return "success";
    case "CANCELLED":
      return "danger";
    default:
      return "neutral";
  }
}

export function isCampaignDeletableStatus(status: string): boolean {
  return status !== "IN_PROGRESS";
}

export function fetchStatusBadgeVariant(status: string): BadgeVariant {
  switch (status) {
    case "SUCCESS":
      return "success";
    case "FAILED":
      return "danger";
    case "FETCHING":
      return "info";
    case "QUEUED":
      return "neutral";
    default:
      return "info";
  }
}

export function statusLabel(
  t: (key: string) => string,
  status: string,
): string {
  switch (status) {
    case "DRAFT":
      return t("accessReviewCampaignsTab.status.draft");
    case "IN_PROGRESS":
      return t("accessReviewCampaignsTab.status.in_progress");
    case "PENDING_ACTIONS":
      return t("accessReviewCampaignsTab.status.pending_actions");
    case "COMPLETED":
      return t("accessReviewCampaignsTab.status.completed");
    case "CANCELLED":
      return t("accessReviewCampaignsTab.status.cancelled");
    default:
      return status;
  }
}

export function decisionBadgeVariant(decision: string): BadgeVariant {
  switch (decision) {
    case "APPROVED":
      return "success";
    case "REVOKE":
      return "danger";
    case "DEFER":
      return "warning";
    case "ESCALATE":
      return "info";
    default:
      return "neutral";
  }
}

export function decisionLabel(
  t: (key: string) => string,
  decision: string,
): string {
  switch (decision) {
    case "PENDING":
      return t("campaignDetailPage.decisions.pending");
    case "APPROVED":
      return t("campaignDetailPage.decisions.approved");
    case "REVOKE":
      return t("campaignDetailPage.decisions.revoke");
    case "DEFER":
      return t("campaignDetailPage.decisions.defer");
    case "ESCALATE":
      return t("campaignDetailPage.decisions.escalate");
    default:
      return decision;
  }
}

export function flagBadgeVariant(flag: string): BadgeVariant {
  switch (flag) {
    case "ORPHANED":
    case "TERMINATED_USER":
    case "CONTRACTOR_EXPIRED":
      return "danger";
    case "DORMANT":
    case "EXCESSIVE":
    case "SOD_CONFLICT":
    case "PRIVILEGED_ACCESS":
    case "ROLE_CREEP":
    case "ROLE_MISMATCH":
      return "warning";
    case "NO_BUSINESS_JUSTIFICATION":
    case "OUT_OF_DEPARTMENT":
    case "SHARED_ACCOUNT":
    case "INACTIVE":
    case "NEW":
      return "info";
    default:
      return "neutral";
  }
}

export const flagGroups = [
  {
    label: "Account",
    flags: [
      { value: "ORPHANED" as const, label: "Orphan account" },
      { value: "DORMANT" as const, label: "Dormant" },
      { value: "TERMINATED_USER" as const, label: "Terminated user" },
      { value: "CONTRACTOR_EXPIRED" as const, label: "Contractor expired" },
    ],
  },
  {
    label: "Privileges",
    flags: [
      { value: "EXCESSIVE" as const, label: "Excessive privileges" },
      { value: "SOD_CONFLICT" as const, label: "SoD conflict" },
      { value: "PRIVILEGED_ACCESS" as const, label: "Privileged access" },
      { value: "ROLE_CREEP" as const, label: "Role creep" },
    ],
  },
  {
    label: "Anomaly",
    flags: [
      { value: "NO_BUSINESS_JUSTIFICATION" as const, label: "No justification" },
      { value: "OUT_OF_DEPARTMENT" as const, label: "Out of department" },
      { value: "SHARED_ACCOUNT" as const, label: "Shared account" },
    ],
  },
];

export function flagLabel(flag: string): string {
  for (const group of flagGroups) {
    for (const f of group.flags) {
      if (f.value === flag) return f.label;
    }
  }
  if (flag === "NONE") return "None";
  // Legacy flag values not shown in the grouped dropdown
  if (flag === "INACTIVE") return "Inactive";
  if (flag === "ROLE_MISMATCH") return "Role mismatch";
  if (flag === "NEW") return "New";
  return flag;
}

export function formatStatus(status: string): string {
  return status.replace(/_/g, " ");
}

export function NotAvailable() {
  return (
    <span className="text-xs text-txt-tertiary">N/A</span>
  );
}

export function MfaStatusIcon({
  status,
  label,
}: {
  status: string;
  label: string;
}) {
  if (status === "ENABLED") {
    return (
      <span role="img" aria-label={label} title={label} className="inline-flex text-txt-success">
        <IconCircleCheck size={16} />
      </span>
    );
  }

  if (status === "DISABLED") {
    return (
      <span role="img" aria-label={label} title={label} className="inline-flex text-txt-danger">
        <IconCircleX size={16} />
      </span>
    );
  }

  return (
    <span role="img" aria-label={label} title={label} className="inline-flex text-txt-tertiary">
      <IconCircleQuestionmark size={16} />
    </span>
  );
}

export function AuthMethodStatus({
  method,
  label,
  unknownLabel,
}: {
  method: string;
  label: string;
  unknownLabel: string;
}) {
  if (method === "UNKNOWN") {
    return (
      <span
        role="img"
        aria-label={unknownLabel}
        title={unknownLabel}
        className="inline-flex text-txt-tertiary"
      >
        <IconCircleQuestionmark size={16} />
      </span>
    );
  }

  return (
    <span
      aria-label={label}
      title={label}
      className="text-xs text-txt-primary"
    >
      {label}
    </span>
  );
}

export function BooleanStatusIcon({
  value,
  trueLabel,
  falseLabel,
  unknownLabel,
}: {
  value: boolean | null | undefined;
  trueLabel: string;
  falseLabel: string;
  unknownLabel: string;
}) {
  if (value == null) {
    return (
      <span role="img" aria-label={unknownLabel} title={unknownLabel} className="inline-flex text-txt-tertiary">
        <IconCircleQuestionmark size={16} />
      </span>
    );
  }

  if (value) {
    return (
      <span role="img" aria-label={trueLabel} title={trueLabel} className="inline-flex text-txt-success">
        <IconCircleCheck size={16} />
      </span>
    );
  }

  return (
    <span role="img" aria-label={falseLabel} title={falseLabel} className="inline-flex text-txt-danger">
      <IconCircleX size={16} />
    </span>
  );
}

export function LastLoginStatus({
  lastLogin,
  formatted,
  unknownLabel,
  compact = false,
}: {
  lastLogin: string | null | undefined;
  formatted: string;
  unknownLabel: string;
  compact?: boolean;
}) {
  if (lastLogin) {
    if (compact) {
      return (
        <span
          aria-label={formatted}
          title={formatted}
          className="text-xs tabular-nums text-txt-primary"
        >
          {shortAgeFormat(lastLogin)}
        </span>
      );
    }

    return (
      <span role="img" aria-label={formatted} className="inline-flex min-w-0 items-center gap-1.5" title={formatted}>
        <IconCircleCheck size={16} className="shrink-0 text-txt-tertiary" />
        <span className="truncate" aria-hidden="true">{formatted}</span>
      </span>
    );
  }

  return (
    <span role="img" aria-label={unknownLabel} title={unknownLabel} className="inline-flex text-txt-tertiary">
      <IconCircleQuestionmark size={16} />
    </span>
  );
}
